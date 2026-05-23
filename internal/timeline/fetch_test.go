//nolint:testpackage // white-box test populates unexported timelineQuery fields directly.
package timeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

// fakeQuerier returns the pre-built timelineQuery pages in order, simulating
// shurcooL/githubv4's struct-population semantics without an HTTP round-trip.
type fakeQuerier struct {
	t        *testing.T
	pages    []timelineQuery
	idx      int
	queryErr error
	calls    int
}

func (f *fakeQuerier) Query(_ context.Context, q any, _ map[string]any) error {
	f.calls++
	if f.queryErr != nil {
		return f.queryErr
	}
	dst, ok := q.(*timelineQuery)
	if !ok {
		f.t.Fatalf("Query received %T, want *timelineQuery", q)
	}
	if f.idx >= len(f.pages) {
		f.t.Fatalf("unexpected extra Query call (have %d pages)", len(f.pages))
	}
	*dst = f.pages[f.idx]
	f.idx++
	return nil
}

func newPRPage(t *testing.T, nodes []prTimelineNode, endCursor string, hasNext bool) timelineQuery {
	t.Helper()
	var q timelineQuery
	q.Repository.IssueOrPullRequest.Typename = "PullRequest"
	q.Repository.IssueOrPullRequest.PullRequest.TimelineItems.Nodes = nodes
	q.Repository.IssueOrPullRequest.PullRequest.TimelineItems.PageInfo.EndCursor = githubv4.String(endCursor)
	q.Repository.IssueOrPullRequest.PullRequest.TimelineItems.PageInfo.HasNextPage = githubv4.Boolean(hasNext)
	return q
}

func newIssuePage(t *testing.T, nodes []issueTimelineNode, endCursor string, hasNext bool) timelineQuery {
	t.Helper()
	var q timelineQuery
	q.Repository.IssueOrPullRequest.Typename = "Issue"
	q.Repository.IssueOrPullRequest.Issue.TimelineItems.Nodes = nodes
	q.Repository.IssueOrPullRequest.Issue.TimelineItems.PageInfo.EndCursor = githubv4.String(endCursor)
	q.Repository.IssueOrPullRequest.Issue.TimelineItems.PageInfo.HasNextPage = githubv4.Boolean(hasNext)
	return q
}

func TestFetch_paginatesAndSortsChronologically(t *testing.T) {
	t.Parallel()
	earlier := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	later := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

	labeled := prTimelineNode{Typename: "LabeledEvent"}
	labeled.LabeledEvent.Actor.Login = "alice"
	labeled.LabeledEvent.CreatedAt = dt(later)
	labeled.LabeledEvent.Label.Name = "bug"

	review := prTimelineNode{Typename: "PullRequestReview"}
	review.PullRequestReview.Author.Login = "bob"
	review.PullRequestReview.SubmittedAt = dt(earlier)
	review.PullRequestReview.State = githubv4.PullRequestReviewStateApproved

	fake := &fakeQuerier{
		t: t,
		pages: []timelineQuery{
			newPRPage(t, []prTimelineNode{labeled}, "cur1", true),
			newPRPage(t, []prTimelineNode{review}, "", false),
		},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 123)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Type != "PullRequestReview" || events[1].Type != "LabeledEvent" {
		t.Errorf("events not sorted chronologically: %+v", events)
	}
	if fake.calls != 2 {
		t.Errorf("Query was called %d times, want 2 (one per page)", fake.calls)
	}
}

func TestFetch_handlesIssueTimeline(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	comment := issueTimelineNode{Typename: "IssueComment"}
	comment.IssueComment.Author.Login = "alice"
	comment.IssueComment.CreatedAt = dt(ts)
	comment.IssueComment.Body = "first reaction"

	fake := &fakeQuerier{
		t:     t,
		pages: []timelineQuery{newIssuePage(t, []issueTimelineNode{comment}, "", false)},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 7)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 1 || events[0].Type != "IssueComment" || events[0].Summary != "first reaction" {
		t.Errorf("unexpected events: %+v", events)
	}
}

func TestFetch_notFoundWhenIssueOrPullRequestIsNull(t *testing.T) {
	t.Parallel()
	// Typename left empty mimics GraphQL returning `issueOrPullRequest: null`.
	fake := &fakeQuerier{t: t, pages: []timelineQuery{{}}}

	_, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 99999)
	if err == nil {
		t.Fatal("expected error for missing issue/PR")
	}
	if !errorContains(err, "not found") {
		t.Errorf("error %q should contain 'not found'", err)
	}
}

func TestFetch_propagatesQueryErrors(t *testing.T) {
	t.Parallel()
	fake := &fakeQuerier{t: t, queryErr: errors.New("boom")}

	_, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 1)
	if err == nil || !errorContains(err, "boom") {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestFetch_rejectsInvalidInput(t *testing.T) {
	t.Parallel()
	fake := &fakeQuerier{t: t}
	if _, err := Fetch(context.Background(), fake, Repo{}, 1); err == nil {
		t.Error("expected error for empty repo")
	}
	if _, err := Fetch(context.Background(), fake, Repo{Owner: "o", Name: "r"}, 0); err == nil {
		t.Error("expected error for zero issue/PR number")
	}
}

func errorContains(err error, substr string) bool {
	return err != nil && stringContains(err.Error(), substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

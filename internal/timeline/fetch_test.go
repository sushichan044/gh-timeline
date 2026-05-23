//nolint:testpackage // white-box test populates unexported timelineQuery fields directly.
package timeline

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

// fakeQuerier serves pre-built timelineQuery pages keyed by the GraphQL
// `skip` variable, simulating shurcooL/githubv4's struct-population semantics
// without an HTTP round-trip. Lookups are safe to call from multiple
// goroutines because Fetch dispatches non-first pages in parallel.
type fakeQuerier struct {
	t        *testing.T
	pages    map[int]timelineQuery
	queryErr error

	mu    sync.Mutex
	calls int
}

func (f *fakeQuerier) Query(_ context.Context, q any, vars map[string]any) error {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()

	if f.queryErr != nil {
		return f.queryErr
	}
	dst, ok := q.(*timelineQuery)
	if !ok {
		f.t.Fatalf("Query received %T, want *timelineQuery", q)
	}
	skip, ok := vars["skip"].(githubv4.Int)
	if !ok {
		f.t.Fatalf("Query missing skip variable, got %T", vars["skip"])
	}
	page, ok := f.pages[int(skip)]
	if !ok {
		f.t.Errorf("unexpected Query for skip=%d", int(skip))
		return nil
	}
	*dst = page
	return nil
}

func (f *fakeQuerier) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func newPRPage(t *testing.T, nodes []prTimelineNode, totalCount int) timelineQuery {
	t.Helper()
	var q timelineQuery
	q.Repository.IssueOrPullRequest.Typename = "PullRequest"
	q.Repository.IssueOrPullRequest.PullRequest.TimelineItems.Nodes = nodes
	q.Repository.IssueOrPullRequest.PullRequest.TimelineItems.TotalCount = githubv4.Int(int32(totalCount))
	return q
}

func newIssuePage(t *testing.T, nodes []issueTimelineNode, totalCount int) timelineQuery {
	t.Helper()
	var q timelineQuery
	q.Repository.IssueOrPullRequest.Typename = "Issue"
	q.Repository.IssueOrPullRequest.Issue.TimelineItems.Nodes = nodes
	q.Repository.IssueOrPullRequest.Issue.TimelineItems.TotalCount = githubv4.Int(int32(totalCount))
	return q
}

func TestFetch_parallelizesAndSortsChronologically(t *testing.T) {
	t.Parallel()

	// Three pages worth of events, totalCount = 250. Each page returns one
	// representative event; their timestamps are deliberately out of order so
	// the final sort is observable.
	tsPage0 := time.Date(2026, 1, 3, 9, 0, 0, 0, time.UTC) // middle in chrono order
	tsPage1 := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC) // earliest
	tsPage2 := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC) // latest

	page0Node := prTimelineNode{Typename: "LabeledEvent"}
	page0Node.LabeledEvent.Actor.Login = "alice"
	page0Node.LabeledEvent.CreatedAt = dt(tsPage0)
	page0Node.LabeledEvent.Label.Name = "bug"

	page1Node := prTimelineNode{Typename: "PullRequestReview"}
	page1Node.PullRequestReview.Author.Login = "bob"
	page1Node.PullRequestReview.SubmittedAt = dt(tsPage1)
	page1Node.PullRequestReview.State = githubv4.PullRequestReviewStateApproved

	page2Node := prTimelineNode{Typename: "MergedEvent"}
	page2Node.MergedEvent.Actor.Login = "carol"
	page2Node.MergedEvent.CreatedAt = dt(tsPage2)

	fake := &fakeQuerier{
		t: t,
		pages: map[int]timelineQuery{
			0:   newPRPage(t, []prTimelineNode{page0Node}, 250),
			100: newPRPage(t, []prTimelineNode{page1Node}, 250),
			200: newPRPage(t, []prTimelineNode{page2Node}, 250),
		},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 123)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[0].Type != "PullRequestReview" || events[1].Type != "LabeledEvent" || events[2].Type != "MergedEvent" {
		t.Errorf("events not sorted chronologically: %+v", events)
	}
	if calls := fake.callCount(); calls != 3 {
		t.Errorf("Query was called %d times, want 3 (one per page)", calls)
	}
}

func TestFetch_singlePageSkipsParallelDispatch(t *testing.T) {
	t.Parallel()

	node := prTimelineNode{Typename: "LabeledEvent"}
	node.LabeledEvent.Actor.Login = "alice"
	node.LabeledEvent.CreatedAt = dt(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))
	node.LabeledEvent.Label.Name = "bug"

	fake := &fakeQuerier{
		t: t,
		pages: map[int]timelineQuery{
			0: newPRPage(t, []prTimelineNode{node}, 50),
		},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 123)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 1 || events[0].Type != "LabeledEvent" {
		t.Errorf("unexpected events: %+v", events)
	}
	if calls := fake.callCount(); calls != 1 {
		t.Errorf("Query was called %d times, want 1 (single page below page size)", calls)
	}
}

func TestFetch_emptyTimelineMakesOneCall(t *testing.T) {
	t.Parallel()

	fake := &fakeQuerier{
		t: t,
		pages: map[int]timelineQuery{
			0: newIssuePage(t, nil, 0),
		},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 7)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events, want 0", len(events))
	}
	if calls := fake.callCount(); calls != 1 {
		t.Errorf("Query was called %d times, want 1", calls)
	}
}

func TestFetch_exactlyPageSizeMakesOneCall(t *testing.T) {
	t.Parallel()

	node := issueTimelineNode{Typename: "IssueComment"}
	node.IssueComment.Author.Login = "alice"
	node.IssueComment.CreatedAt = dt(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))
	node.IssueComment.Body = "hello"

	fake := &fakeQuerier{
		t: t,
		pages: map[int]timelineQuery{
			0: newIssuePage(t, []issueTimelineNode{node}, 100),
		},
	}

	events, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 7)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if calls := fake.callCount(); calls != 1 {
		t.Errorf("Query was called %d times, want 1 (totalCount equals page size)", calls)
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
		t: t,
		pages: map[int]timelineQuery{
			0: newIssuePage(t, []issueTimelineNode{comment}, 1),
		},
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
	fake := &fakeQuerier{t: t, pages: map[int]timelineQuery{0: {}}}

	_, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 99999)
	if err == nil {
		t.Fatal("expected error for missing issue/PR")
	}
	if !errorContains(err, "not found") {
		t.Errorf("error %q should contain 'not found'", err)
	}
}

func TestFetch_propagatesFirstPageError(t *testing.T) {
	t.Parallel()
	fake := &fakeQuerier{t: t, queryErr: errors.New("boom")}

	_, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 1)
	if err == nil || !errorContains(err, "boom") {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestFetch_propagatesParallelPageError(t *testing.T) {
	t.Parallel()

	node := prTimelineNode{Typename: "LabeledEvent"}
	node.LabeledEvent.Actor.Login = "alice"
	node.LabeledEvent.CreatedAt = dt(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC))
	node.LabeledEvent.Label.Name = "bug"

	// First page succeeds (totalCount=200 so one extra parallel page is dispatched);
	// the parallel page is unregistered so the fake reports an error via t.Errorf.
	// We swap to a custom fake that returns an error for skip!=0 to assert the
	// surface bubbles up.
	fake := &erroringPagedQuerier{
		t:     t,
		first: newPRPage(t, []prTimelineNode{node}, 200),
		err:   errors.New("page-2-broke"),
	}

	_, err := Fetch(context.Background(), fake, Repo{Owner: "octo", Name: "demo"}, 1)
	if err == nil || !errorContains(err, "page-2-broke") {
		t.Fatalf("expected wrapped page-2-broke error, got %v", err)
	}
}

// erroringPagedQuerier returns the first page successfully and an error for
// every subsequent skip, so tests can verify parallel-page errors bubble up.
type erroringPagedQuerier struct {
	t     *testing.T
	first timelineQuery
	err   error
}

func (f *erroringPagedQuerier) Query(_ context.Context, q any, vars map[string]any) error {
	skip, ok := vars["skip"].(githubv4.Int)
	if !ok {
		f.t.Fatalf("Query missing skip variable, got %T", vars["skip"])
	}
	if int(skip) == 0 {
		dst, dstOK := q.(*timelineQuery)
		if !dstOK {
			f.t.Fatalf("Query received %T, want *timelineQuery", q)
		}
		*dst = f.first
		return nil
	}
	return f.err
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

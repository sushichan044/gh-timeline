//nolint:testpackage // white-box test exercises unexported dispatchPRNode / fragment structs directly.
package timeline

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

func dt(t time.Time) githubv4.DateTime { return githubv4.DateTime{Time: t} }

func uri(t *testing.T, raw string) githubv4.URI {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url %q: %v", raw, err)
	}
	return githubv4.URI{URL: u}
}

// TestDispatchPRNode_richSummariesPerType drives dispatchPRNode with a
// representative slice of the PullRequestTimelineItems union. The goal is to
// pin the rendered Type / Actor / Summary shape for each event variant so
// future schema additions don't silently regress them.
func TestDispatchPRNode_richSummariesPerType(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		node        prTimelineNode
		wantType    string
		wantActor   string
		wantSummary string
	}{
		{
			name: "PullRequestCommit uses commit headline and committed date",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "PullRequestCommit"}
				n.PullRequestCommit.ID = "PC_node"
				n.PullRequestCommit.Commit.MessageHeadline = "feat: add a thing"
				n.PullRequestCommit.Commit.CommittedDate = dt(ts)
				n.PullRequestCommit.Commit.OID = "abc123"
				n.PullRequestCommit.Commit.Author.User.Login = "alice"
				return n
			}(),
			wantType:    "PullRequestCommit",
			wantActor:   "alice",
			wantSummary: "feat: add a thing",
		},
		{
			name: "PullRequestReview pairs state with body first line",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "PullRequestReview"}
				n.PullRequestReview.Author.Login = "bob"
				n.PullRequestReview.SubmittedAt = dt(ts)
				n.PullRequestReview.State = githubv4.PullRequestReviewStateApproved
				n.PullRequestReview.Body = "looks good!\nmore details"
				n.PullRequestReview.DatabaseID = int64(42)
				return n
			}(),
			wantType:    "PullRequestReview",
			wantActor:   "bob",
			wantSummary: "APPROVED: looks good!",
		},
		{
			name: "IssueComment shows first line of body, truncated",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueComment"}
				n.IssueComment.Author.Login = "carol"
				n.IssueComment.CreatedAt = dt(ts)
				n.IssueComment.Body = "hello there\nignored second line"
				return n
			}(),
			wantType:    "IssueComment",
			wantActor:   "carol",
			wantSummary: "hello there",
		},
		{
			name: "LabeledEvent uses the added verb with label name",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "LabeledEvent"}
				n.LabeledEvent.Actor.Login = "dave"
				n.LabeledEvent.CreatedAt = dt(ts)
				n.LabeledEvent.Label.Name = "bug"
				return n
			}(),
			wantType:    "LabeledEvent",
			wantActor:   "dave",
			wantSummary: "added label bug",
		},
		{
			name: "AssignedEvent says who got assigned",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "AssignedEvent"}
				n.AssignedEvent.Actor.Login = "ada"
				n.AssignedEvent.CreatedAt = dt(ts)
				n.AssignedEvent.Assignee.User.Login = "bea"
				return n
			}(),
			wantType:    "AssignedEvent",
			wantActor:   "ada",
			wantSummary: "assigned bea",
		},
		{
			name: "UnlabeledEvent uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "UnlabeledEvent"}
				n.UnlabeledEvent.Actor.Login = "dave"
				n.UnlabeledEvent.CreatedAt = dt(ts)
				n.UnlabeledEvent.Label.Name = "wontfix"
				return n
			}(),
			wantType:    "UnlabeledEvent",
			wantActor:   "dave",
			wantSummary: "removed label wontfix",
		},
		{
			name: "MergedEvent shows commit SHA and base ref",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "MergedEvent"}
				n.MergedEvent.Actor.Login = "eve"
				n.MergedEvent.CreatedAt = dt(ts)
				n.MergedEvent.Commit.OID = "deadbeef0000"
				n.MergedEvent.MergeRefName = "main"
				return n
			}(),
			wantType:    "MergedEvent",
			wantActor:   "eve",
			wantSummary: "merged deadbee into main",
		},
		{
			name: "ReviewRequestedEvent says who was asked",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ReviewRequestedEvent"}
				n.ReviewRequestedEvent.Actor.Login = "frank"
				n.ReviewRequestedEvent.CreatedAt = dt(ts)
				n.ReviewRequestedEvent.RequestedReviewer.User.Login = "grace"
				return n
			}(),
			wantType:    "ReviewRequestedEvent",
			wantActor:   "frank",
			wantSummary: "requested review from grace",
		},
		{
			name: "ReviewRequestedEvent falls back to team slug when user is empty",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ReviewRequestedEvent"}
				n.ReviewRequestedEvent.Actor.Login = "frank"
				n.ReviewRequestedEvent.CreatedAt = dt(ts)
				n.ReviewRequestedEvent.RequestedReviewer.Team.Slug = "core"
				return n
			}(),
			wantType:    "ReviewRequestedEvent",
			wantActor:   "frank",
			wantSummary: "requested review from team:core",
		},
		{
			name: "ReviewRequestRemovedEvent uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ReviewRequestRemovedEvent"}
				n.ReviewRequestRemovedEvent.Actor.Login = "frank"
				n.ReviewRequestRemovedEvent.CreatedAt = dt(ts)
				n.ReviewRequestRemovedEvent.RequestedReviewer.User.Login = "grace"
				return n
			}(),
			wantType:    "ReviewRequestRemovedEvent",
			wantActor:   "frank",
			wantSummary: "removed review request from grace",
		},
		{
			name: "ReadyForReviewEvent uses fixed wording",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ReadyForReviewEvent"}
				n.ReadyForReviewEvent.Actor.Login = "heidi"
				n.ReadyForReviewEvent.CreatedAt = dt(ts)
				return n
			}(),
			wantType:    "ReadyForReviewEvent",
			wantActor:   "heidi",
			wantSummary: "marked ready for review",
		},
		{
			name: "ConvertToDraftEvent uses fixed wording",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ConvertToDraftEvent"}
				n.ConvertToDraftEvent.Actor.Login = "ivan"
				n.ConvertToDraftEvent.CreatedAt = dt(ts)
				return n
			}(),
			wantType:    "ConvertToDraftEvent",
			wantActor:   "ivan",
			wantSummary: "converted to draft",
		},
		{
			name: "HeadRefForcePushedEvent shows before → after commit pair",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "HeadRefForcePushedEvent"}
				n.HeadRefForcePushedEvent.Actor.Login = "judy"
				n.HeadRefForcePushedEvent.CreatedAt = dt(ts)
				n.HeadRefForcePushedEvent.BeforeCommit.OID = "abcdef1234567890"
				n.HeadRefForcePushedEvent.AfterCommit.OID = "1234567890abcdef"
				return n
			}(),
			wantType:    "HeadRefForcePushedEvent",
			wantActor:   "judy",
			wantSummary: "force-pushed head: abcdef1 → 1234567",
		},
		{
			name: "HeadRefForcePushedEvent falls back to after-only when before is empty",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "HeadRefForcePushedEvent"}
				n.HeadRefForcePushedEvent.Actor.Login = "judy"
				n.HeadRefForcePushedEvent.CreatedAt = dt(ts)
				n.HeadRefForcePushedEvent.AfterCommit.OID = "1234567890abcdef"
				return n
			}(),
			wantType:    "HeadRefForcePushedEvent",
			wantActor:   "judy",
			wantSummary: "force-pushed head to 1234567",
		},
		{
			name: "ClosedEvent includes stateReason when present",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ClosedEvent"}
				n.ClosedEvent.Actor.Login = "kate"
				n.ClosedEvent.CreatedAt = dt(ts)
				n.ClosedEvent.StateReason = "NOT_PLANNED"
				return n
			}(),
			wantType:    "ClosedEvent",
			wantActor:   "kate",
			wantSummary: "closed: NOT_PLANNED",
		},
		{
			name: "RenamedTitleEvent shows both titles",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "RenamedTitleEvent"}
				n.RenamedTitleEvent.Actor.Login = "leo"
				n.RenamedTitleEvent.CreatedAt = dt(ts)
				n.RenamedTitleEvent.PreviousTitle = "WIP"
				n.RenamedTitleEvent.CurrentTitle = "Fix the thing"
				return n
			}(),
			wantType:    "RenamedTitleEvent",
			wantActor:   "leo",
			wantSummary: `renamed: "WIP" → "Fix the thing"`,
		},
		{
			name: "CrossReferencedEvent points at the source issue or PR",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "CrossReferencedEvent"}
				n.CrossReferencedEvent.Actor.Login = "mia"
				n.CrossReferencedEvent.CreatedAt = dt(ts)
				n.CrossReferencedEvent.Source.PullRequest.Number = 42
				n.CrossReferencedEvent.Source.PullRequest.Title = "Other PR"
				n.CrossReferencedEvent.Source.PullRequest.Repository.NameWithOwner = "octo/other"
				return n
			}(),
			wantType:    "CrossReferencedEvent",
			wantActor:   "mia",
			wantSummary: "referenced from octo/other#42: Other PR",
		},
		{
			name: "ReviewDismissedEvent quotes the dismissing reason",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ReviewDismissedEvent"}
				n.ReviewDismissedEvent.Actor.Login = "noah"
				n.ReviewDismissedEvent.CreatedAt = dt(ts)
				n.ReviewDismissedEvent.Review.Author.Login = "oscar"
				n.ReviewDismissedEvent.DismissalMessage = "stale review"
				return n
			}(),
			wantType:    "ReviewDismissedEvent",
			wantActor:   "noah",
			wantSummary: "dismissed review by oscar: stale review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dispatchPRNode(tt.node)
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Actor != tt.wantActor {
				t.Errorf("Actor = %q, want %q", got.Actor, tt.wantActor)
			}
			if got.Summary != tt.wantSummary {
				t.Errorf("Summary = %q, want %q", got.Summary, tt.wantSummary)
			}
			if !got.Timestamp.Equal(ts) {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, ts)
			}
		})
	}
}

// TestDispatchPRNode_unknownTypeFallback asserts that an event the dispatcher
// has no handler for (e.g. a future schema addition) still surfaces with its
// raw __typename and an empty summary, so it shows up as `[Type] @-` in text
// rendering rather than vanishing silently.
func TestDispatchPRNode_unknownTypeFallback(t *testing.T) {
	t.Parallel()
	got := dispatchPRNode(prTimelineNode{Typename: "BrandNewEventType"})
	if got.Type != "BrandNewEventType" {
		t.Errorf("Type = %q, want BrandNewEventType", got.Type)
	}
	if got.Actor != "" {
		t.Errorf("Actor = %q, want empty for unknown event", got.Actor)
	}
	if got.Summary != "" {
		t.Errorf("Summary = %q, want empty for unknown event", got.Summary)
	}
}

// TestDispatchIssueNode_handlesSharedEvents covers the Issue dispatcher path
// for events shared with the PR timeline. PR-only events (e.g. MergedEvent)
// are intentionally excluded since they aren't members of
// IssueTimelineItems.
func TestDispatchIssueNode_handlesSharedEvents(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)

	n := issueTimelineNode{Typename: "LabeledEvent"}
	n.LabeledEvent.Actor.Login = "alice"
	n.LabeledEvent.CreatedAt = dt(ts)
	n.LabeledEvent.Label.Name = "needs-triage"

	got := dispatchIssueNode(n)
	if got.Type != "LabeledEvent" || got.Actor != "alice" || got.Summary != "added label needs-triage" {
		t.Errorf("unexpected event: %+v", got)
	}
}

// TestHandleIssueComment_truncatesAndCarriesIDs ensures the comment handler
// populates CommentID + URL in Ref so downstream `gh api` calls can drill
// into the full body when the 72-rune summary gets cut.
func TestHandleIssueComment_truncatesAndCarriesIDs(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", 100)
	var f issueCommentFragment
	f.ID = "IC_node"
	f.DatabaseID = int64(99)
	f.Author.Login = "carol"
	f.CreatedAt = dt(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC))
	f.Body = githubv4.String(long)
	f.URL = uri(t, "https://api.github.com/repos/o/r/issues/comments/99")

	e := handleIssueComment("IssueComment", f)
	if e.Ref.CommentID != 99 {
		t.Errorf("CommentID = %d, want 99", e.Ref.CommentID)
	}
	if e.Ref.URL == "" {
		t.Error("expected non-empty Ref.URL")
	}
	if !strings.HasSuffix(e.Summary, "…") {
		t.Errorf("expected truncation ellipsis, summary = %q", e.Summary)
	}
	if rc := len([]rune(e.Summary)); rc != 72 {
		t.Errorf("summary rune count = %d, want 72", rc)
	}
}

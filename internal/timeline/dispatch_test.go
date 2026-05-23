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
		{
			name: "SubIssueAddedEvent names the linked sub-issue",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "SubIssueAddedEvent"}
				n.SubIssueAddedEvent.Actor.Login = "pat"
				n.SubIssueAddedEvent.CreatedAt = dt(ts)
				n.SubIssueAddedEvent.SubIssue.Number = 7
				n.SubIssueAddedEvent.SubIssue.Title = "Implement child"
				n.SubIssueAddedEvent.SubIssue.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantType:    "SubIssueAddedEvent",
			wantActor:   "pat",
			wantSummary: "added sub-issue octo/repo#7: Implement child",
		},
		{
			name: "SubIssueRemovedEvent uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "SubIssueRemovedEvent"}
				n.SubIssueRemovedEvent.Actor.Login = "pat"
				n.SubIssueRemovedEvent.CreatedAt = dt(ts)
				n.SubIssueRemovedEvent.SubIssue.Number = 7
				n.SubIssueRemovedEvent.SubIssue.Title = "Implement child"
				n.SubIssueRemovedEvent.SubIssue.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantType:    "SubIssueRemovedEvent",
			wantActor:   "pat",
			wantSummary: "removed sub-issue octo/repo#7: Implement child",
		},
		{
			name: "SubIssueAddedEvent falls back to bare verb when issue ref is empty",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "SubIssueAddedEvent"}
				n.SubIssueAddedEvent.Actor.Login = "pat"
				n.SubIssueAddedEvent.CreatedAt = dt(ts)
				return n
			}(),
			wantType:    "SubIssueAddedEvent",
			wantActor:   "pat",
			wantSummary: "added sub-issue",
		},
		{
			name: "ParentIssueAddedEvent names the linked parent issue",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ParentIssueAddedEvent"}
				n.ParentIssueAddedEvent.Actor.Login = "quinn"
				n.ParentIssueAddedEvent.CreatedAt = dt(ts)
				n.ParentIssueAddedEvent.Parent.Number = 3
				n.ParentIssueAddedEvent.Parent.Title = "Epic"
				n.ParentIssueAddedEvent.Parent.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantType:    "ParentIssueAddedEvent",
			wantActor:   "quinn",
			wantSummary: "added parent issue octo/repo#3: Epic",
		},
		{
			name: "ParentIssueRemovedEvent uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ParentIssueRemovedEvent"}
				n.ParentIssueRemovedEvent.Actor.Login = "quinn"
				n.ParentIssueRemovedEvent.CreatedAt = dt(ts)
				n.ParentIssueRemovedEvent.Parent.Number = 3
				n.ParentIssueRemovedEvent.Parent.Title = "Epic"
				n.ParentIssueRemovedEvent.Parent.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantType:    "ParentIssueRemovedEvent",
			wantActor:   "quinn",
			wantSummary: "removed parent issue octo/repo#3: Epic",
		},
		{
			name: "BlockedByAddedEvent names the blocking issue",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "BlockedByAddedEvent"}
				n.BlockedByAddedEvent.Actor.Login = "rita"
				n.BlockedByAddedEvent.CreatedAt = dt(ts)
				n.BlockedByAddedEvent.BlockingIssue.Number = 99
				n.BlockedByAddedEvent.BlockingIssue.Title = "Upstream blocker"
				n.BlockedByAddedEvent.BlockingIssue.Repository.NameWithOwner = "octo/dep"
				return n
			}(),
			wantType:    "BlockedByAddedEvent",
			wantActor:   "rita",
			wantSummary: "blocked by octo/dep#99: Upstream blocker",
		},
		{
			name: "BlockedByRemovedEvent uses the no-longer phrasing",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "BlockedByRemovedEvent"}
				n.BlockedByRemovedEvent.Actor.Login = "rita"
				n.BlockedByRemovedEvent.CreatedAt = dt(ts)
				n.BlockedByRemovedEvent.BlockingIssue.Number = 99
				n.BlockedByRemovedEvent.BlockingIssue.Title = "Upstream blocker"
				n.BlockedByRemovedEvent.BlockingIssue.Repository.NameWithOwner = "octo/dep"
				return n
			}(),
			wantType:    "BlockedByRemovedEvent",
			wantActor:   "rita",
			wantSummary: "no longer blocked by octo/dep#99: Upstream blocker",
		},
		{
			name: "BlockingAddedEvent names the blocked downstream issue",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "BlockingAddedEvent"}
				n.BlockingAddedEvent.Actor.Login = "sue"
				n.BlockingAddedEvent.CreatedAt = dt(ts)
				n.BlockingAddedEvent.BlockedIssue.Number = 12
				n.BlockingAddedEvent.BlockedIssue.Title = "Downstream consumer"
				n.BlockingAddedEvent.BlockedIssue.Repository.NameWithOwner = "octo/consumer"
				return n
			}(),
			wantType:    "BlockingAddedEvent",
			wantActor:   "sue",
			wantSummary: "blocking octo/consumer#12: Downstream consumer",
		},
		{
			name: "BlockingRemovedEvent uses the no-longer phrasing",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "BlockingRemovedEvent"}
				n.BlockingRemovedEvent.Actor.Login = "sue"
				n.BlockingRemovedEvent.CreatedAt = dt(ts)
				n.BlockingRemovedEvent.BlockedIssue.Number = 12
				n.BlockingRemovedEvent.BlockedIssue.Title = "Downstream consumer"
				n.BlockingRemovedEvent.BlockedIssue.Repository.NameWithOwner = "octo/consumer"
				return n
			}(),
			wantType:    "BlockingRemovedEvent",
			wantActor:   "sue",
			wantSummary: "no longer blocking octo/consumer#12: Downstream consumer",
		},
		{
			name: "AddedToProjectV2Event quotes the project title",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "AddedToProjectV2Event"}
				n.AddedToProjectV2Event.Actor.Login = "tara"
				n.AddedToProjectV2Event.CreatedAt = dt(ts)
				n.AddedToProjectV2Event.Project.Title = "Roadmap"
				return n
			}(),
			wantType:    "AddedToProjectV2Event",
			wantActor:   "tara",
			wantSummary: `added to project "Roadmap"`,
		},
		{
			name: "RemovedFromProjectV2Event uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "RemovedFromProjectV2Event"}
				n.RemovedFromProjectV2Event.Actor.Login = "tara"
				n.RemovedFromProjectV2Event.CreatedAt = dt(ts)
				n.RemovedFromProjectV2Event.Project.Title = "Roadmap"
				return n
			}(),
			wantType:    "RemovedFromProjectV2Event",
			wantActor:   "tara",
			wantSummary: `removed from project "Roadmap"`,
		},
		{
			name: "ProjectV2ItemStatusChangedEvent shows project, previous and new status",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ProjectV2ItemStatusChangedEvent"}
				n.ProjectV2ItemStatusChangedEvent.Actor.Login = "uma"
				n.ProjectV2ItemStatusChangedEvent.CreatedAt = dt(ts)
				n.ProjectV2ItemStatusChangedEvent.Project.Title = "Roadmap"
				n.ProjectV2ItemStatusChangedEvent.PreviousStatus = "Todo"
				n.ProjectV2ItemStatusChangedEvent.Status = "In Progress"
				return n
			}(),
			wantType:    "ProjectV2ItemStatusChangedEvent",
			wantActor:   "uma",
			wantSummary: `status changed in project "Roadmap": "Todo" → "In Progress"`,
		},
		{
			name: "ConvertedFromDraftEvent reports the draft conversion",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "ConvertedFromDraftEvent"}
				n.ConvertedFromDraftEvent.Actor.Login = "vic"
				n.ConvertedFromDraftEvent.CreatedAt = dt(ts)
				n.ConvertedFromDraftEvent.Project.Title = "Roadmap"
				return n
			}(),
			wantType:    "ConvertedFromDraftEvent",
			wantActor:   "vic",
			wantSummary: `converted from draft "Roadmap"`,
		},
		{
			name: "IssueFieldAddedEvent quotes field name with its value",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueFieldAddedEvent"}
				n.IssueFieldAddedEvent.Actor.Login = "wes"
				n.IssueFieldAddedEvent.CreatedAt = dt(ts)
				n.IssueFieldAddedEvent.IssueField.IssueFieldCommon.Name = "Priority"
				n.IssueFieldAddedEvent.Value = "P1"
				return n
			}(),
			wantType:    "IssueFieldAddedEvent",
			wantActor:   "wes",
			wantSummary: `added field "Priority": P1`,
		},
		{
			name: "IssueFieldChangedEvent shows previous and new value",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueFieldChangedEvent"}
				n.IssueFieldChangedEvent.Actor.Login = "wes"
				n.IssueFieldChangedEvent.CreatedAt = dt(ts)
				n.IssueFieldChangedEvent.IssueField.IssueFieldCommon.Name = "Priority"
				n.IssueFieldChangedEvent.PreviousValue = "P1"
				n.IssueFieldChangedEvent.NewValue = "P0"
				return n
			}(),
			wantType:    "IssueFieldChangedEvent",
			wantActor:   "wes",
			wantSummary: `changed field "Priority": "P1" → "P0"`,
		},
		{
			name: "IssueFieldRemovedEvent quotes the removed field name",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueFieldRemovedEvent"}
				n.IssueFieldRemovedEvent.Actor.Login = "wes"
				n.IssueFieldRemovedEvent.CreatedAt = dt(ts)
				n.IssueFieldRemovedEvent.IssueField.IssueFieldCommon.Name = "Priority"
				return n
			}(),
			wantType:    "IssueFieldRemovedEvent",
			wantActor:   "wes",
			wantSummary: `removed field "Priority"`,
		},
		{
			name: "IssueTypeAddedEvent quotes the issue type name",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueTypeAddedEvent"}
				n.IssueTypeAddedEvent.Actor.Login = "xena"
				n.IssueTypeAddedEvent.CreatedAt = dt(ts)
				n.IssueTypeAddedEvent.IssueType.Name = "Bug"
				return n
			}(),
			wantType:    "IssueTypeAddedEvent",
			wantActor:   "xena",
			wantSummary: `set issue type to "Bug"`,
		},
		{
			name: "IssueTypeChangedEvent shows previous and new issue type",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueTypeChangedEvent"}
				n.IssueTypeChangedEvent.Actor.Login = "xena"
				n.IssueTypeChangedEvent.CreatedAt = dt(ts)
				n.IssueTypeChangedEvent.PrevIssueType.Name = "Task"
				n.IssueTypeChangedEvent.IssueType.Name = "Bug"
				return n
			}(),
			wantType:    "IssueTypeChangedEvent",
			wantActor:   "xena",
			wantSummary: `changed issue type: "Task" → "Bug"`,
		},
		{
			name: "IssueTypeRemovedEvent uses the removed verb",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueTypeRemovedEvent"}
				n.IssueTypeRemovedEvent.Actor.Login = "xena"
				n.IssueTypeRemovedEvent.CreatedAt = dt(ts)
				n.IssueTypeRemovedEvent.IssueType.Name = "Bug"
				return n
			}(),
			wantType:    "IssueTypeRemovedEvent",
			wantActor:   "xena",
			wantSummary: `removed issue type "Bug"`,
		},
		{
			name: "IssueCommentPinnedEvent uses fixed wording",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueCommentPinnedEvent"}
				n.IssueCommentPinnedEvent.Actor.Login = "yui"
				n.IssueCommentPinnedEvent.CreatedAt = dt(ts)
				n.IssueCommentPinnedEvent.IssueComment.DatabaseID = int64(101)
				return n
			}(),
			wantType:    "IssueCommentPinnedEvent",
			wantActor:   "yui",
			wantSummary: "pinned comment",
		},
		{
			name: "IssueCommentUnpinnedEvent uses fixed wording",
			node: func() prTimelineNode {
				n := prTimelineNode{Typename: "IssueCommentUnpinnedEvent"}
				n.IssueCommentUnpinnedEvent.Actor.Login = "yui"
				n.IssueCommentUnpinnedEvent.CreatedAt = dt(ts)
				return n
			}(),
			wantType:    "IssueCommentUnpinnedEvent",
			wantActor:   "yui",
			wantSummary: "unpinned comment",
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

// TestDispatchIssueNode_subIssueFamily exercises the sub-issue / blocking
// family on the Issue dispatcher — these are the events that motivated
// adding the new fragments, since they fire predominantly on issues rather
// than PRs.
func TestDispatchIssueNode_subIssueFamily(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 5, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		node        issueTimelineNode
		wantSummary string
	}{
		{
			name: "SubIssueAddedEvent surfaces the linked sub-issue",
			node: func() issueTimelineNode {
				n := issueTimelineNode{Typename: "SubIssueAddedEvent"}
				n.SubIssueAddedEvent.Actor.Login = "alice"
				n.SubIssueAddedEvent.CreatedAt = dt(ts)
				n.SubIssueAddedEvent.SubIssue.Number = 42
				n.SubIssueAddedEvent.SubIssue.Title = "Child task"
				n.SubIssueAddedEvent.SubIssue.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantSummary: "added sub-issue octo/repo#42: Child task",
		},
		{
			name: "ParentIssueAddedEvent surfaces the linked parent",
			node: func() issueTimelineNode {
				n := issueTimelineNode{Typename: "ParentIssueAddedEvent"}
				n.ParentIssueAddedEvent.Actor.Login = "alice"
				n.ParentIssueAddedEvent.CreatedAt = dt(ts)
				n.ParentIssueAddedEvent.Parent.Number = 1
				n.ParentIssueAddedEvent.Parent.Title = "Tracking issue"
				n.ParentIssueAddedEvent.Parent.Repository.NameWithOwner = "octo/repo"
				return n
			}(),
			wantSummary: "added parent issue octo/repo#1: Tracking issue",
		},
		{
			name: "BlockingAddedEvent surfaces the blocked downstream issue",
			node: func() issueTimelineNode {
				n := issueTimelineNode{Typename: "BlockingAddedEvent"}
				n.BlockingAddedEvent.Actor.Login = "alice"
				n.BlockingAddedEvent.CreatedAt = dt(ts)
				n.BlockingAddedEvent.BlockedIssue.Number = 5
				n.BlockingAddedEvent.BlockedIssue.Title = "Consumer"
				n.BlockingAddedEvent.BlockedIssue.Repository.NameWithOwner = "octo/consumer"
				return n
			}(),
			wantSummary: "blocking octo/consumer#5: Consumer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dispatchIssueNode(tt.node)
			if got.Summary != tt.wantSummary {
				t.Errorf("Summary = %q, want %q", got.Summary, tt.wantSummary)
			}
			if got.Actor != "alice" {
				t.Errorf("Actor = %q, want %q", got.Actor, "alice")
			}
		})
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

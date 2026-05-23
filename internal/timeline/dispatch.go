package timeline

import (
	"fmt"

	"github.com/shurcooL/githubv4"
)

// dispatchPRNode converts one prTimelineNode into the normalized Event. Falls
// back to a type-only Event when the GraphQL __typename is not one our table
// knows about — newly added union members surface this way until they get a
// dedicated handler.
//
//nolint:cyclop,funlen,gocyclo,goconst // GraphQL union dispatcher — complexity and string repetition mirror the schema's union surface.
func dispatchPRNode(n prTimelineNode) Event {
	t := string(n.Typename)
	switch t {
	// Shared with Issue
	case "IssueComment":
		return handleIssueComment(t, n.IssueComment)
	case "LabeledEvent", "UnlabeledEvent":
		return handleLabeled(t, pickLabeled(t, n))
	case "AssignedEvent", "UnassignedEvent":
		return handleAssigned(t, pickAssigned(t, n))
	case "MilestonedEvent", "DemilestonedEvent":
		return handleMilestoned(t, pickMilestoned(t, n))
	case "RenamedTitleEvent":
		return handleRenamedTitle(t, n.RenamedTitleEvent)
	case "ClosedEvent":
		return handleClosed(t, n.ClosedEvent)
	case "ReopenedEvent":
		return handleSimpleWord(t, "reopened", n.ReopenedEvent.commonEvent)
	case "LockedEvent":
		return handleLocked(t, n.LockedEvent)
	case "UnlockedEvent":
		return handleSimpleWord(t, "unlocked", n.UnlockedEvent)
	case "PinnedEvent":
		return handleSimpleWord(t, "pinned", n.PinnedEvent)
	case "UnpinnedEvent":
		return handleSimpleWord(t, "unpinned", n.UnpinnedEvent)
	case "SubscribedEvent", "UnsubscribedEvent", "MentionedEvent", "CommentDeletedEvent",
		"UnmarkedAsDuplicateEvent":
		return handleSimpleWord(t, "", pickCommon(t, n))
	case "CrossReferencedEvent":
		return handleCrossReferenced(t, n.CrossReferencedEvent)
	case "ReferencedEvent":
		return handleReferenced(t, n.ReferencedEvent)
	case "MarkedAsDuplicateEvent":
		return handleMarkedAsDuplicate(t, n.MarkedAsDuplicateEvent)
	case "ConvertedToDiscussionEvent":
		return handleConvertedToDiscussion(t, n.ConvertedToDiscussionEvent)
	case "TransferredEvent":
		return handleTransferred(t, n.TransferredEvent)
	case "ConnectedEvent":
		return handleConnected(t, "connected to", n.ConnectedEvent)
	case "DisconnectedEvent":
		return handleConnected(t, "disconnected from", n.DisconnectedEvent)
	case "AddedToProjectEvent":
		return handleProjectChange(t, "added to project", n.AddedToProjectEvent)
	case "RemovedFromProjectEvent":
		return handleProjectChange(t, "removed from project", n.RemovedFromProjectEvent)
	case "ConvertedNoteToIssueEvent":
		return handleProjectChange(t, "converted note to issue in project", n.ConvertedNoteToIssueEvent)
	case "MovedColumnsInProjectEvent":
		return handleMovedColumns(t, n.MovedColumnsInProjectEvent)
	case "UserBlockedEvent":
		return handleUserBlocked(t, n.UserBlockedEvent)

	// PR-only
	case "PullRequestCommit":
		return handlePullRequestCommit(t, n.PullRequestCommit)
	case "PullRequestReview":
		return handlePullRequestReview(t, n.PullRequestReview)
	case "PullRequestReviewThread":
		return Event{Type: t, Ref: Ref{NodeID: graphqlIDString(n.PullRequestReviewThread.ID)}}
	case "PullRequestRevisionMarker":
		return Event{
			Type:      t,
			Timestamp: n.PullRequestRevisionMarker.CreatedAt.Time,
			Ref:       Ref{SHA: string(n.PullRequestRevisionMarker.LastSeenCommit.OID)},
		}
	case "PullRequestCommitCommentThread":
		return Event{
			Type: t,
			Ref: Ref{
				NodeID: graphqlIDString(n.PullRequestCommitCommentThread.ID),
				SHA:    string(n.PullRequestCommitCommentThread.Commit.OID),
			},
		}
	case "MergedEvent":
		return handleMerged(t, n.MergedEvent)
	case "ReviewRequestedEvent", "ReviewRequestRemovedEvent":
		return handleReviewRequested(t, pickReviewReq(t, n))
	case "ReviewDismissedEvent":
		return handleReviewDismissed(t, n.ReviewDismissedEvent)
	case "ReadyForReviewEvent":
		return handleSimpleWord(t, "marked ready for review", n.ReadyForReviewEvent.commonEvent)
	case "ConvertToDraftEvent":
		return handleSimpleWord(t, "converted to draft", n.ConvertToDraftEvent.commonEvent)
	case "HeadRefForcePushedEvent":
		return handleForcePushed(t, "force-pushed head", n.HeadRefForcePushedEvent)
	case "BaseRefForcePushedEvent":
		return handleForcePushed(t, "force-pushed base", n.BaseRefForcePushedEvent)
	case "BaseRefChangedEvent":
		return handleBaseRefChanged(t, n.BaseRefChangedEvent)
	case "BaseRefDeletedEvent":
		return handleBaseRefDeleted(t, n.BaseRefDeletedEvent)
	case "HeadRefDeletedEvent":
		return handleHeadRefDeleted(t, n.HeadRefDeletedEvent)
	case "HeadRefRestoredEvent":
		return handleSimpleWord(t, "restored head ref", n.HeadRefRestoredEvent.commonEvent)
	case "DeployedEvent":
		return handleDeployed(t, n.DeployedEvent)
	case "DeploymentEnvironmentChangedEvent":
		return handleDeploymentEnvChanged(t, n.DeploymentEnvironmentChangedEvent)
	case "AutoMergeEnabledEvent":
		return handleSimpleWord(t, "auto-merge enabled", n.AutoMergeEnabledEvent.commonEvent)
	case "AutoMergeDisabledEvent":
		return handleSimpleWord(t, "auto-merge disabled", n.AutoMergeDisabledEvent.commonEvent)
	case "AutoRebaseEnabledEvent":
		return handleSimpleWord(t, "auto-rebase enabled", n.AutoRebaseEnabledEvent.commonEvent)
	case "AutoSquashEnabledEvent":
		return handleSimpleWord(t, "auto-squash enabled", n.AutoSquashEnabledEvent.commonEvent)
	case "AutomaticBaseChangeSucceededEvent":
		return handleAutomaticBaseChange(t, "auto base change succeeded", n.AutomaticBaseChangeSucceededEvent)
	case "AutomaticBaseChangeFailedEvent":
		return handleAutomaticBaseChange(t, "auto base change failed", n.AutomaticBaseChangeFailedEvent)
	case "AddedToMergeQueueEvent":
		return handleSimpleWord(t, "added to merge queue", n.AddedToMergeQueueEvent.commonEvent)
	case "RemovedFromMergeQueueEvent":
		return handleSimpleWord(t, "removed from merge queue", n.RemovedFromMergeQueueEvent.commonEvent)
	}
	return Event{Type: t}
}

// dispatchIssueNode is the IssueTimelineItems counterpart of [dispatchPRNode].
//
//nolint:cyclop,funlen,gocyclo // GraphQL union dispatcher — see dispatchPRNode.
func dispatchIssueNode(n issueTimelineNode) Event {
	t := string(n.Typename)
	switch t {
	case "IssueComment":
		return handleIssueComment(t, n.IssueComment)
	case "LabeledEvent":
		return handleLabeled(t, n.LabeledEvent)
	case "UnlabeledEvent":
		return handleLabeled(t, n.UnlabeledEvent)
	case "AssignedEvent":
		return handleAssigned(t, n.AssignedEvent)
	case "UnassignedEvent":
		return handleAssigned(t, n.UnassignedEvent)
	case "MilestonedEvent":
		return handleMilestoned(t, n.MilestonedEvent)
	case "DemilestonedEvent":
		return handleMilestoned(t, n.DemilestonedEvent)
	case "RenamedTitleEvent":
		return handleRenamedTitle(t, n.RenamedTitleEvent)
	case "ClosedEvent":
		return handleClosed(t, n.ClosedEvent)
	case "ReopenedEvent":
		return handleSimpleWord(t, "reopened", n.ReopenedEvent.commonEvent)
	case "LockedEvent":
		return handleLocked(t, n.LockedEvent)
	case "UnlockedEvent":
		return handleSimpleWord(t, "unlocked", n.UnlockedEvent)
	case "PinnedEvent":
		return handleSimpleWord(t, "pinned", n.PinnedEvent)
	case "UnpinnedEvent":
		return handleSimpleWord(t, "unpinned", n.UnpinnedEvent)
	case "SubscribedEvent":
		return handleSimpleWord(t, "", n.SubscribedEvent)
	case "UnsubscribedEvent":
		return handleSimpleWord(t, "", n.UnsubscribedEvent)
	case "MentionedEvent":
		return handleSimpleWord(t, "", n.MentionedEvent)
	case "CommentDeletedEvent":
		return handleSimpleWord(t, "", n.CommentDeletedEvent)
	case "UnmarkedAsDuplicateEvent":
		return handleSimpleWord(t, "", n.UnmarkedAsDuplicateEvent)
	case "CrossReferencedEvent":
		return handleCrossReferenced(t, n.CrossReferencedEvent)
	case "ReferencedEvent":
		return handleReferenced(t, n.ReferencedEvent)
	case "MarkedAsDuplicateEvent":
		return handleMarkedAsDuplicate(t, n.MarkedAsDuplicateEvent)
	case "ConvertedToDiscussionEvent":
		return handleConvertedToDiscussion(t, n.ConvertedToDiscussionEvent)
	case "TransferredEvent":
		return handleTransferred(t, n.TransferredEvent)
	case "ConnectedEvent":
		return handleConnected(t, "connected to", n.ConnectedEvent)
	case "DisconnectedEvent":
		return handleConnected(t, "disconnected from", n.DisconnectedEvent)
	case "AddedToProjectEvent":
		return handleProjectChange(t, "added to project", n.AddedToProjectEvent)
	case "RemovedFromProjectEvent":
		return handleProjectChange(t, "removed from project", n.RemovedFromProjectEvent)
	case "ConvertedNoteToIssueEvent":
		return handleProjectChange(t, "converted note to issue in project", n.ConvertedNoteToIssueEvent)
	case "MovedColumnsInProjectEvent":
		return handleMovedColumns(t, n.MovedColumnsInProjectEvent)
	case "UserBlockedEvent":
		return handleUserBlocked(t, n.UserBlockedEvent)
	}
	return Event{Type: t}
}

// --- helpers to keep the PR dispatcher's case bodies one-liner -----------

func pickLabeled(t string, n prTimelineNode) labeledEventFragment {
	if t == "UnlabeledEvent" {
		return n.UnlabeledEvent
	}
	return n.LabeledEvent
}

func pickAssigned(t string, n prTimelineNode) assignedEventFragment {
	if t == "UnassignedEvent" {
		return n.UnassignedEvent
	}
	return n.AssignedEvent
}

func pickMilestoned(t string, n prTimelineNode) milestonedEventFragment {
	if t == "DemilestonedEvent" {
		return n.DemilestonedEvent
	}
	return n.MilestonedEvent
}

func pickReviewReq(t string, n prTimelineNode) reviewRequestedEventFragment {
	if t == "ReviewRequestRemovedEvent" {
		return n.ReviewRequestRemovedEvent
	}
	return n.ReviewRequestedEvent
}

// pickCommon returns the commonEvent block for typenames that we represent
// with the bare commonEvent fragment.
func pickCommon(t string, n prTimelineNode) commonEvent {
	switch t {
	case "SubscribedEvent":
		return n.SubscribedEvent
	case "UnsubscribedEvent":
		return n.UnsubscribedEvent
	case "MentionedEvent":
		return n.MentionedEvent
	case "CommentDeletedEvent":
		return n.CommentDeletedEvent
	case "UnmarkedAsDuplicateEvent":
		return n.UnmarkedAsDuplicateEvent
	}
	return commonEvent{}
}

// --- handlers ------------------------------------------------------------

func handleIssueComment(typename string, f issueCommentFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Author.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   truncate(firstLine(string(f.Body))),
		Ref: Ref{
			NodeID:    graphqlIDString(f.ID),
			CommentID: f.DatabaseID,
			URL:       uriString(f.URL),
		},
	}
}

func handleLabeled(typename string, f labeledEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   string(f.Label.Name),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleAssigned(typename string, f assignedEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   f.Assignee.pick(),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleMilestoned(typename string, f milestonedEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   string(f.MilestoneTitle),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleRenamedTitle(typename string, f renamedTitleEventFragment) Event {
	prev := truncate(string(f.PreviousTitle))
	curr := truncate(string(f.CurrentTitle))
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("renamed: %q → %q", prev, curr),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleClosed(typename string, f closedEventFragment) Event {
	summary := "closed"
	if reason := string(f.StateReason); reason != "" {
		summary = fmt.Sprintf("closed: %s", reason)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleLocked(typename string, f lockedEventFragment) Event {
	summary := "locked"
	if reason := string(f.LockReason); reason != "" {
		summary = fmt.Sprintf("locked: %s", reason)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleSimpleWord(typename, word string, f commonEvent) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   word,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleCrossReferenced(typename string, f crossReferencedEventFragment) Event {
	repo, num, title := subjectInfo(f.Source)
	summary := fmt.Sprintf("referenced from %s#%d", repo, num)
	if title != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(title))
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleReferenced(typename string, f referencedEventFragment) Event {
	summary := "referenced"
	if sha := string(f.Commit.OID); sha != "" {
		summary = fmt.Sprintf("referenced from commit %s", shortSHA(sha))
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID), SHA: string(f.Commit.OID)},
	}
}

func handleMarkedAsDuplicate(typename string, f markedAsDuplicateEventFragment) Event {
	repo, num, _ := subjectInfo(f.Canonical)
	summary := "marked as duplicate"
	if num != 0 {
		summary = fmt.Sprintf("duplicate of %s#%d", repo, num)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleConvertedToDiscussion(typename string, f convertedToDiscussionEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("converted to discussion #%d", int(f.Discussion.Number)),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleTransferred(typename string, f transferredEventFragment) Event {
	summary := "transferred"
	if from := string(f.FromRepository.NameWithOwner); from != "" {
		summary = fmt.Sprintf("transferred from %s", from)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleConnected(typename, verb string, f connectedEventFragment) Event {
	repo, num, _ := subjectInfo(f.Subject)
	summary := verb
	if num != 0 {
		summary = fmt.Sprintf("%s %s#%d", verb, repo, num)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleProjectChange(typename, verb string, f projectChangeEventFragment) Event {
	summary := verb
	if name := string(f.Project.Name); name != "" {
		summary = fmt.Sprintf("%s %q", verb, name)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleMovedColumns(typename string, f movedColumnsInProjectEventFragment) Event {
	prev := string(f.PreviousProjectColumnName)
	curr := string(f.ProjectColumnName)
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("moved %q → %q", prev, curr),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleUserBlocked(typename string, f userBlockedEventFragment) Event {
	target := string(f.Subject.Login)
	summary := "blocked user"
	if target != "" {
		summary = fmt.Sprintf("blocked %s", target)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handlePullRequestCommit(typename string, f pullRequestCommitFragment) Event {
	actor := string(f.Commit.Author.User.Login)
	if actor == "" {
		actor = string(f.Commit.Author.Name)
	}
	return Event{
		Type:      typename,
		Actor:     actor,
		Timestamp: f.Commit.CommittedDate.Time,
		Summary:   truncate(firstLine(string(f.Commit.MessageHeadline))),
		Ref: Ref{
			NodeID: graphqlIDString(f.ID),
			SHA:    string(f.Commit.OID),
			URL:    uriString(f.URL),
		},
	}
}

func handlePullRequestReview(typename string, f pullRequestReviewFragment) Event {
	ts := f.SubmittedAt.Time
	if ts.IsZero() {
		ts = f.CreatedAt.Time
	}
	summary := string(f.State)
	if body := firstLine(string(f.Body)); body != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(body))
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Author.Login),
		Timestamp: ts,
		Summary:   summary,
		Ref: Ref{
			NodeID:   graphqlIDString(f.ID),
			ReviewID: f.DatabaseID,
			URL:      uriString(f.URL),
		},
	}
}

func handleMerged(typename string, f mergedEventFragment) Event {
	actor := string(f.Actor.Login)
	summary := "merged"
	if actor != "" {
		summary = fmt.Sprintf("merged by %s", actor)
	}
	return Event{
		Type:      typename,
		Actor:     actor,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref: Ref{
			NodeID: graphqlIDString(f.ID),
			SHA:    string(f.Commit.OID),
			URL:    uriString(f.URL),
		},
	}
}

func handleReviewRequested(typename string, f reviewRequestedEventFragment) Event {
	target := pickReviewerLabel(f.RequestedReviewer)
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   target,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleReviewDismissed(typename string, f reviewDismissedEventFragment) Event {
	author := string(f.Review.Author.Login)
	summary := "dismissed review"
	if author != "" {
		summary = fmt.Sprintf("dismissed review by %s", author)
	}
	if msg := firstLine(string(f.DismissalMessage)); msg != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(msg))
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleForcePushed(typename, verb string, f forcePushedEventFragment) Event {
	before := string(f.BeforeCommit.OID)
	after := string(f.AfterCommit.OID)
	summary := verb
	switch {
	case before != "" && after != "":
		summary = fmt.Sprintf("%s: %s → %s", verb, shortSHA(before), shortSHA(after))
	case after != "":
		summary = fmt.Sprintf("%s to %s", verb, shortSHA(after))
	case before != "":
		summary = fmt.Sprintf("%s from %s", verb, shortSHA(before))
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID), SHA: after},
	}
}

func handleBaseRefChanged(typename string, f baseRefChangedEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("base ref changed: %s → %s", string(f.PreviousRefName), string(f.CurrentRefName)),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleBaseRefDeleted(typename string, f baseRefDeletedEventFragment) Event {
	summary := "base ref deleted"
	if name := string(f.BaseRefName); name != "" {
		summary = fmt.Sprintf("base ref deleted: %s", name)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleHeadRefDeleted(typename string, f headRefDeletedEventFragment) Event {
	summary := "head ref deleted"
	if name := string(f.HeadRefName); name != "" {
		summary = fmt.Sprintf("deleted branch %s", name)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleDeployed(typename string, f deployedEventFragment) Event {
	env := string(f.Deployment.Environment)
	summary := "deployed"
	if env != "" {
		summary = fmt.Sprintf("deployed to %s", env)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleDeploymentEnvChanged(typename string, f deploymentEnvironmentChangedEventFragment) Event {
	env := string(f.DeploymentStatus.Deployment.Environment)
	summary := "deployment environment changed"
	if env != "" {
		summary = fmt.Sprintf("deployment environment changed to %s", env)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleAutomaticBaseChange(typename, verb string, f automaticBaseChangeEventFragment) Event {
	summary := verb
	if oldBase, newBase := string(f.OldBase), string(f.NewBase); oldBase != "" && newBase != "" {
		summary = fmt.Sprintf("%s: %s → %s", verb, oldBase, newBase)
	}
	return Event{
		Type:      typename,
		Actor:     string(f.Actor.Login),
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

// --- value converters ----------------------------------------------------

// graphqlIDString defensively turns a GraphQL ID into a string. The library
// types ID as `interface{}` because IDs are opaque scalars, but in practice
// GitHub always sends them as JSON strings.
func graphqlIDString(id githubv4.ID) string {
	switch v := id.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

// uriString extracts the string form of a githubv4.URI, returning "" for the
// zero value (no underlying [url.URL]).
func uriString(u githubv4.URI) string {
	if u.URL == nil {
		return ""
	}
	return u.URL.String()
}

// shortSHA truncates a Git SHA to the conventional 7-character display form.
func shortSHA(sha string) string {
	const short = 7
	if len(sha) <= short {
		return sha
	}
	return sha[:short]
}

// subjectInfo extracts the repository / number / title common to Issue and
// PullRequest variants of a ReferencedSubject union value. Returns (repo,
// number, title); a zero number means neither side was populated.
func subjectInfo(s subjectFragment) (string, int, string) {
	if n := int(s.PullRequest.Number); n != 0 {
		return string(s.PullRequest.Repository.NameWithOwner), n, string(s.PullRequest.Title)
	}
	if n := int(s.Issue.Number); n != 0 {
		return string(s.Issue.Repository.NameWithOwner), n, string(s.Issue.Title)
	}
	return "", 0, ""
}

// pickReviewerLabel produces a short label for whichever member of the
// RequestedReviewer union is populated.
func pickReviewerLabel(r userOrTeamFragment) string {
	if login := string(r.User.Login); login != "" {
		return login
	}
	if login := string(r.Bot.Login); login != "" {
		return login
	}
	if login := string(r.Mannequin.Login); login != "" {
		return login
	}
	if slug := string(r.Team.Slug); slug != "" {
		return "team:" + slug
	}
	return ""
}

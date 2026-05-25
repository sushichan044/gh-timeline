package timeline

import "fmt"

// dispatchPRNode converts one prTimelineNode into the normalized Event. Falls
// back to a type-only Event when the GraphQL __typename is not one our table
// knows about — newly added union members surface this way until they get a
// dedicated handler.
//
//nolint:cyclop,funlen,gocyclo,goconst // GraphQL union dispatcher — complexity and string repetition mirror the schema's union surface.
func dispatchPRNode(n prTimelineNode) Event {
	t := n.Typename
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
	case "SubIssueAddedEvent":
		return handleSubIssue(t, "added sub-issue", n.SubIssueAddedEvent)
	case "SubIssueRemovedEvent":
		return handleSubIssue(t, "removed sub-issue", n.SubIssueRemovedEvent)
	case "ParentIssueAddedEvent":
		return handleParentIssue(t, "added parent issue", n.ParentIssueAddedEvent)
	case "ParentIssueRemovedEvent":
		return handleParentIssue(t, "removed parent issue", n.ParentIssueRemovedEvent)
	case "BlockedByAddedEvent":
		return handleBlockedBy(t, "blocked by", n.BlockedByAddedEvent)
	case "BlockedByRemovedEvent":
		return handleBlockedBy(t, "no longer blocked by", n.BlockedByRemovedEvent)
	case "BlockingAddedEvent":
		return handleBlocking(t, "blocking", n.BlockingAddedEvent)
	case "BlockingRemovedEvent":
		return handleBlocking(t, "no longer blocking", n.BlockingRemovedEvent)
	case "AddedToProjectV2Event":
		return handleProjectV2Change(t, "added to project", n.AddedToProjectV2Event)
	case "RemovedFromProjectV2Event":
		return handleProjectV2Change(t, "removed from project", n.RemovedFromProjectV2Event)
	case "ProjectV2ItemStatusChangedEvent":
		return handleProjectV2StatusChanged(t, n.ProjectV2ItemStatusChangedEvent)
	case "ConvertedFromDraftEvent":
		return handleProjectV2Change(t, "converted from draft", n.ConvertedFromDraftEvent)
	case "IssueFieldAddedEvent":
		return handleIssueFieldAdded(t, n.IssueFieldAddedEvent)
	case "IssueFieldChangedEvent":
		return handleIssueFieldChanged(t, n.IssueFieldChangedEvent)
	case "IssueFieldRemovedEvent":
		return handleIssueFieldRemoved(t, n.IssueFieldRemovedEvent)
	case "IssueTypeAddedEvent":
		return handleIssueType(t, "set issue type to", n.IssueTypeAddedEvent)
	case "IssueTypeChangedEvent":
		return handleIssueTypeChanged(t, n.IssueTypeChangedEvent)
	case "IssueTypeRemovedEvent":
		return handleIssueType(t, "removed issue type", n.IssueTypeRemovedEvent)
	case "IssueCommentPinnedEvent":
		return handleIssueCommentPin(t, "pinned comment", n.IssueCommentPinnedEvent)
	case "IssueCommentUnpinnedEvent":
		return handleIssueCommentPin(t, "unpinned comment", n.IssueCommentUnpinnedEvent)

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
			Ref:       Ref{SHA: n.PullRequestRevisionMarker.LastSeenCommit.OID},
		}
	case "PullRequestCommitCommentThread":
		return Event{
			Type: t,
			Ref: Ref{
				NodeID: graphqlIDString(n.PullRequestCommitCommentThread.ID),
				SHA:    n.PullRequestCommitCommentThread.Commit.OID,
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
	t := n.Typename
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
	case "SubIssueAddedEvent":
		return handleSubIssue(t, "added sub-issue", n.SubIssueAddedEvent)
	case "SubIssueRemovedEvent":
		return handleSubIssue(t, "removed sub-issue", n.SubIssueRemovedEvent)
	case "ParentIssueAddedEvent":
		return handleParentIssue(t, "added parent issue", n.ParentIssueAddedEvent)
	case "ParentIssueRemovedEvent":
		return handleParentIssue(t, "removed parent issue", n.ParentIssueRemovedEvent)
	case "BlockedByAddedEvent":
		return handleBlockedBy(t, "blocked by", n.BlockedByAddedEvent)
	case "BlockedByRemovedEvent":
		return handleBlockedBy(t, "no longer blocked by", n.BlockedByRemovedEvent)
	case "BlockingAddedEvent":
		return handleBlocking(t, "blocking", n.BlockingAddedEvent)
	case "BlockingRemovedEvent":
		return handleBlocking(t, "no longer blocking", n.BlockingRemovedEvent)
	case "AddedToProjectV2Event":
		return handleProjectV2Change(t, "added to project", n.AddedToProjectV2Event)
	case "RemovedFromProjectV2Event":
		return handleProjectV2Change(t, "removed from project", n.RemovedFromProjectV2Event)
	case "ProjectV2ItemStatusChangedEvent":
		return handleProjectV2StatusChanged(t, n.ProjectV2ItemStatusChangedEvent)
	case "ConvertedFromDraftEvent":
		return handleProjectV2Change(t, "converted from draft", n.ConvertedFromDraftEvent)
	case "IssueFieldAddedEvent":
		return handleIssueFieldAdded(t, n.IssueFieldAddedEvent)
	case "IssueFieldChangedEvent":
		return handleIssueFieldChanged(t, n.IssueFieldChangedEvent)
	case "IssueFieldRemovedEvent":
		return handleIssueFieldRemoved(t, n.IssueFieldRemovedEvent)
	case "IssueTypeAddedEvent":
		return handleIssueType(t, "set issue type to", n.IssueTypeAddedEvent)
	case "IssueTypeChangedEvent":
		return handleIssueTypeChanged(t, n.IssueTypeChangedEvent)
	case "IssueTypeRemovedEvent":
		return handleIssueType(t, "removed issue type", n.IssueTypeRemovedEvent)
	case "IssueCommentPinnedEvent":
		return handleIssueCommentPin(t, "pinned comment", n.IssueCommentPinnedEvent)
	case "IssueCommentUnpinnedEvent":
		return handleIssueCommentPin(t, "unpinned comment", n.IssueCommentUnpinnedEvent)
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
		Actor:     f.Author.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   truncate(firstLine(f.Body)),
		Ref: Ref{
			NodeID:    graphqlIDString(f.ID),
			CommentID: f.DatabaseID,
			URL:       uriString(f.URL),
		},
	}
}

func handleLabeled(typename string, f labeledEventFragment) Event {
	verb := "added label"
	if typename == "UnlabeledEvent" {
		verb = "removed label"
	}
	summary := verb
	if name := f.Label.Name; name != "" {
		summary = fmt.Sprintf("%s %s", verb, name)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleAssigned(typename string, f assignedEventFragment) Event {
	verb := "assigned"
	if typename == "UnassignedEvent" {
		verb = "unassigned"
	}
	summary := verb
	if target := f.Assignee.pick(); target != "" {
		summary = fmt.Sprintf("%s %s", verb, target)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleMilestoned(typename string, f milestonedEventFragment) Event {
	verb := "added to milestone"
	if typename == "DemilestonedEvent" {
		verb = "removed from milestone"
	}
	summary := verb
	if title := f.MilestoneTitle; title != "" {
		summary = fmt.Sprintf("%s %q", verb, title)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleRenamedTitle(typename string, f renamedTitleEventFragment) Event {
	prev := truncate(f.PreviousTitle)
	curr := truncate(f.CurrentTitle)
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("renamed: %q → %q", prev, curr),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleClosed(typename string, f closedEventFragment) Event {
	summary := "closed"
	if reason := f.StateReason; reason != "" {
		summary = fmt.Sprintf("closed: %s", reason)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleLocked(typename string, f lockedEventFragment) Event {
	summary := "locked"
	if reason := f.LockReason; reason != "" {
		summary = fmt.Sprintf("locked: %s", reason)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleSimpleWord(typename, word string, f commonEvent) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   word,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleCrossReferenced(typename string, f crossReferencedEventFragment) Event {
	repo, num := subjectInfo(f.Source)
	summary := fmt.Sprintf("referenced from %s#%d", repo, num)
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleReferenced(typename string, f referencedEventFragment) Event {
	summary := "referenced"
	if sha := f.Commit.OID; sha != "" {
		summary = fmt.Sprintf("referenced from commit %s", shortSHA(sha))
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID), SHA: f.Commit.OID},
	}
}

func handleMarkedAsDuplicate(typename string, f markedAsDuplicateEventFragment) Event {
	repo, num := subjectInfo(f.Canonical)
	summary := "marked as duplicate"
	if num != 0 {
		summary = fmt.Sprintf("duplicate of %s#%d", repo, num)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleConvertedToDiscussion(typename string, f convertedToDiscussionEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("converted to discussion #%d", int(f.Discussion.Number)),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleTransferred(typename string, f transferredEventFragment) Event {
	summary := "transferred"
	if from := f.FromRepository.NameWithOwner; from != "" {
		summary = fmt.Sprintf("transferred from %s", from)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleConnected(typename, verb string, f connectedEventFragment) Event {
	repo, num := subjectInfo(f.Subject)
	summary := verb
	if num != 0 {
		summary = fmt.Sprintf("%s %s#%d", verb, repo, num)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleProjectChange(typename, verb string, f projectChangeEventFragment) Event {
	summary := verb
	if name := f.Project.Name; name != "" {
		summary = fmt.Sprintf("%s %q", verb, name)
	}
	if col := f.ProjectColumnName; col != "" {
		summary = fmt.Sprintf("%s (column %q)", summary, col)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleMovedColumns(typename string, f movedColumnsInProjectEventFragment) Event {
	prev := f.PreviousProjectColumnName
	curr := f.ProjectColumnName
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("moved %q → %q", prev, curr),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleUserBlocked(typename string, f userBlockedEventFragment) Event {
	target := f.Subject.Login
	summary := "blocked user"
	if target != "" {
		summary = fmt.Sprintf("blocked %s", target)
	}
	if dur := f.BlockDuration; dur != "" {
		summary = fmt.Sprintf("%s (%s)", summary, dur)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handlePullRequestCommit(typename string, f pullRequestCommitFragment) Event {
	actor := f.Commit.Author.User.Login
	if actor == "" {
		actor = f.Commit.Author.Name
	}
	return Event{
		Type:      typename,
		Actor:     actor,
		Timestamp: f.Commit.CommittedDate.Time,
		Summary:   truncate(firstLine(f.Commit.MessageHeadline)),
		Ref: Ref{
			NodeID: graphqlIDString(f.ID),
			SHA:    f.Commit.OID,
			URL:    uriString(f.URL),
		},
	}
}

func handlePullRequestReview(typename string, f pullRequestReviewFragment) Event {
	ts := f.SubmittedAt.Time
	if ts.IsZero() {
		ts = f.CreatedAt.Time
	}
	summary := f.State
	if body := firstLine(f.Body); body != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(body))
	}
	return Event{
		Type:      typename,
		Actor:     f.Author.Login,
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
	sha := f.Commit.OID
	mergeRef := f.MergeRefName
	summary := "merged"
	switch {
	case sha != "" && mergeRef != "":
		summary = fmt.Sprintf("merged %s into %s", shortSHA(sha), mergeRef)
	case sha != "":
		summary = fmt.Sprintf("merged %s", shortSHA(sha))
	case mergeRef != "":
		summary = fmt.Sprintf("merged into %s", mergeRef)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref: Ref{
			NodeID: graphqlIDString(f.ID),
			SHA:    sha,
			URL:    uriString(f.URL),
		},
	}
}

func handleReviewRequested(typename string, f reviewRequestedEventFragment) Event {
	verb := "requested review from"
	if typename == "ReviewRequestRemovedEvent" {
		verb = "removed review request from"
	}
	summary := verb
	if target := pickReviewerLabel(f.RequestedReviewer); target != "" {
		summary = fmt.Sprintf("%s %s", verb, target)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleReviewDismissed(typename string, f reviewDismissedEventFragment) Event {
	author := f.Review.Author.Login
	summary := "dismissed review"
	if author != "" {
		summary = fmt.Sprintf("dismissed review by %s", author)
	}
	if msg := firstLine(f.DismissalMessage); msg != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(msg))
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleForcePushed(typename, verb string, f forcePushedEventFragment) Event {
	before := f.BeforeCommit.OID
	after := f.AfterCommit.OID
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
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID), SHA: after},
	}
}

func handleBaseRefChanged(typename string, f baseRefChangedEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   fmt.Sprintf("base ref changed: %s → %s", f.PreviousRefName, f.CurrentRefName),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleBaseRefDeleted(typename string, f baseRefDeletedEventFragment) Event {
	summary := "base ref deleted"
	if name := f.BaseRefName; name != "" {
		summary = fmt.Sprintf("base ref deleted: %s", name)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleHeadRefDeleted(typename string, f headRefDeletedEventFragment) Event {
	summary := "head ref deleted"
	if name := f.HeadRefName; name != "" {
		summary = fmt.Sprintf("deleted branch %s", name)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleDeployed(typename string, f deployedEventFragment) Event {
	env := f.Deployment.Environment
	summary := "deployed"
	if env != "" {
		summary = fmt.Sprintf("deployed to %s", env)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleDeploymentEnvChanged(typename string, f deploymentEnvironmentChangedEventFragment) Event {
	env := f.DeploymentStatus.Deployment.Environment
	summary := "deployment environment changed"
	if env != "" {
		summary = fmt.Sprintf("deployment environment changed to %s", env)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

// issueRefSummary appends the standard "OWNER/REPO#N: TITLE" suffix to verb
// when the referenced issue's number is non-zero. The handlers for sub-issue,
// parent-issue, and blocking pairs all share this shape.
func issueRefSummary(verb string, ref issueRefFragment) string {
	num := int(ref.Number)
	if num == 0 {
		return verb
	}
	return fmt.Sprintf("%s %s#%d", verb, ref.Repository.NameWithOwner, num)
}

func handleSubIssue(typename, verb string, f subIssueEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   issueRefSummary(verb, f.SubIssue),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleParentIssue(typename, verb string, f parentIssueEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   issueRefSummary(verb, f.Parent),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleBlockedBy(typename, verb string, f blockedByEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   issueRefSummary(verb, f.BlockingIssue),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleBlocking(typename, verb string, f blockingEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   issueRefSummary(verb, f.BlockedIssue),
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleProjectV2Change(typename, verb string, f projectV2ChangeEventFragment) Event {
	summary := verb
	if title := f.Project.Title; title != "" {
		summary = fmt.Sprintf("%s %q", verb, title)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleProjectV2StatusChanged(typename string, f projectV2StatusChangedEventFragment) Event {
	prev := f.PreviousStatus
	curr := f.Status
	summary := fmt.Sprintf("status changed: %q → %q", prev, curr)
	if title := f.Project.Title; title != "" {
		summary = fmt.Sprintf("status changed in project %q: %q → %q", title, prev, curr)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueFieldAdded(typename string, f issueFieldAddedEventFragment) Event {
	name := f.IssueField.IssueFieldCommon.Name
	summary := "added field"
	if name != "" {
		summary = fmt.Sprintf("added field %q", name)
	}
	if val := f.Value; val != "" {
		summary = fmt.Sprintf("%s: %s", summary, truncate(val))
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueFieldChanged(typename string, f issueFieldChangedEventFragment) Event {
	name := f.IssueField.IssueFieldCommon.Name
	prev := f.PreviousValue
	curr := f.NewValue
	summary := "changed field"
	if name != "" {
		summary = fmt.Sprintf("changed field %q", name)
	}
	if prev != "" || curr != "" {
		summary = fmt.Sprintf("%s: %q → %q", summary, prev, curr)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueFieldRemoved(typename string, f issueFieldRemovedEventFragment) Event {
	summary := "removed field"
	if name := f.IssueField.IssueFieldCommon.Name; name != "" {
		summary = fmt.Sprintf("removed field %q", name)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueType(typename, verb string, f issueTypeEventFragment) Event {
	summary := verb
	if name := f.IssueType.Name; name != "" {
		summary = fmt.Sprintf("%s %q", verb, name)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueTypeChanged(typename string, f issueTypeChangedEventFragment) Event {
	prev := f.PrevIssueType.Name
	curr := f.IssueType.Name
	summary := "changed issue type"
	if prev != "" || curr != "" {
		summary = fmt.Sprintf("changed issue type: %q → %q", prev, curr)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

func handleIssueCommentPin(typename, verb string, f issueCommentPinEventFragment) Event {
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   verb,
		Ref: Ref{
			NodeID:    graphqlIDString(f.ID),
			CommentID: f.IssueComment.DatabaseID,
			URL:       uriString(f.IssueComment.URL),
		},
	}
}

func handleAutomaticBaseChange(typename, verb string, f automaticBaseChangeEventFragment) Event {
	summary := verb
	if oldBase, newBase := f.OldBase, f.NewBase; oldBase != "" && newBase != "" {
		summary = fmt.Sprintf("%s: %s → %s", verb, oldBase, newBase)
	}
	return Event{
		Type:      typename,
		Actor:     f.Actor.Login,
		Timestamp: f.CreatedAt.Time,
		Summary:   summary,
		Ref:       Ref{NodeID: graphqlIDString(f.ID)},
	}
}

// --- value converters ----------------------------------------------------

// graphqlIDString defensively turns a GraphQL ID into a string. The library
// types ID as `interface{}` because IDs are opaque scalars, but in practice
// GitHub always sends them as JSON strings.
func graphqlIDString(id any) string {
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
func uriString(u URI) string {
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
func subjectInfo(s subjectFragment) (string, int) {
	if n := int(s.PullRequest.Number); n != 0 {
		return s.PullRequest.Repository.NameWithOwner, n
	}
	if n := int(s.Issue.Number); n != 0 {
		return s.Issue.Repository.NameWithOwner, n
	}
	return "", 0
}

// pickReviewerLabel produces a short label for whichever member of the
// RequestedReviewer union is populated.
func pickReviewerLabel(r userOrTeamFragment) string {
	if login := r.User.Login; login != "" {
		return login
	}
	if login := r.Bot.Login; login != "" {
		return login
	}
	if login := r.Mannequin.Login; login != "" {
		return login
	}
	if slug := r.Team.Slug; slug != "" {
		return "team:" + slug
	}
	return ""
}

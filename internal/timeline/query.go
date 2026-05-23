package timeline

import "github.com/shurcooL/githubv4"

// Repo is the minimal repo coordinate Fetch needs.
type Repo struct {
	Owner string
	Name  string
}

// pageInfo is the cursor pagination data shurcooL/githubv4 returns alongside
// each timelineItems page.
type pageInfo struct {
	EndCursor   githubv4.String
	HasNextPage githubv4.Boolean
}

// prTimelineNode is one timeline item under a PullRequest. The Typename field
// holds the GraphQL `__typename` discriminator; the matching `... on Foo`
// fragment is populated by shurcooL/githubv4 while the rest are zero values.
//

type prTimelineNode struct {
	Typename githubv4.String `graphql:"__typename"`

	// Shared with Issue timeline
	IssueComment               issueCommentFragment               `graphql:"... on IssueComment"`
	LabeledEvent               labeledEventFragment               `graphql:"... on LabeledEvent"`
	UnlabeledEvent             labeledEventFragment               `graphql:"... on UnlabeledEvent"`
	AssignedEvent              assignedEventFragment              `graphql:"... on AssignedEvent"`
	UnassignedEvent            assignedEventFragment              `graphql:"... on UnassignedEvent"`
	MilestonedEvent            milestonedEventFragment            `graphql:"... on MilestonedEvent"`
	DemilestonedEvent          milestonedEventFragment            `graphql:"... on DemilestonedEvent"`
	RenamedTitleEvent          renamedTitleEventFragment          `graphql:"... on RenamedTitleEvent"`
	ClosedEvent                closedEventFragment                `graphql:"... on ClosedEvent"`
	ReopenedEvent              reopenedEventFragment              `graphql:"... on ReopenedEvent"`
	LockedEvent                lockedEventFragment                `graphql:"... on LockedEvent"`
	UnlockedEvent              commonEvent                        `graphql:"... on UnlockedEvent"`
	PinnedEvent                commonEvent                        `graphql:"... on PinnedEvent"`
	UnpinnedEvent              commonEvent                        `graphql:"... on UnpinnedEvent"`
	SubscribedEvent            commonEvent                        `graphql:"... on SubscribedEvent"`
	UnsubscribedEvent          commonEvent                        `graphql:"... on UnsubscribedEvent"`
	MentionedEvent             commonEvent                        `graphql:"... on MentionedEvent"`
	CommentDeletedEvent        commonEvent                        `graphql:"... on CommentDeletedEvent"`
	CrossReferencedEvent       crossReferencedEventFragment       `graphql:"... on CrossReferencedEvent"`
	ReferencedEvent            referencedEventFragment            `graphql:"... on ReferencedEvent"`
	MarkedAsDuplicateEvent     markedAsDuplicateEventFragment     `graphql:"... on MarkedAsDuplicateEvent"`
	UnmarkedAsDuplicateEvent   commonEvent                        `graphql:"... on UnmarkedAsDuplicateEvent"`
	ConvertedToDiscussionEvent convertedToDiscussionEventFragment `graphql:"... on ConvertedToDiscussionEvent"`
	TransferredEvent           transferredEventFragment           `graphql:"... on TransferredEvent"`
	ConnectedEvent             connectedEventFragment             `graphql:"... on ConnectedEvent"`
	DisconnectedEvent          connectedEventFragment             `graphql:"... on DisconnectedEvent"`
	AddedToProjectEvent        projectChangeEventFragment         `graphql:"... on AddedToProjectEvent"`
	RemovedFromProjectEvent    projectChangeEventFragment         `graphql:"... on RemovedFromProjectEvent"`
	MovedColumnsInProjectEvent movedColumnsInProjectEventFragment `graphql:"... on MovedColumnsInProjectEvent"`
	ConvertedNoteToIssueEvent  projectChangeEventFragment         `graphql:"... on ConvertedNoteToIssueEvent"`
	UserBlockedEvent           userBlockedEventFragment           `graphql:"... on UserBlockedEvent"`

	// Shared with Issue — sub-issue / parent / blocking family
	SubIssueAddedEvent      subIssueEventFragment    `graphql:"... on SubIssueAddedEvent"`
	SubIssueRemovedEvent    subIssueEventFragment    `graphql:"... on SubIssueRemovedEvent"`
	ParentIssueAddedEvent   parentIssueEventFragment `graphql:"... on ParentIssueAddedEvent"`
	ParentIssueRemovedEvent parentIssueEventFragment `graphql:"... on ParentIssueRemovedEvent"`
	BlockedByAddedEvent     blockedByEventFragment   `graphql:"... on BlockedByAddedEvent"`
	BlockedByRemovedEvent   blockedByEventFragment   `graphql:"... on BlockedByRemovedEvent"`
	BlockingAddedEvent      blockingEventFragment    `graphql:"... on BlockingAddedEvent"`
	BlockingRemovedEvent    blockingEventFragment    `graphql:"... on BlockingRemovedEvent"`

	// Shared with Issue — ProjectV2 family
	AddedToProjectV2Event           projectV2ChangeEventFragment        `graphql:"... on AddedToProjectV2Event"`
	RemovedFromProjectV2Event       projectV2ChangeEventFragment        `graphql:"... on RemovedFromProjectV2Event"`
	ProjectV2ItemStatusChangedEvent projectV2StatusChangedEventFragment `graphql:"... on ProjectV2ItemStatusChangedEvent"`
	ConvertedFromDraftEvent         projectV2ChangeEventFragment        `graphql:"... on ConvertedFromDraftEvent"`

	// Shared with Issue — Issue field / type family
	IssueFieldAddedEvent   issueFieldAddedEventFragment   `graphql:"... on IssueFieldAddedEvent"`
	IssueFieldChangedEvent issueFieldChangedEventFragment `graphql:"... on IssueFieldChangedEvent"`
	IssueFieldRemovedEvent issueFieldRemovedEventFragment `graphql:"... on IssueFieldRemovedEvent"`
	IssueTypeAddedEvent    issueTypeEventFragment         `graphql:"... on IssueTypeAddedEvent"`
	IssueTypeChangedEvent  issueTypeChangedEventFragment  `graphql:"... on IssueTypeChangedEvent"`
	IssueTypeRemovedEvent  issueTypeEventFragment         `graphql:"... on IssueTypeRemovedEvent"`

	// Shared with Issue — issue comment pin
	IssueCommentPinnedEvent   issueCommentPinEventFragment `graphql:"... on IssueCommentPinnedEvent"`
	IssueCommentUnpinnedEvent issueCommentPinEventFragment `graphql:"... on IssueCommentUnpinnedEvent"`

	// PR-only
	PullRequestCommit                 pullRequestCommitFragment                 `graphql:"... on PullRequestCommit"`
	PullRequestReview                 pullRequestReviewFragment                 `graphql:"... on PullRequestReview"`
	PullRequestReviewThread           pullRequestReviewThreadFragment           `graphql:"... on PullRequestReviewThread"`
	PullRequestRevisionMarker         pullRequestRevisionMarkerFragment         `graphql:"... on PullRequestRevisionMarker"`
	PullRequestCommitCommentThread    pullRequestCommitCommentThreadFragment    `graphql:"... on PullRequestCommitCommentThread"`
	MergedEvent                       mergedEventFragment                       `graphql:"... on MergedEvent"`
	ReviewRequestedEvent              reviewRequestedEventFragment              `graphql:"... on ReviewRequestedEvent"`
	ReviewRequestRemovedEvent         reviewRequestedEventFragment              `graphql:"... on ReviewRequestRemovedEvent"`
	ReviewDismissedEvent              reviewDismissedEventFragment              `graphql:"... on ReviewDismissedEvent"`
	ReadyForReviewEvent               readyForReviewEventFragment               `graphql:"... on ReadyForReviewEvent"`
	ConvertToDraftEvent               convertToDraftEventFragment               `graphql:"... on ConvertToDraftEvent"`
	HeadRefForcePushedEvent           forcePushedEventFragment                  `graphql:"... on HeadRefForcePushedEvent"`
	BaseRefForcePushedEvent           forcePushedEventFragment                  `graphql:"... on BaseRefForcePushedEvent"`
	BaseRefChangedEvent               baseRefChangedEventFragment               `graphql:"... on BaseRefChangedEvent"`
	BaseRefDeletedEvent               baseRefDeletedEventFragment               `graphql:"... on BaseRefDeletedEvent"`
	HeadRefDeletedEvent               headRefDeletedEventFragment               `graphql:"... on HeadRefDeletedEvent"`
	HeadRefRestoredEvent              headRefRestoredEventFragment              `graphql:"... on HeadRefRestoredEvent"`
	DeployedEvent                     deployedEventFragment                     `graphql:"... on DeployedEvent"`
	DeploymentEnvironmentChangedEvent deploymentEnvironmentChangedEventFragment `graphql:"... on DeploymentEnvironmentChangedEvent"`
	AutoMergeEnabledEvent             autoChangeEventFragment                   `graphql:"... on AutoMergeEnabledEvent"`
	AutoMergeDisabledEvent            autoChangeEventFragment                   `graphql:"... on AutoMergeDisabledEvent"`
	AutoRebaseEnabledEvent            autoChangeEventFragment                   `graphql:"... on AutoRebaseEnabledEvent"`
	AutoSquashEnabledEvent            autoChangeEventFragment                   `graphql:"... on AutoSquashEnabledEvent"`
	AutomaticBaseChangeSucceededEvent automaticBaseChangeEventFragment          `graphql:"... on AutomaticBaseChangeSucceededEvent"`
	AutomaticBaseChangeFailedEvent    automaticBaseChangeEventFragment          `graphql:"... on AutomaticBaseChangeFailedEvent"`
	AddedToMergeQueueEvent            mergeQueueEventFragment                   `graphql:"... on AddedToMergeQueueEvent"`
	RemovedFromMergeQueueEvent        mergeQueueEventFragment                   `graphql:"... on RemovedFromMergeQueueEvent"`
}

// issueTimelineNode is the IssueTimelineItems variant — a strict subset of the
// PR union without PR-specific items.
//

type issueTimelineNode struct {
	Typename githubv4.String `graphql:"__typename"`

	IssueComment               issueCommentFragment               `graphql:"... on IssueComment"`
	LabeledEvent               labeledEventFragment               `graphql:"... on LabeledEvent"`
	UnlabeledEvent             labeledEventFragment               `graphql:"... on UnlabeledEvent"`
	AssignedEvent              assignedEventFragment              `graphql:"... on AssignedEvent"`
	UnassignedEvent            assignedEventFragment              `graphql:"... on UnassignedEvent"`
	MilestonedEvent            milestonedEventFragment            `graphql:"... on MilestonedEvent"`
	DemilestonedEvent          milestonedEventFragment            `graphql:"... on DemilestonedEvent"`
	RenamedTitleEvent          renamedTitleEventFragment          `graphql:"... on RenamedTitleEvent"`
	ClosedEvent                closedEventFragment                `graphql:"... on ClosedEvent"`
	ReopenedEvent              reopenedEventFragment              `graphql:"... on ReopenedEvent"`
	LockedEvent                lockedEventFragment                `graphql:"... on LockedEvent"`
	UnlockedEvent              commonEvent                        `graphql:"... on UnlockedEvent"`
	PinnedEvent                commonEvent                        `graphql:"... on PinnedEvent"`
	UnpinnedEvent              commonEvent                        `graphql:"... on UnpinnedEvent"`
	SubscribedEvent            commonEvent                        `graphql:"... on SubscribedEvent"`
	UnsubscribedEvent          commonEvent                        `graphql:"... on UnsubscribedEvent"`
	MentionedEvent             commonEvent                        `graphql:"... on MentionedEvent"`
	CommentDeletedEvent        commonEvent                        `graphql:"... on CommentDeletedEvent"`
	CrossReferencedEvent       crossReferencedEventFragment       `graphql:"... on CrossReferencedEvent"`
	ReferencedEvent            referencedEventFragment            `graphql:"... on ReferencedEvent"`
	MarkedAsDuplicateEvent     markedAsDuplicateEventFragment     `graphql:"... on MarkedAsDuplicateEvent"`
	UnmarkedAsDuplicateEvent   commonEvent                        `graphql:"... on UnmarkedAsDuplicateEvent"`
	ConvertedToDiscussionEvent convertedToDiscussionEventFragment `graphql:"... on ConvertedToDiscussionEvent"`
	TransferredEvent           transferredEventFragment           `graphql:"... on TransferredEvent"`
	ConnectedEvent             connectedEventFragment             `graphql:"... on ConnectedEvent"`
	DisconnectedEvent          connectedEventFragment             `graphql:"... on DisconnectedEvent"`
	AddedToProjectEvent        projectChangeEventFragment         `graphql:"... on AddedToProjectEvent"`
	RemovedFromProjectEvent    projectChangeEventFragment         `graphql:"... on RemovedFromProjectEvent"`
	MovedColumnsInProjectEvent movedColumnsInProjectEventFragment `graphql:"... on MovedColumnsInProjectEvent"`
	ConvertedNoteToIssueEvent  projectChangeEventFragment         `graphql:"... on ConvertedNoteToIssueEvent"`
	UserBlockedEvent           userBlockedEventFragment           `graphql:"... on UserBlockedEvent"`

	// Sub-issue / parent / blocking family
	SubIssueAddedEvent      subIssueEventFragment    `graphql:"... on SubIssueAddedEvent"`
	SubIssueRemovedEvent    subIssueEventFragment    `graphql:"... on SubIssueRemovedEvent"`
	ParentIssueAddedEvent   parentIssueEventFragment `graphql:"... on ParentIssueAddedEvent"`
	ParentIssueRemovedEvent parentIssueEventFragment `graphql:"... on ParentIssueRemovedEvent"`
	BlockedByAddedEvent     blockedByEventFragment   `graphql:"... on BlockedByAddedEvent"`
	BlockedByRemovedEvent   blockedByEventFragment   `graphql:"... on BlockedByRemovedEvent"`
	BlockingAddedEvent      blockingEventFragment    `graphql:"... on BlockingAddedEvent"`
	BlockingRemovedEvent    blockingEventFragment    `graphql:"... on BlockingRemovedEvent"`

	// ProjectV2 family
	AddedToProjectV2Event           projectV2ChangeEventFragment        `graphql:"... on AddedToProjectV2Event"`
	RemovedFromProjectV2Event       projectV2ChangeEventFragment        `graphql:"... on RemovedFromProjectV2Event"`
	ProjectV2ItemStatusChangedEvent projectV2StatusChangedEventFragment `graphql:"... on ProjectV2ItemStatusChangedEvent"`
	ConvertedFromDraftEvent         projectV2ChangeEventFragment        `graphql:"... on ConvertedFromDraftEvent"`

	// Issue field / type family
	IssueFieldAddedEvent   issueFieldAddedEventFragment   `graphql:"... on IssueFieldAddedEvent"`
	IssueFieldChangedEvent issueFieldChangedEventFragment `graphql:"... on IssueFieldChangedEvent"`
	IssueFieldRemovedEvent issueFieldRemovedEventFragment `graphql:"... on IssueFieldRemovedEvent"`
	IssueTypeAddedEvent    issueTypeEventFragment         `graphql:"... on IssueTypeAddedEvent"`
	IssueTypeChangedEvent  issueTypeChangedEventFragment  `graphql:"... on IssueTypeChangedEvent"`
	IssueTypeRemovedEvent  issueTypeEventFragment         `graphql:"... on IssueTypeRemovedEvent"`

	// Issue comment pin
	IssueCommentPinnedEvent   issueCommentPinEventFragment `graphql:"... on IssueCommentPinnedEvent"`
	IssueCommentUnpinnedEvent issueCommentPinEventFragment `graphql:"... on IssueCommentUnpinnedEvent"`
}

// timelineQuery is the top-level query shape passed to githubv4.Client.Query.
// Only one of the two `... on` branches under IssueOrPullRequest is populated
// per response; the other stays zero-valued.
type timelineQuery struct {
	Repository struct {
		IssueOrPullRequest struct {
			Typename    githubv4.String `graphql:"__typename"`
			PullRequest struct {
				TimelineItems struct {
					Nodes    []prTimelineNode
					PageInfo pageInfo
				} `graphql:"timelineItems(first: 100, after: $cursor)"`
			} `graphql:"... on PullRequest"`
			Issue struct {
				TimelineItems struct {
					Nodes    []issueTimelineNode
					PageInfo pageInfo
				} `graphql:"timelineItems(first: 100, after: $cursor)"`
			} `graphql:"... on Issue"`
		} `graphql:"issueOrPullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

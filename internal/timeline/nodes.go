package timeline

import "github.com/shurcooL/githubv4"

// Each struct below mirrors one member of the GraphQL
// PullRequestTimelineItems / IssueTimelineItems union. They are embedded into
// [prTimelineNode] / [issueTimelineNode] via `graphql:"... on TypeName"` tags
// so that the shurcooL/githubv4 library generates the inline fragments
// automatically.

type actorFragment struct {
	Login githubv4.String
}

type labelFragment struct {
	Name githubv4.String
}

type userOrTeamFragment struct {
	User struct {
		Login githubv4.String
	} `graphql:"... on User"`
	Bot struct {
		Login githubv4.String
	} `graphql:"... on Bot"`
	Mannequin struct {
		Login githubv4.String
	} `graphql:"... on Mannequin"`
	Team struct {
		Slug githubv4.String
	} `graphql:"... on Team"`
}

// assigneeFragment mirrors GitHub's Assignee union
// (User | Bot | Mannequin | Organization). GraphQL forbids bare selections on
// a union, so each member is spelled out via `... on T`.
type assigneeFragment struct {
	User struct {
		Login githubv4.String
	} `graphql:"... on User"`
	Bot struct {
		Login githubv4.String
	} `graphql:"... on Bot"`
	Mannequin struct {
		Login githubv4.String
	} `graphql:"... on Mannequin"`
	Organization struct {
		Login githubv4.String
	} `graphql:"... on Organization"`
}

// pick returns the first non-empty login across the union's members.
func (a assigneeFragment) pick() string {
	for _, login := range []githubv4.String{a.User.Login, a.Bot.Login, a.Mannequin.Login, a.Organization.Login} {
		if s := string(login); s != "" {
			return s
		}
	}
	return ""
}

type subjectFragment struct {
	Issue struct {
		Number     githubv4.Int
		Title      githubv4.String
		Repository struct {
			NameWithOwner githubv4.String
		}
	} `graphql:"... on Issue"`
	PullRequest struct {
		Number     githubv4.Int
		Title      githubv4.String
		Repository struct {
			NameWithOwner githubv4.String
		}
	} `graphql:"... on PullRequest"`
}

type commitFragment struct {
	OID             githubv4.GitObjectID
	MessageHeadline githubv4.String
	CommittedDate   githubv4.DateTime
	Author          struct {
		Name githubv4.String
		User actorFragment
	}
}

// commonEvent groups the {id, actor, createdAt} selection used by almost every
// non-commit timeline event.
type commonEvent struct {
	ID        githubv4.ID
	Actor     actorFragment
	CreatedAt githubv4.DateTime
}

// --- Fragments shared by both PR and Issue timelines -----------------------

type issueCommentFragment struct {
	ID         githubv4.ID
	DatabaseID int64 `graphql:"databaseId"`
	Author     actorFragment
	Body       githubv4.String
	CreatedAt  githubv4.DateTime
	URL        githubv4.URI
}

type labeledEventFragment struct {
	commonEvent

	Label labelFragment
}

type assignedEventFragment struct {
	commonEvent

	Assignee assigneeFragment
}

type milestonedEventFragment struct {
	commonEvent

	MilestoneTitle githubv4.String
}

type renamedTitleEventFragment struct {
	commonEvent

	PreviousTitle githubv4.String
	CurrentTitle  githubv4.String
}

type closedEventFragment struct {
	commonEvent

	StateReason githubv4.String
}

type reopenedEventFragment struct {
	commonEvent
}

type lockedEventFragment struct {
	commonEvent

	LockReason githubv4.String
}

type crossReferencedEventFragment struct {
	commonEvent

	Source subjectFragment
}

type referencedEventFragment struct {
	commonEvent

	Subject subjectFragment
	Commit  struct {
		OID githubv4.GitObjectID
	}
}

type markedAsDuplicateEventFragment struct {
	commonEvent

	Canonical subjectFragment
}

type convertedToDiscussionEventFragment struct {
	commonEvent

	Discussion struct {
		Number githubv4.Int
	}
}

type transferredEventFragment struct {
	commonEvent

	FromRepository struct {
		NameWithOwner githubv4.String
	}
}

type connectedEventFragment struct {
	commonEvent

	Subject subjectFragment
}

type projectChangeEventFragment struct {
	commonEvent

	Project struct {
		Name githubv4.String
	}
	ProjectColumnName githubv4.String
}

type movedColumnsInProjectEventFragment struct {
	commonEvent

	PreviousProjectColumnName githubv4.String
	ProjectColumnName         githubv4.String
}

type userBlockedEventFragment struct {
	commonEvent

	Subject struct {
		Login githubv4.String
	}
	BlockDuration githubv4.String
}

// issueRefFragment captures the {repo, number, title} triple used by the
// sub-issue / parent / blocking event family. The structure mirrors what
// subjectFragment exposes for the Issue branch, but these events reference
// the related issue directly without going through a union.
type issueRefFragment struct {
	Number     githubv4.Int
	Title      githubv4.String
	Repository struct {
		NameWithOwner githubv4.String
	}
}

type subIssueEventFragment struct {
	commonEvent

	SubIssue issueRefFragment
}

type parentIssueEventFragment struct {
	commonEvent

	Parent issueRefFragment
}

type blockedByEventFragment struct {
	commonEvent

	BlockingIssue issueRefFragment
}

type blockingEventFragment struct {
	commonEvent

	BlockedIssue issueRefFragment
}

// projectV2ChangeEventFragment is shared by AddedToProjectV2Event,
// RemovedFromProjectV2Event, and ConvertedFromDraftEvent — all three expose
// the same {project} shape.
type projectV2ChangeEventFragment struct {
	commonEvent

	Project struct {
		Number githubv4.Int
		Title  githubv4.String
	}
}

type projectV2StatusChangedEventFragment struct {
	commonEvent

	Project struct {
		Number githubv4.Int
		Title  githubv4.String
	}
	PreviousStatus githubv4.String
	Status         githubv4.String
}

// issueFieldUnionFragment selects the common `name` from the IssueFields
// union (IssueFieldDate | IssueFieldNumber | IssueFieldSingleSelect |
// IssueFieldText). All four implement the IssueFieldCommon interface, so a
// single `... on IssueFieldCommon` spread covers them.
type issueFieldUnionFragment struct {
	IssueFieldCommon struct {
		Name githubv4.String
	} `graphql:"... on IssueFieldCommon"`
}

type issueFieldAddedEventFragment struct {
	commonEvent

	IssueField issueFieldUnionFragment
	Value      githubv4.String
}

type issueFieldChangedEventFragment struct {
	commonEvent

	IssueField    issueFieldUnionFragment
	PreviousValue githubv4.String
	NewValue      githubv4.String
}

type issueFieldRemovedEventFragment struct {
	commonEvent

	IssueField issueFieldUnionFragment
}

type issueTypeEventFragment struct {
	commonEvent

	IssueType struct {
		Name githubv4.String
	}
}

type issueTypeChangedEventFragment struct {
	commonEvent

	IssueType struct {
		Name githubv4.String
	}
	PrevIssueType struct {
		Name githubv4.String
	}
}

// issueCommentPinEventFragment carries the pinned/unpinned IssueComment's
// identifiers so downstream tooling can fetch the comment body via `gh api`.
type issueCommentPinEventFragment struct {
	commonEvent

	IssueComment struct {
		ID         githubv4.ID
		DatabaseID int64 `graphql:"databaseId"`
		URL        githubv4.URI
	}
}

// --- PR-only fragments -----------------------------------------------------

type pullRequestCommitFragment struct {
	ID     githubv4.ID
	Commit commitFragment
	URL    githubv4.URI
}

type pullRequestReviewFragment struct {
	ID          githubv4.ID
	DatabaseID  int64 `graphql:"databaseId"`
	Author      actorFragment
	State       githubv4.PullRequestReviewState
	Body        githubv4.String
	SubmittedAt githubv4.DateTime
	CreatedAt   githubv4.DateTime
	URL         githubv4.URI
}

type pullRequestReviewThreadFragment struct {
	ID githubv4.ID
}

type pullRequestRevisionMarkerFragment struct {
	CreatedAt      githubv4.DateTime
	LastSeenCommit struct {
		OID githubv4.GitObjectID
	}
}

type pullRequestCommitCommentThreadFragment struct {
	ID     githubv4.ID
	Commit struct {
		OID githubv4.GitObjectID
	}
}

type mergedEventFragment struct {
	commonEvent

	URL    githubv4.URI
	Commit struct {
		OID githubv4.GitObjectID
	}
	MergeRefName githubv4.String
}

type reviewRequestedEventFragment struct {
	commonEvent

	RequestedReviewer userOrTeamFragment
}

type reviewDismissedEventFragment struct {
	commonEvent

	DismissalMessage githubv4.String
	Review           struct {
		Author actorFragment
	}
}

type readyForReviewEventFragment struct {
	commonEvent
}

type convertToDraftEventFragment struct {
	commonEvent
}

type forcePushedEventFragment struct {
	commonEvent

	BeforeCommit struct {
		OID githubv4.GitObjectID
	}
	AfterCommit struct {
		OID githubv4.GitObjectID
	}
}

type baseRefChangedEventFragment struct {
	commonEvent

	PreviousRefName githubv4.String
	CurrentRefName  githubv4.String
}

type baseRefDeletedEventFragment struct {
	commonEvent

	BaseRefName githubv4.String
}

type headRefDeletedEventFragment struct {
	commonEvent

	HeadRefName githubv4.String
}

type headRefRestoredEventFragment struct {
	commonEvent
}

type deployedEventFragment struct {
	commonEvent

	Deployment struct {
		Environment githubv4.String
	}
}

type deploymentEnvironmentChangedEventFragment struct {
	commonEvent

	DeploymentStatus struct {
		Deployment struct {
			Environment githubv4.String
		}
	}
}

type autoChangeEventFragment struct {
	commonEvent
}

type automaticBaseChangeEventFragment struct {
	commonEvent

	OldBase githubv4.String
	NewBase githubv4.String
}

type mergeQueueEventFragment struct {
	commonEvent
}

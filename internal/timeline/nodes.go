package timeline

// Each struct below mirrors one member of the GraphQL
// PullRequestTimelineItems / IssueTimelineItems union. They are embedded into
// [prTimelineNode] / [issueTimelineNode] via `graphql:"... on TypeName"` tags
// so that the go-gh GraphQL client generates the inline fragments
// automatically.

type actorFragment struct {
	Login string
}

type labelFragment struct {
	Name string
}

type userOrTeamFragment struct {
	User struct {
		Login string
	} `graphql:"... on User"`
	Bot struct {
		Login string
	} `graphql:"... on Bot"`
	Mannequin struct {
		Login string
	} `graphql:"... on Mannequin"`
	Team struct {
		Slug string
	} `graphql:"... on Team"`
}

// assigneeFragment mirrors GitHub's Assignee union
// (User | Bot | Mannequin | Organization). GraphQL forbids bare selections on
// a union, so each member is spelled out via `... on T`.
type assigneeFragment struct {
	User struct {
		Login string
	} `graphql:"... on User"`
	Bot struct {
		Login string
	} `graphql:"... on Bot"`
	Mannequin struct {
		Login string
	} `graphql:"... on Mannequin"`
	Organization struct {
		Login string
	} `graphql:"... on Organization"`
}

// pick returns the first non-empty login across the union's members.
func (a assigneeFragment) pick() string {
	for _, login := range []string{a.User.Login, a.Bot.Login, a.Mannequin.Login, a.Organization.Login} {
		if login != "" {
			return login
		}
	}
	return ""
}

type subjectFragment struct {
	Issue struct {
		Number     int32
		Title      string
		Repository struct {
			NameWithOwner string
		}
	} `graphql:"... on Issue"`
	PullRequest struct {
		Number     int32
		Title      string
		Repository struct {
			NameWithOwner string
		}
	} `graphql:"... on PullRequest"`
}

type commitFragment struct {
	OID             string
	MessageHeadline string
	CommittedDate   DateTime
	Author          struct {
		Name string
		User actorFragment
	}
}

// commonEvent groups the {id, actor, createdAt} selection used by almost every
// non-commit timeline event.
type commonEvent struct {
	ID        any
	Actor     actorFragment
	CreatedAt DateTime
}

// --- Fragments shared by both PR and Issue timelines -----------------------

type issueCommentFragment struct {
	ID         any
	DatabaseID int64 `graphql:"databaseId"`
	Author     actorFragment
	Body       string
	CreatedAt  DateTime
	URL        URI
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

	MilestoneTitle string
}

type renamedTitleEventFragment struct {
	commonEvent

	PreviousTitle string
	CurrentTitle  string
}

type closedEventFragment struct {
	commonEvent

	StateReason string
}

type reopenedEventFragment struct {
	commonEvent
}

type lockedEventFragment struct {
	commonEvent

	LockReason string
}

type crossReferencedEventFragment struct {
	commonEvent

	Source subjectFragment
}

type referencedEventFragment struct {
	commonEvent

	Subject subjectFragment
	Commit  struct {
		OID string
	}
}

type markedAsDuplicateEventFragment struct {
	commonEvent

	Canonical subjectFragment
}

type convertedToDiscussionEventFragment struct {
	commonEvent

	Discussion struct {
		Number int32
	}
}

type transferredEventFragment struct {
	commonEvent

	FromRepository struct {
		NameWithOwner string
	}
}

type connectedEventFragment struct {
	commonEvent

	Subject subjectFragment
}

type projectChangeEventFragment struct {
	commonEvent

	Project struct {
		Name string
	}
	ProjectColumnName string
}

type movedColumnsInProjectEventFragment struct {
	commonEvent

	PreviousProjectColumnName string
	ProjectColumnName         string
}

type userBlockedEventFragment struct {
	commonEvent

	Subject struct {
		Login string
	}
	BlockDuration string
}

// issueRefFragment captures the {repo, number, title} triple used by the
// sub-issue / parent / blocking event family. The structure mirrors what
// subjectFragment exposes for the Issue branch, but these events reference
// the related issue directly without going through a union.
type issueRefFragment struct {
	Number     int32
	Title      string
	Repository struct {
		NameWithOwner string
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
		Number int32
		Title  string
	}
}

type projectV2StatusChangedEventFragment struct {
	commonEvent

	Project struct {
		Number int32
		Title  string
	}
	PreviousStatus string
	Status         string
}

// issueFieldUnionFragment selects the common `name` from the IssueFields
// union (IssueFieldDate | IssueFieldNumber | IssueFieldSingleSelect |
// IssueFieldText). All four implement the IssueFieldCommon interface, so a
// single `... on IssueFieldCommon` spread covers them.
type issueFieldUnionFragment struct {
	IssueFieldCommon struct {
		Name string
	} `graphql:"... on IssueFieldCommon"`
}

type issueFieldAddedEventFragment struct {
	commonEvent

	IssueField issueFieldUnionFragment
	Value      string
}

type issueFieldChangedEventFragment struct {
	commonEvent

	IssueField    issueFieldUnionFragment
	PreviousValue string
	NewValue      string
}

type issueFieldRemovedEventFragment struct {
	commonEvent

	IssueField issueFieldUnionFragment
}

type issueTypeEventFragment struct {
	commonEvent

	IssueType struct {
		Name string
	}
}

type issueTypeChangedEventFragment struct {
	commonEvent

	IssueType struct {
		Name string
	}
	PrevIssueType struct {
		Name string
	}
}

// issueCommentPinEventFragment carries the pinned/unpinned IssueComment's
// identifiers so downstream tooling can fetch the comment body via `gh api`.
type issueCommentPinEventFragment struct {
	commonEvent

	IssueComment struct {
		ID         any
		DatabaseID int64 `graphql:"databaseId"`
		URL        URI
	}
}

// --- PR-only fragments -----------------------------------------------------

type pullRequestCommitFragment struct {
	ID     any
	Commit commitFragment
	URL    URI
}

type pullRequestReviewFragment struct {
	ID          any
	DatabaseID  int64 `graphql:"databaseId"`
	Author      actorFragment
	State       string
	Body        string
	SubmittedAt DateTime
	CreatedAt   DateTime
	URL         URI
}

type pullRequestReviewThreadFragment struct {
	ID any
}

type pullRequestRevisionMarkerFragment struct {
	CreatedAt      DateTime
	LastSeenCommit struct {
		OID string
	}
}

type pullRequestCommitCommentThreadFragment struct {
	ID     any
	Commit struct {
		OID string
	}
}

type mergedEventFragment struct {
	commonEvent

	URL    URI
	Commit struct {
		OID string
	}
	MergeRefName string
}

type reviewRequestedEventFragment struct {
	commonEvent

	RequestedReviewer userOrTeamFragment
}

type reviewDismissedEventFragment struct {
	commonEvent

	DismissalMessage string
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
		OID string
	}
	AfterCommit struct {
		OID string
	}
}

type baseRefChangedEventFragment struct {
	commonEvent

	PreviousRefName string
	CurrentRefName  string
}

type baseRefDeletedEventFragment struct {
	commonEvent

	BaseRefName string
}

type headRefDeletedEventFragment struct {
	commonEvent

	HeadRefName string
}

type headRefRestoredEventFragment struct {
	commonEvent
}

type deployedEventFragment struct {
	commonEvent

	Deployment struct {
		Environment string
	}
}

type deploymentEnvironmentChangedEventFragment struct {
	commonEvent

	DeploymentStatus struct {
		Deployment struct {
			Environment string
		}
	}
}

type autoChangeEventFragment struct {
	commonEvent
}

type automaticBaseChangeEventFragment struct {
	commonEvent

	OldBase string
	NewBase string
}

type mergeQueueEventFragment struct {
	commonEvent
}

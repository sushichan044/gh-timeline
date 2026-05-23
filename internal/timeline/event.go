// Package timeline fetches and renders GitHub PR timeline events.
package timeline

import "time"

// Type constants for the timeline events surfaced explicitly by this tool.
// Other event types from the GitHub API are still rendered with their raw
// `event` string — these constants only cover the cases with custom logic.
const (
	TypeCommitted      = "committed"
	TypeReviewed       = "reviewed"
	TypeCommented      = "commented"
	TypeLabeled        = "labeled"
	TypeUnlabeled      = "unlabeled"
	TypeAssigned       = "assigned"
	TypeUnassigned     = "unassigned"
	TypeReviewReq      = "review_requested"
	TypeHeadForcePush  = "head_ref_force_pushed"
	TypeReadyForReview = "ready_for_review"
	TypeMerged         = "merged"
)

// Event is a normalized timeline entry suitable for both human-readable and
// machine-readable output.
type Event struct {
	Type      string    `json:"type"`
	Actor     string    `json:"actor,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Summary   string    `json:"summary"`
	Ref       Ref       `json:"ref,omitzero"`
}

// Ref holds identifiers an AI agent can pass to `gh api` to fetch the full
// payload for an event. Empty fields are omitted from JSON.
type Ref struct {
	NodeID    string `json:"node_id,omitempty"`
	SHA       string `json:"sha,omitempty"`
	ReviewID  int64  `json:"review_id,omitempty"`
	CommentID int64  `json:"comment_id,omitempty"`
	URL       string `json:"url,omitempty"`
}

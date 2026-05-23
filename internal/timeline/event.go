// Package timeline fetches and renders GitHub Issue / PR timeline events via
// the GraphQL API.
package timeline

import "time"

// Event is a normalized timeline entry suitable for both human-readable and
// machine-readable output. Type is the GraphQL __typename of the source node
// (PascalCase, e.g. "PullRequestCommit", "LabeledEvent").
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

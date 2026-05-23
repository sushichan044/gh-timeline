package timeline

import "time"

// rawEvent mirrors the union of fields the GitHub issues/timeline endpoint
// returns across all event variants. Unused fields are omitted.
//
// API docs: https://docs.github.com/en/rest/issues/timeline
type rawEvent struct {
	// Common fields
	Event     string     `json:"event"`
	NodeID    string     `json:"node_id"`
	URL       string     `json:"url"`
	CreatedAt *time.Time `json:"created_at"`
	Actor     *rawUser   `json:"actor"`

	// Review (`reviewed`)
	ID          int64      `json:"id"`
	State       string     `json:"state"`
	SubmittedAt *time.Time `json:"submitted_at"`
	User        *rawUser   `json:"user"`

	// Comment (`commented`)
	Body string `json:"body"`

	// Commit (`committed`)
	SHA       string     `json:"sha"`
	Message   string     `json:"message"`
	Author    *rawAuthor `json:"author"`
	Committer *rawAuthor `json:"committer"`

	// Labels (`labeled` / `unlabeled`)
	Label *rawLabel `json:"label"`

	// Assignees (`assigned` / `unassigned`)
	Assignee *rawUser `json:"assignee"`

	// Review request (`review_requested` / `review_request_removed`)
	RequestedReviewer *rawUser `json:"requested_reviewer"`
	RequestedTeam     *rawTeam `json:"requested_team"`
}

type rawUser struct {
	Login string `json:"login"`
}

type rawAuthor struct {
	Name  string     `json:"name"`
	Email string     `json:"email"`
	Date  *time.Time `json:"date"`
}

type rawLabel struct {
	Name string `json:"name"`
}

type rawTeam struct {
	Slug string `json:"slug"`
}

// normalize converts an API event into the public Event shape.
func (r rawEvent) normalize() Event {
	return Event{
		Type:      r.Event,
		Actor:     r.actorLogin(),
		Timestamp: r.timestamp(),
		Summary:   r.summary(),
		Ref:       r.ref(),
	}
}

// actorLogin picks the most relevant login for each event variant.
// For `committed` events the API returns no `actor` block — the commit
// author/committer is the closest stand-in.
func (r rawEvent) actorLogin() string {
	if r.Actor != nil && r.Actor.Login != "" {
		return r.Actor.Login
	}
	if r.User != nil && r.User.Login != "" {
		return r.User.Login
	}
	if r.Event == TypeCommitted && r.Author != nil {
		return r.Author.Name
	}
	return ""
}

// timestamp resolves the event time across variants.
//
//	committed → committer.date (fall back to author.date)
//	reviewed  → submitted_at
//	others    → created_at
func (r rawEvent) timestamp() time.Time {
	switch r.Event {
	case TypeCommitted:
		if r.Committer != nil && r.Committer.Date != nil {
			return *r.Committer.Date
		}
		if r.Author != nil && r.Author.Date != nil {
			return *r.Author.Date
		}
	case TypeReviewed:
		if r.SubmittedAt != nil {
			return *r.SubmittedAt
		}
	}
	if r.CreatedAt != nil {
		return *r.CreatedAt
	}
	return time.Time{}
}

// ref captures identifiers an agent can pass to `gh api`.
func (r rawEvent) ref() Ref {
	ref := Ref{NodeID: r.NodeID, URL: r.URL}
	switch r.Event {
	case TypeCommitted:
		ref.SHA = r.SHA
	case TypeReviewed:
		ref.ReviewID = r.ID
	case TypeCommented:
		ref.CommentID = r.ID
	}
	return ref
}

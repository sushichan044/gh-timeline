package timeline_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/sushichan044/gh-timeline/internal/timeline"
)

// fakeClient drives Fetch's pagination loop deterministically without an
// httptest server.
type fakeClient struct {
	t        *testing.T
	requests []string
	pages    [][]byte
	links    []string
	err      error
}

func (f *fakeClient) Request(_, path string, _ io.Reader) (*http.Response, error) {
	f.requests = append(f.requests, path)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.pages) == 0 {
		f.t.Fatalf("unexpected extra request to %q", path)
	}
	body := f.pages[0]
	link := f.links[0]
	f.pages = f.pages[1:]
	f.links = f.links[1:]
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     http.Header{},
	}
	if link != "" {
		resp.Header.Set("Link", link)
	}
	return resp, nil
}

func TestFetch_paginatesAndSortsChronologically(t *testing.T) {
	t.Parallel()
	page1 := `[{"event":"labeled","created_at":"2026-01-02T10:00:00Z","actor":{"login":"alice"},"label":{"name":"bug"}}]`
	page2 := `[{"event":"reviewed","submitted_at":"2026-01-01T09:00:00Z","user":{"login":"bob"},"state":"approved","id":42}]`

	nextURL := "https://api.github.com/repositories/1/issues/123/timeline?page=2"
	client := &fakeClient{
		t:     t,
		pages: [][]byte{[]byte(page1), []byte(page2)},
		links: []string{`<` + nextURL + `>; rel="next", <https://api.github.com/...>; rel="last"`, ""},
	}

	events, err := timeline.Fetch(client, timeline.Repo{Owner: "owner", Name: "repo"}, 123)
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Type != "reviewed" || events[1].Type != "labeled" {
		t.Errorf("events not sorted chronologically: %+v", events)
	}
	want := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	if !events[0].Timestamp.Equal(want) {
		t.Errorf("first event timestamp = %v, want %v", events[0].Timestamp, want)
	}

	const wantFirstPath = "repos/owner/repo/issues/123/timeline?per_page=100"
	if got := firstRequest(t, client.requests); got != wantFirstPath {
		t.Errorf("first request = %q, want per_page=100", got)
	}
	if len(client.requests) != 2 || client.requests[1] != nextURL {
		t.Errorf("did not follow Link rel=\"next\": got %v", client.requests)
	}
}

func TestFetch_returnsFriendlyErrorOn404(t *testing.T) {
	t.Parallel()
	client := &fakeClient{t: t, err: &api.HTTPError{StatusCode: http.StatusNotFound, RequestURL: &url.URL{Path: "/x"}}}

	_, err := timeline.Fetch(client, timeline.Repo{Owner: "owner", Name: "repo"}, 999)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestFetch_propagatesNon404Errors(t *testing.T) {
	t.Parallel()
	client := &fakeClient{t: t, err: errors.New("boom")}
	_, err := timeline.Fetch(client, timeline.Repo{Owner: "o", Name: "r"}, 1)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestFetch_rejectsInvalidInput(t *testing.T) {
	t.Parallel()
	client := &fakeClient{t: t}
	if _, err := timeline.Fetch(client, timeline.Repo{}, 1); err == nil {
		t.Error("expected error for empty repo")
	}
	if _, err := timeline.Fetch(client, timeline.Repo{Owner: "o", Name: "r"}, 0); err == nil {
		t.Error("expected error for zero PR number")
	}
}

// TestFetch_normalizesEachEventType drives Fetch with a synthetic JSON payload
// per event type and asserts the public-facing Event shape — what humans see
// in text mode and what AI agents parse from JSON.
func TestFetch_normalizesEachEventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		json        string
		wantType    string
		wantActor   string
		wantSummary string
		checkRef    func(t *testing.T, r timeline.Ref)
	}{
		{
			name:        "committed exposes first line of commit message and SHA",
			json:        `[{"event":"committed","sha":"abc123","message":"feat: add a thing\n\nlong body","committer":{"date":"2026-01-01T10:00:00Z","name":"Alice"},"author":{"name":"Alice"}}]`,
			wantType:    "committed",
			wantActor:   "Alice",
			wantSummary: "feat: add a thing",
			checkRef: func(t *testing.T, r timeline.Ref) {
				t.Helper()
				if r.SHA != "abc123" {
					t.Errorf("ref.SHA = %q, want abc123", r.SHA)
				}
			},
		},
		{
			name:        "reviewed exposes state and review ID",
			json:        `[{"event":"reviewed","submitted_at":"2026-01-01T10:00:00Z","user":{"login":"bob"},"state":"approved","id":42}]`,
			wantType:    "reviewed",
			wantActor:   "bob",
			wantSummary: "approved",
			checkRef: func(t *testing.T, r timeline.Ref) {
				t.Helper()
				if r.ReviewID != 42 {
					t.Errorf("ref.ReviewID = %d, want 42", r.ReviewID)
				}
			},
		},
		{
			name:        "commented truncates body and exposes comment ID",
			json:        `[{"event":"commented","created_at":"2026-01-01T10:00:00Z","actor":{"login":"carol"},"body":"looks good!\nnevermind","id":99}]`,
			wantType:    "commented",
			wantActor:   "carol",
			wantSummary: "looks good!",
			checkRef: func(t *testing.T, r timeline.Ref) {
				t.Helper()
				if r.CommentID != 99 {
					t.Errorf("ref.CommentID = %d, want 99", r.CommentID)
				}
			},
		},
		{
			name:        "head_ref_force_pushed surfaces fixed label",
			json:        `[{"event":"head_ref_force_pushed","created_at":"2026-01-01T10:00:00Z","actor":{"login":"dave"}}]`,
			wantType:    "head_ref_force_pushed",
			wantActor:   "dave",
			wantSummary: "force-pushed",
		},
		{
			name:        "ready_for_review surfaces fixed label",
			json:        `[{"event":"ready_for_review","created_at":"2026-01-01T10:00:00Z","actor":{"login":"eve"}}]`,
			wantType:    "ready_for_review",
			wantActor:   "eve",
			wantSummary: "marked ready for review",
		},
		{
			name:        "labeled surfaces the label name",
			json:        `[{"event":"labeled","created_at":"2026-01-01T10:00:00Z","actor":{"login":"frank"},"label":{"name":"bug"}}]`,
			wantType:    "labeled",
			wantActor:   "frank",
			wantSummary: "bug",
		},
		{
			name:        "assigned surfaces the assignee login",
			json:        `[{"event":"assigned","created_at":"2026-01-01T10:00:00Z","actor":{"login":"grace"},"assignee":{"login":"heidi"}}]`,
			wantType:    "assigned",
			wantActor:   "grace",
			wantSummary: "heidi",
		},
		{
			name:        "review_requested surfaces reviewer login",
			json:        `[{"event":"review_requested","created_at":"2026-01-01T10:00:00Z","actor":{"login":"ivan"},"requested_reviewer":{"login":"judy"}}]`,
			wantType:    "review_requested",
			wantActor:   "ivan",
			wantSummary: "judy",
		},
		{
			name:        "merged calls out the merging actor",
			json:        `[{"event":"merged","created_at":"2026-01-01T10:00:00Z","actor":{"login":"kate"}}]`,
			wantType:    "merged",
			wantActor:   "kate",
			wantSummary: "merged by kate",
		},
		{
			name:        "unknown events still come through with raw type",
			json:        `[{"event":"subscribed","created_at":"2026-01-01T10:00:00Z","actor":{"login":"leo"}}]`,
			wantType:    "subscribed",
			wantActor:   "leo",
			wantSummary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := fetchOne(t, tt.json)
			assertEvent(t, e, tt.wantType, tt.wantActor, tt.wantSummary)
			if tt.checkRef != nil {
				tt.checkRef(t, e.Ref)
			}
		})
	}
}

func fetchOne(t *testing.T, payload string) timeline.Event {
	t.Helper()
	client := &fakeClient{
		t:     t,
		pages: [][]byte{[]byte(payload)},
		links: []string{""},
	}
	events, err := timeline.Fetch(client, timeline.Repo{Owner: "o", Name: "r"}, 1)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	return events[0]
}

func assertEvent(t *testing.T, got timeline.Event, wantType, wantActor, wantSummary string) {
	t.Helper()
	if got.Type != wantType {
		t.Errorf("type = %q, want %q", got.Type, wantType)
	}
	if got.Actor != wantActor {
		t.Errorf("actor = %q, want %q", got.Actor, wantActor)
	}
	if got.Summary != wantSummary {
		t.Errorf("summary = %q, want %q", got.Summary, wantSummary)
	}
}

// TestFetch_truncatesLongFreeFormSummaries asserts the 72-rune cap applies
// uniformly to free-form fields (commit message, comment body) including
// multi-byte text.
func TestFetch_truncatesLongFreeFormSummaries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		event      string
		field      string
		input      string
		wantSuffix string
	}{
		{
			name:       "long commit message truncated with ellipsis",
			event:      "committed",
			field:      "message",
			input:      strings.Repeat("a", 100),
			wantSuffix: "…",
		},
		{
			name:       "multi-byte commit message truncated by rune count",
			event:      "committed",
			field:      "message",
			input:      strings.Repeat("あ", 80),
			wantSuffix: "…",
		},
		{
			name:       "long comment body truncated with ellipsis",
			event:      "commented",
			field:      "body",
			input:      strings.Repeat("x", 100),
			wantSuffix: "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload string
			if tt.event == "committed" {
				payload = `[{"event":"committed","sha":"x","message":"` + tt.input + `","committer":{"date":"2026-01-01T10:00:00Z","name":"a"},"author":{"name":"a"}}]`
			} else {
				payload = `[{"event":"commented","created_at":"2026-01-01T10:00:00Z","actor":{"login":"a"},"body":"` + tt.input + `","id":1}]`
			}
			client := &fakeClient{t: t, pages: [][]byte{[]byte(payload)}, links: []string{""}}
			events, err := timeline.Fetch(client, timeline.Repo{Owner: "o", Name: "r"}, 1)
			if err != nil {
				t.Fatalf("Fetch: %v", err)
			}
			summary := events[0].Summary
			if !strings.HasSuffix(summary, tt.wantSuffix) {
				t.Errorf("summary %q should end with %q", summary, tt.wantSuffix)
			}
			if utf8RuneCount(summary) != 72 {
				t.Errorf("summary rune count = %d, want 72: %q", utf8RuneCount(summary), summary)
			}
		})
	}
}

func utf8RuneCount(s string) int { return len([]rune(s)) }

func firstRequest(t *testing.T, reqs []string) string {
	t.Helper()
	if len(reqs) == 0 {
		t.Fatal("no requests recorded")
	}
	return reqs[0]
}

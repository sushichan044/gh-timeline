package timeline_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sushichan044/gh-timeline/internal/timeline"
)

func sampleEvents() []timeline.Event {
	return []timeline.Event{
		{
			Type:      "reviewed",
			Actor:     "bob",
			Timestamp: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			Summary:   "approved",
			Ref:       timeline.Ref{ReviewID: 42, URL: "https://api.github.com/.../reviews/42"},
		},
		{
			Type:      "labeled",
			Actor:     "alice",
			Timestamp: time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC),
			Summary:   "bug",
		},
	}
}

func TestRenderText_oneLinePerEvent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := timeline.RenderText(&buf, sampleEvents()); err != nil {
		t.Fatalf("RenderText: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), buf.String())
	}
	if lines[0] != "2026-01-01T09:00:00Z [reviewed] @bob: approved" {
		t.Errorf("line 1 = %q", lines[0])
	}
	if lines[1] != "2026-01-02T10:30:00Z [labeled] @alice: bug" {
		t.Errorf("line 2 = %q", lines[1])
	}
}

func TestRenderText_emptyActorAndSummaryUseDash(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	events := []timeline.Event{
		{Type: "subscribed", Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	if err := timeline.RenderText(&buf, events); err != nil {
		t.Fatalf("RenderText: %v", err)
	}
	if got := strings.TrimRight(buf.String(), "\n"); got != "2026-01-01T00:00:00Z [subscribed] @-: -" {
		t.Errorf("got %q", got)
	}
}

func TestRenderJSON_emitsArray(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := timeline.RenderJSON(&buf, sampleEvents()); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	var decoded []timeline.Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("decoded %d events, want 2", len(decoded))
	}
	if decoded[0].Ref.ReviewID != 42 {
		t.Errorf("review ID lost in JSON round-trip: %+v", decoded[0].Ref)
	}
}

func TestRenderJSON_emptyInputEmitsEmptyArray(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := timeline.RenderJSON(&buf, nil); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Errorf("got %q, want []", got)
	}
}

package timeline

import (
	"encoding/json"
	"fmt"
	"io"
)

// RenderText writes events as one line per event for human consumption.
// Format: `2026-01-02T10:00:00Z [labeled] @alice: bug`.
func RenderText(w io.Writer, events []Event) error {
	for _, e := range events {
		actor := e.Actor
		if actor == "" {
			actor = "-"
		}
		summary := e.Summary
		if summary == "" {
			summary = "-"
		}
		_, err := fmt.Fprintf(w, "%s [%s] @%s: %s\n",
			e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			e.Type,
			actor,
			summary,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderJSON writes events as an indented JSON array. Always emits a valid
// array, even when empty.
func RenderJSON(w io.Writer, events []Event) error {
	if events == nil {
		events = []Event{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(events)
}

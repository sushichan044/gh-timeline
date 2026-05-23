package timeline

import (
	"encoding/json"
	"fmt"
	"io"
)

// RenderText writes events as one line per event for human consumption.
//
// When Summary is non-empty the line is `{ts} [{type}] @{actor}: {summary}`.
// When Summary is empty — the fallback path for unknown / known-but-noisy
// events — the trailing `: -` is dropped and the line ends after the actor.
func RenderText(w io.Writer, events []Event) error {
	for _, e := range events {
		actor := e.Actor
		if actor == "" {
			actor = "-"
		}
		ts := e.Timestamp.UTC().Format("2006-01-02T15:04:05Z")
		if e.Summary == "" {
			if _, err := fmt.Fprintf(w, "%s [%s] @%s\n", ts, e.Type, actor); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s [%s] @%s: %s\n", ts, e.Type, actor, e.Summary); err != nil {
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

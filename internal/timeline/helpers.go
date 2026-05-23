package timeline

import (
	"strings"
	"unicode/utf8"
)

// truncateLength is the per-event maximum number of runes used for free-form
// summary text. Chosen per spec — 72 matches the conventional commit subject
// limit and keeps a single timeline line readable in an 80-column terminal.
const truncateLength = 72

// firstLine returns the substring up to (but not including) the first '\n' or
// '\r'. Trailing whitespace on that line is also stripped because commit
// messages often carry it.
func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimRight(s, " \t")
}

// truncate clamps s to truncateLength runes, appending an ellipsis when text
// was cut. Counted in runes (not bytes) so multi-byte characters are handled
// safely.
func truncate(s string) string {
	if utf8.RuneCountInString(s) <= truncateLength {
		return s
	}
	runes := []rune(s)
	return string(runes[:truncateLength-1]) + "…"
}

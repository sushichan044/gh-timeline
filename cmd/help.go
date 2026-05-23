package cmd

import (
	"fmt"
	"io"
	"io/fs"
)

const humanHelp = `gh timeline — render a Pull Request's full event timeline.

Usage:
  gh timeline [flags] <pr-number>
  gh timeline skills <subcommand> [flags]

Flags:
  -R, --repo OWNER/REPO   Repository (defaults to the current git repository)
      --json              Emit normalized JSON instead of text
      --no-json           Force text output even when running under an AI agent
      --version           Print version
  -h, --help              Show this help

Examples:
  gh timeline 1234
  gh timeline --repo cli/cli 1234
  gh timeline --json --repo cli/cli 1234 | jq '.[].type'
  gh timeline skills install            # install the embedded AI agent skill

When run under an AI agent (Claude Code, Cursor, Codex, …) the default output
switches to JSON automatically. Pass --no-json to opt out.
`

// writeHelp prints the human-facing usage, or the embedded SKILL.md body for
// agent runtimes. The agent variant is the same content `skills install`
// drops on disk, so an agent can read it directly from --help.
func writeHelp(w io.Writer, agent bool, skillFS fs.FS) {
	if agent && skillFS != nil {
		body, err := skillContent(skillFS)
		if err == nil {
			_, _ = w.Write(body)
			return
		}
		// Falling through to the human help is safer than printing nothing.
		fmt.Fprintf(w, "warning: could not load embedded skill content: %v\n\n", err)
	}
	_, _ = io.WriteString(w, humanHelp)
}

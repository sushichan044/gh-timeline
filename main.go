// Command gh-timeline is a GitHub CLI extension that prints a Pull Request's
// full timeline of events in chronological order. The skill bundle is
// embedded by [embed.go] at this same package so the //go:embed directive can
// reach the top-level `skills/` directory.
package main

import (
	"context"
	"os"

	"github.com/sushichan044/gh-timeline/cmd"
)

func main() {
	os.Exit(cmd.Run(context.Background(), os.Args, os.Stdout, os.Stderr, skillFS, referenceMD))
}

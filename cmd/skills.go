package cmd

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/Songmu/skillsmith"

	"github.com/sushichan044/gh-timeline/internal/version"
)

const (
	subcommandSkills = "skills"
	skillName        = "gh-timeline"
)

// runSkills handles the `gh timeline skills <subcommand>` family by delegating
// to the embedded skillsmith dispatcher. It is the single point of integration
// with the skill bundle — everything skill-related (loading, install, list,
// status, etc.) goes through here.
func runSkills(ctx context.Context, args []string, stderr io.Writer, newSkills func() (SkillsRunner, error)) int {
	s, err := newSkills()
	if err != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", err)
		return exitError
	}
	if runErr := s.Run(ctx, args); runErr != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", runErr)
		return exitError
	}
	return exitOK
}

// newSmith returns a configured skillsmith instance bound to skillFS, which
// must contain a single top-level `skills/` directory holding the
// gh-timeline skill files.
func newSmith(skillFS fs.FS) (*skillsmith.Smith, error) {
	return skillsmith.New(skillName, semverFor(version.Get()), skillFS)
}

// semverFor turns the runtime version string ("v1.2.3 (rev: abc1234)",
// "dev", "unknown") into something skillsmith's semver validation accepts.
func semverFor(v string) string {
	if i := strings.IndexByte(v, ' '); i > 0 {
		v = v[:i]
	}
	switch v {
	case "", "dev", "unknown":
		return "v0.0.0-dev"
	}
	return v
}

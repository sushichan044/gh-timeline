// Package cmd implements the gh-timeline CLI surface. main.go owns the
// embedded skill filesystem and passes it in via [Run].
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/jehiah/agentdetection"
	"github.com/spf13/pflag"

	"github.com/sushichan044/gh-timeline/internal/timeline"
	"github.com/sushichan044/gh-timeline/internal/version"
)

// Exit codes used throughout the CLI.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// Run is the binary entry point. It parses args, dispatches subcommands, and
// returns the process exit code. skillFS must hold the embedded skill bundle
// owned by main.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer, skillFS fs.FS) int {
	return RunWithDeps(ctx, args, stdout, stderr, defaultDeps(skillFS))
}

// Deps groups host-dependent collaborators so tests can substitute them.
// Production code uses [Run] which fills these from real implementations.
type Deps struct {
	IsAgent     func() bool
	NewClient   func() (timeline.RESTClient, error)
	CurrentRepo func() (timeline.Repo, error)
	NewSkills   func() (SkillsRunner, error)
	SkillFS     fs.FS // read by writeHelp for agent --help; not required for non-agent runs
}

// SkillsRunner is the bit of *skillsmith.Smith [Run] actually uses, kept tiny
// for substitutability.
type SkillsRunner interface {
	Run(ctx context.Context, args []string) error
}

func defaultDeps(skillFS fs.FS) Deps {
	return Deps{
		IsAgent: agentdetection.IsAgent,
		NewClient: func() (timeline.RESTClient, error) {
			return api.DefaultRESTClient()
		},
		CurrentRepo: func() (timeline.Repo, error) {
			r, err := repository.Current()
			if err != nil {
				return timeline.Repo{}, err
			}
			return timeline.Repo{Owner: r.Owner, Name: r.Name}, nil
		},
		NewSkills: func() (SkillsRunner, error) {
			return newSmith(skillFS)
		},
		SkillFS: skillFS,
	}
}

// RunWithDeps is the test-friendly entry point — production code uses [Run]
// which calls this with real implementations.
func RunWithDeps(ctx context.Context, args []string, stdout, stderr io.Writer, d Deps) int {
	rest := args
	if len(rest) > 0 {
		rest = rest[1:]
	}

	if len(rest) > 0 && rest[0] == subcommandSkills {
		return runSkills(ctx, rest[1:], stderr, d.NewSkills)
	}

	return runTimeline(rest, stdout, stderr, d)
}

func runTimeline(args []string, stdout, stderr io.Writer, d Deps) int {
	opts, err := parseFlags(args, stderr, d.SkillFS)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return exitOK
		}
		fmt.Fprintf(stderr, "gh timeline: %v\n", err)
		return exitUsage
	}

	agent := d.IsAgent()

	if opts.help {
		writeHelp(stdout, agent, d.SkillFS)
		return exitOK
	}
	if opts.showVersion {
		fmt.Fprintf(stdout, "gh-timeline %s\n", version.Get())
		return exitOK
	}

	if opts.prNumber <= 0 {
		fmt.Fprintln(stderr, "gh timeline: missing PR number")
		writeHelp(stderr, false, d.SkillFS)
		return exitUsage
	}

	repo, err := resolveRepo(opts.repoFlag, d.CurrentRepo)
	if err != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", err)
		return exitError
	}

	client, err := d.NewClient()
	if err != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", err)
		return exitError
	}

	events, err := timeline.Fetch(client, repo, opts.prNumber)
	if err != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", err)
		return exitError
	}

	if useJSON(opts, agent) {
		if renderErr := timeline.RenderJSON(stdout, events); renderErr != nil {
			fmt.Fprintf(stderr, "gh timeline: %v\n", renderErr)
			return exitError
		}
	} else if renderErr := timeline.RenderText(stdout, events); renderErr != nil {
		fmt.Fprintf(stderr, "gh timeline: %v\n", renderErr)
		return exitError
	}
	return exitOK
}

func useJSON(opts flagOpts, agent bool) bool {
	if opts.jsonSet {
		return opts.jsonOutput
	}
	return agent
}

func resolveRepo(repoFlag string, current func() (timeline.Repo, error)) (timeline.Repo, error) {
	if repoFlag == "" {
		repo, err := current()
		if err != nil {
			return timeline.Repo{}, fmt.Errorf(
				"could not determine current repository (pass --repo OWNER/REPO): %w",
				err,
			)
		}
		return repo, nil
	}
	r, err := repository.Parse(repoFlag)
	if err != nil {
		return timeline.Repo{}, fmt.Errorf("invalid --repo %q: %w", repoFlag, err)
	}
	return timeline.Repo{Owner: r.Owner, Name: r.Name}, nil
}

type flagOpts struct {
	repoFlag    string
	jsonOutput  bool
	jsonSet     bool // distinguishes "user did not pass --json/--no-json" from explicit false
	help        bool
	showVersion bool
	prNumber    int
}

func parseFlags(args []string, stderr io.Writer, skillFS fs.FS) (flagOpts, error) {
	fs := pflag.NewFlagSet("gh timeline", pflag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { writeHelp(stderr, false, skillFS) }

	var (
		opts       flagOpts
		jsonFlag   bool
		noJSONFlag bool
	)
	fs.StringVarP(&opts.repoFlag, "repo", "R", "", "Repository in OWNER/REPO format (default: current repo)")
	fs.BoolVar(&jsonFlag, "json", false, "Emit normalized JSON instead of text")
	fs.BoolVar(&noJSONFlag, "no-json", false, "Force text output even under an AI agent")
	fs.BoolVarP(&opts.help, "help", "h", false, "Show help")
	fs.BoolVar(&opts.showVersion, "version", false, "Show version")

	if err := fs.Parse(args); err != nil {
		return flagOpts{}, err
	}

	switch {
	case fs.Changed("json") && fs.Changed("no-json"):
		return flagOpts{}, errors.New("--json and --no-json are mutually exclusive")
	case fs.Changed("json"):
		opts.jsonSet = true
		opts.jsonOutput = jsonFlag
	case fs.Changed("no-json"):
		opts.jsonSet = true
		opts.jsonOutput = !noJSONFlag
	}

	rest := fs.Args()
	if len(rest) > 1 {
		return flagOpts{}, fmt.Errorf("expected one PR number, got %d extra args", len(rest)-1)
	}
	if len(rest) == 1 {
		n, err := strconv.Atoi(rest[0])
		if err != nil {
			return flagOpts{}, fmt.Errorf("invalid PR number %q", rest[0])
		}
		opts.prNumber = n
	}
	return opts, nil
}

package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/sushichan044/gh-timeline/cmd"
	"github.com/sushichan044/gh-timeline/internal/timeline"
)

// stubQuerier is a no-op timeline.GraphQLQuerier — cmd-level tests inject a
// canned Fetch result, so the querier is never actually consulted.
type stubQuerier struct{}

func (stubQuerier) Query(context.Context, any, map[string]any) error { return nil }

type fakeSkills struct {
	called bool
	args   []string
	err    error
}

func (f *fakeSkills) Run(_ context.Context, args []string) error {
	f.called = true
	f.args = args
	return f.err
}

// newTestSkillFS mimics the embedded skill bundle so tests can assert that
// agent --help reads from it without depending on the real //go:embed output.
func newTestSkillFS() fstest.MapFS {
	return fstest.MapFS{
		"skills/gh-timeline/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: gh-timeline\n---\n\n# gh-timeline skill body\n"),
		},
	}
}

func newTestDeps(events []timeline.Event, fetchErr error, agent bool, fs *fakeSkills) cmd.Deps {
	return cmd.Deps{
		IsAgent: func() bool { return agent },
		NewClient: func(context.Context) (timeline.GraphQLQuerier, error) {
			return stubQuerier{}, nil
		},
		Fetch: func(context.Context, timeline.GraphQLQuerier, timeline.Repo, int) ([]timeline.Event, error) {
			return events, fetchErr
		},
		CurrentRepo: func() (timeline.Repo, error) {
			return timeline.Repo{Owner: "cli", Name: "cli"}, nil
		},
		NewSkills: func() (cmd.SkillsRunner, error) { return fs, nil },
		SkillFS:   newTestSkillFS(),
	}
}

func reviewedEvent() timeline.Event {
	return timeline.Event{
		Type:      "PullRequestReview",
		Actor:     "bob",
		Timestamp: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
		Summary:   "APPROVED",
	}
}

func TestRun_textRenderingByDefault(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "123"},
		&stdout, &stderr,
		newTestDeps([]timeline.Event{reviewedEvent()}, nil, false, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[PullRequestReview] @bob: APPROVED") {
		t.Errorf("missing rendered event line, got: %q", stdout.String())
	}
}

func TestRun_agentRuntimeSwitchesToJSON(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "123"},
		&stdout, &stderr,
		newTestDeps([]timeline.Event{reviewedEvent()}, nil, true, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	var decoded []timeline.Event
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("expected JSON output, got %q: %v", stdout.String(), err)
	}
	if len(decoded) != 1 || decoded[0].Type != "PullRequestReview" {
		t.Errorf("unexpected events: %+v", decoded)
	}
}

func TestRun_noJSONForcesTextEvenForAgent(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "--no-json", "123"},
		&stdout, &stderr,
		newTestDeps([]timeline.Event{reviewedEvent()}, nil, true, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if strings.HasPrefix(strings.TrimSpace(stdout.String()), "[") {
		t.Errorf("expected text output, got JSON-shaped: %q", stdout.String())
	}
}

func TestRun_agentHelpEmitsSkillContent(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "--help"},
		&stdout, &stderr,
		newTestDeps(nil, nil, true, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "---") || !strings.Contains(out, "gh-timeline") {
		t.Errorf("agent help should be the SKILL.md body, got: %q", out[:min(80, len(out))])
	}
}

func TestRun_humanHelpForNonAgents(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "--help"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("human help missing Usage section: %q", stdout.String())
	}
}

func TestRun_skillsSubcommandDelegates(t *testing.T) {
	t.Parallel()
	fs := &fakeSkills{}
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "skills", "install", "--dry-run"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, fs),
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !fs.called {
		t.Fatal("fakeSkills.Run was not invoked")
	}
	if len(fs.args) != 2 || fs.args[0] != "install" || fs.args[1] != "--dry-run" {
		t.Errorf("skills subcommand received %v", fs.args)
	}
}

func TestRun_skillsErrorBubblesUp(t *testing.T) {
	t.Parallel()
	fs := &fakeSkills{err: errors.New("install failed")}
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "skills", "install"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, fs),
	)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "install failed") {
		t.Errorf("stderr should contain underlying error, got %q", stderr.String())
	}
}

func TestRun_missingPRNumberExits2(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, &fakeSkills{}),
	)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing PR number") {
		t.Errorf("stderr should mention missing PR number, got %q", stderr.String())
	}
}

func TestRun_versionFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "--version"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, &fakeSkills{}),
	)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.HasPrefix(stdout.String(), "gh-timeline ") {
		t.Errorf("version line malformed: %q", stdout.String())
	}
}

func TestRun_jsonAndNoJSONAreMutuallyExclusive(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "--json", "--no-json", "123"},
		&stdout, &stderr,
		newTestDeps(nil, nil, false, &fakeSkills{}),
	)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Errorf("stderr should explain mutual exclusion, got %q", stderr.String())
	}
}

func TestRun_fetchErrorExits1(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := cmd.RunWithDeps(context.Background(),
		[]string{"gh-timeline", "123"},
		&stdout, &stderr,
		newTestDeps(nil, errors.New("octo/demo#123 not found"), false, &fakeSkills{}),
	)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Errorf("stderr should surface the underlying error, got %q", stderr.String())
	}
}

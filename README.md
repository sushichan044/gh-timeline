# gh-timeline

A [`gh`](https://cli.github.com/) extension that prints an Issue or Pull
Request's full event timeline — commits, reviews, comments, force pushes,
labels, assignments, merges, and every other timeline item GitHub exposes —
in a single chronological view.

Data is fetched via GitHub's GraphQL API
(`issueOrPullRequest.timelineItems`), so the same `<number>` argument works
for both issues and PRs.

Designed to be readable in the terminal **and** trivially parseable by AI
agents. When invoked under an AI agent runtime (Claude Code, Cursor, Codex,
Gemini CLI, …) the extension auto-switches to JSON output and exposes its
embedded skill via `--help`.

## Install

```sh
gh extension install sushichan044/gh-timeline
```

To build from a local clone:

```sh
gh extension install .
```

## Usage

The positional argument is either an Issue/PR number or a full GitHub URL —
both work for issues and PRs. GitHub Enterprise Server URLs are accepted as
long as the host is reachable with the same `gh auth` configuration.

```sh
# Inside a clone of the repo
gh timeline 1234

# Or from anywhere
gh timeline --repo cli/cli 1234

# Paste a URL directly (no --repo needed)
gh timeline https://github.com/cli/cli/pull/1234
gh timeline https://github.com/cli/cli/issues/5678

# Force JSON
gh timeline --json --repo cli/cli 1234 | jq '.[] | {type, actor, timestamp}'

# Force text even under an AI agent
gh timeline --no-json --repo cli/cli 1234
```

`--repo` cannot be combined with a URL — the URL already carries the
repository, and silently overriding it would surprise users.

### Flags

| Flag                    | Description                                           |
| ----------------------- | ----------------------------------------------------- |
| `-R, --repo OWNER/REPO` | Repository (defaults to the current git repository)   |
| `--json`                | Emit normalized JSON instead of text                  |
| `--no-json`             | Force text output even when running under an AI agent |
| `--version`             | Print version                                         |
| `-h, --help`            | Show help                                             |

## Output format

### Text (default for humans)

```
2026-01-02T10:00:00Z [LabeledEvent] @alice: bug
2026-01-02T10:05:00Z [PullRequestReview] @bob: APPROVED
2026-01-02T10:30:00Z [MergedEvent] @carol: merged by carol
2026-01-02T10:31:00Z [SubscribedEvent] @dave
```

When an event has no meaningful one-line summary (e.g. `SubscribedEvent`, or
a brand-new `__typename` the extension does not have a typed handler for),
the trailing `: <summary>` segment is dropped and the line ends after the
actor.

### JSON (default for AI agents, or with `--json`)

```json
[
  {
    "type": "PullRequestReview",
    "actor": "bob",
    "timestamp": "2026-01-02T10:00:00Z",
    "summary": "APPROVED",
    "ref": {
      "node_id": "PRR_kwDO…",
      "review_id": 1234567,
      "url": "https://github.com/OWNER/REPO/pull/123#pullrequestreview-1234567"
    }
  }
]
```

The `type` value is the GraphQL `__typename` of the event (PascalCase).
Summaries are truncated to 72 characters. Use the `ref.sha`,
`ref.review_id`, `ref.comment_id`, or `ref.url` fields with `gh api` to fetch
full content when needed.

## AI agent skill

The extension ships an embedded
[AgentSkill](https://agentskills.io/) describing how an agent should call it.
Install it locally with:

```sh
gh timeline skills install      # writes ~/.agents/skills/gh-timeline/SKILL.md
gh timeline skills status
gh timeline skills uninstall
```

## Releases

Tagging `vX.Y.Z` triggers
[`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile)
to build and attach cross-platform binaries (linux/darwin/windows ×
amd64/arm64).

## Repository topic (manual step)

`gh` discovers extensions by the `gh-extension` topic. After the first
release, set the topic on the GitHub repository settings page — there is no
API for this from the extension itself.

## License

MIT

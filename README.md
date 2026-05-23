# gh-timeline

![```bash gh timeline https://github.com/sushichan044/cc-hooks-ts/issues/20
2025-11-16T07:32:34Z [AssignedEvent] @sushichan044: assigned sushichan044
2025-11-16T07:59:08Z [IssueComment] @sushichan044: done by https://github.com/sushichan044/cc-hooks-ts/pull/21
2025-11-16T07:59:08Z [ClosedEvent] @sushichan044: closed: COMPLETED```](/docs/assets/timeline.png)

A [`gh`](https://cli.github.com/) extension to view the full timeline of any GitHub issue or PR.

When invoked under an AI agent runtime (Claude Code, Cursor, Codex, Gemini CLI, …) the
extension auto-switches to JSON output.

## Install

```sh
gh extension install sushichan044/gh-timeline
```

To build from a local clone:

```sh
gh extension install .
```

## Usage

```sh
# Inside a clone of the repo
gh timeline 12345

# Or from anywhere
gh timeline --repo cli/cli 12345

# Paste a URL directly (no --repo needed)
gh timeline https://github.com/cli/cli/issues/12345
gh timeline https://github.com/cli/cli/pull/5677

# Force JSON
gh timeline --json --repo cli/cli 12345 | jq '.[] | {type, actor, timestamp}'

# Force text even under an AI agent
gh timeline --no-json --repo cli/cli 12345
```

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
$ gh timeline https://github.com/golang/go/issues/50000
2021-12-06T20:30:19Z [IssueComment] @toothrot: #26479 may also be relevant.
2021-12-06T20:30:29Z [MilestonedEvent] @toothrot: added to milestone "Backlog"
2021-12-06T20:30:41Z [LabeledEvent] @toothrot: added label NeedsInvestigation
2021-12-07T02:23:45Z [IssueComment] @zzkcode: @toothrot Thanks for your reply.
2021-12-07T02:23:45Z [MentionedEvent] @toothrot
2021-12-07T02:23:45Z [SubscribedEvent] @toothrot
2021-12-12T08:17:39Z [IssueComment] @zzkcode: Finally, I figure it out and everything is working as expected, so it's…
2021-12-12T08:17:39Z [MentionedEvent] @toothrot
2021-12-12T08:17:39Z [SubscribedEvent] @toothrot
2021-12-12T08:17:53Z [ClosedEvent] @zzkcode: closed: COMPLETED
2022-12-12T09:46:12Z [LockedEvent] @golang: locked
2022-12-12T09:46:13Z [LabeledEvent] @gopherbot: added label FrozenDueToAge
```

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

`type` is the GraphQL `__typename` (PascalCase). Summaries are truncated to
72 characters. Use `ref.sha`, `ref.review_id`, `ref.comment_id`, or
`ref.url` with `gh api` to fetch full content when needed.

## AI agent skill

The extension ships an embedded
[AgentSkill](https://agentskills.io/) describing how an agent should call it.
Install it locally with:

```sh
gh timeline skills install      # writes ~/.agents/skills/gh-timeline/SKILL.md
gh timeline skills status
gh timeline skills uninstall
```

## License

MIT

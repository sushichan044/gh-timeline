# gh-timeline

![First 10 events of https://github.com/cli/cli/issues/10147 retrieved with `gh timeline https://github.com/cli/cli/issues/10147`](/docs/assets/timeline.png)

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

```bash
$ gh timeline https://github.com/cli/cli/issues/10147
2024-12-27T21:11:39Z [LabeledEvent] @cphi-github: added label enhancement
2024-12-27T21:11:50Z [LabeledEvent] @cliAutomation: added label needs-triage
2024-12-27T21:11:53Z [RenamedTitleEvent] @cphi-github: renamed: "Add ability to request timeline events to `gh issue view` response" → "Add ability to request timeline events in `gh issue view` response"
2025-01-02T23:28:45Z [IssueComment] @jtmcg: Hey @cphi-github, thanks for the details on the issue and the problem y…
2025-01-02T23:28:46Z [MentionedEvent] @cphi-github
2025-01-02T23:28:46Z [SubscribedEvent] @cphi-github
2025-01-02T23:28:51Z [AssignedEvent] @jtmcg: assigned jtmcg
2025-01-02T23:29:06Z [LabeledEvent] @jtmcg: added label more-info-needed
2025-01-06T13:51:04Z [IssueComment] @EdouardF: Hello @jtmcg,
2025-01-06T13:51:06Z [MentionedEvent] @jtmcg
2025-01-06T13:51:06Z [SubscribedEvent] @jtmcg
2025-01-09T19:38:41Z [IssueComment] @cphi-github: @jtmcg - I will have a poke when I get a chance!
2025-01-09T19:38:42Z [MentionedEvent] @jtmcg
2025-01-09T19:38:42Z [SubscribedEvent] @jtmcg
2025-01-09T23:41:20Z [IssueComment] @jtmcg: > There is already deviation between the plaintext and JSON output, in …
2025-01-09T23:41:22Z [MentionedEvent] @EdouardF
2025-01-09T23:41:22Z [SubscribedEvent] @EdouardF
2025-01-09T23:41:22Z [MentionedEvent] @cphi-github
2025-01-09T23:41:22Z [SubscribedEvent] @cphi-github
2025-01-09T23:41:33Z [LabeledEvent] @jtmcg: added label needs-design
2025-01-09T23:41:33Z [LabeledEvent] @jtmcg: added label discuss
2025-01-09T23:49:02Z [IssueComment] @cphi-github: I like it. It sounds like it would meet my use case, while being pretty…
2025-01-13T14:56:25Z [SubscribedEvent] @mgnsk
2025-01-21T20:38:43Z [UnlabeledEvent] @jtmcg: removed label discuss
2025-01-21T21:07:46Z [LabeledEvent] @jtmcg: added label needs-product
2025-01-21T21:07:52Z [UnlabeledEvent] @jtmcg: removed label more-info-needed
2025-01-21T21:07:52Z [UnlabeledEvent] @jtmcg: removed label needs-triage
2025-01-21T21:09:54Z [IssueComment] @jtmcg: Hey, @cphi-github, we talked to some of our internal stakeholders to un…
2025-01-21T21:09:56Z [MentionedEvent] @cphi-github
2025-01-21T21:09:56Z [SubscribedEvent] @cphi-github
2025-02-21T18:08:18Z [UnassignedEvent] @jtmcg: unassigned jtmcg
2026-03-13T14:36:34Z [SubscribedEvent] @Vlaaaaaaad

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

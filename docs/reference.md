# gh timeline — reference

GitHub CLI extension that prints an Issue or Pull Request's complete event
timeline (commits, reviews, comments, force pushes, labels, merges, ...) in
chronological order.

## Invocation

```sh
gh timeline --json --repo OWNER/REPO <NUMBER>

# Or pass the URL directly (must not combine with --repo):
gh timeline --json https://github.com/OWNER/REPO/pull/<NUMBER>
gh timeline --json https://github.com/OWNER/REPO/issues/<NUMBER>
```

Inside a clone of the repo, `--repo` can be omitted. GitHub Enterprise Server
URLs are accepted.

Under AI agent runtimes (Claude Code, Cursor, Codex, ...) `--json` is the
default; pass `--no-json` for text.

## Flags

| Flag                    | Description                                     |
| ----------------------- | ----------------------------------------------- |
| `-R, --repo OWNER/REPO` | Repository (defaults to current git repository) |
| `--json`                | Emit JSON instead of text                       |
| `--no-json`             | Force text output                               |
| `--version`             | Print version                                   |
| `-h, --help`            | Show help                                       |

## JSON output

Array of events. Each element:

```json
{
  "type": "PullRequestReview",
  "actor": "octocat",
  "timestamp": "2026-01-02T10:00:00Z",
  "summary": "APPROVED",
  "ref": {
    "node_id": "PRR_kwDOA...",
    "review_id": 1234567,
    "url": "https://github.com/OWNER/REPO/pull/123#pullrequestreview-1234567"
  }
}
```

- `type` — GraphQL `__typename`, PascalCase. Common values: `PullRequestCommit`,
  `PullRequestReview`, `IssueComment`, `LabeledEvent`, `HeadRefForcePushedEvent`,
  `MergedEvent`, `ClosedEvent`, `ReopenedEvent`, `AssignedEvent`,
  `ReviewRequestedEvent`. Any member of
  [`PullRequestTimelineItems`](https://docs.github.com/en/graphql/reference/unions#pullrequesttimelineitems)
  or [`IssueTimelineItems`](https://docs.github.com/en/graphql/reference/unions#issuetimelineitems)
  may appear.
- `actor` — login of who triggered the event (commit author for
  `PullRequestCommit`).
- `timestamp` — RFC 3339 UTC.
- `summary` — short one-line description, **truncated to 72 characters**. May
  be empty.
- `ref` — identifiers to pass back to `gh api`. Varies by event: commits expose
  `sha`, reviews expose `review_id`, comments expose `comment_id`; most events
  also expose `node_id` and `url`.

## Text output

```
2026-01-02T10:00:00Z [LabeledEvent] @alice: added label bug
2026-01-02T10:05:00Z [PullRequestReview] @bob: APPROVED
2026-01-02T10:30:00Z [MergedEvent] @carol: merged deadbee into main
2026-01-02T10:31:00Z [SubscribedEvent] @dave
```

When `summary` is empty, the trailing `: <summary>` is dropped.

## Fetching full content (summary is truncated)

```sh
gh api repos/OWNER/REPO/commits/<ref.sha>
gh api repos/OWNER/REPO/pulls/<NUMBER>/reviews/<ref.review_id>
gh api repos/OWNER/REPO/pulls/<NUMBER>/reviews/<ref.review_id>/comments
gh api repos/OWNER/REPO/issues/comments/<ref.comment_id>
```

`ref.url` also opens in a browser via `gh browse`.

## Examples

```sh
# Overview
gh timeline --json --repo cli/cli 12345 | jq '.[] | {type, actor, timestamp}'

# From a URL
gh timeline --json https://github.com/cli/cli/pull/12345 \
  | jq '.[] | {type, actor, timestamp}'

# Reviews only, newest first
gh timeline --json --repo cli/cli 12345 \
  | jq '[.[] | select(.type == "PullRequestReview")] | reverse'

# Force pushes
gh timeline --json --repo cli/cli 12345 \
  | jq '.[] | select(.type == "HeadRefForcePushedEvent")'
```

---
name: gh-timeline
description: Render a GitHub Issue or Pull Request's full event timeline in chronological order from the terminal, as JSON for AI agents.
license: MIT
compatibility:
  - claude
  - codex
  - agents
allowed_tools:
  - Bash
---

# gh-timeline

A GitHub CLI extension that prints an Issue or Pull Request's complete event
timeline — commits, reviews, comments, force pushes, labels, assignments,
merges, and many more — in chronological order, in a single call.

The data source is GitHub's GraphQL API (`issueOrPullRequest.timelineItems`),
so both issues and PRs work with the same `<number>` argument.

When you (the AI agent) need the full history of an Issue or PR, prefer
`gh timeline` over combining `gh pr view`, `gh pr review list`,
`gh issue view`, and `gh api .../timeline`.

## When to use

- Investigating an Issue or PR's discussion history end-to-end
- Finding when a force-push happened and what was reviewed afterwards
- Auditing who approved / commented / requested changes and in what order
- Producing a chronological summary for a stand-up or status report

## Recommended invocation

```sh
gh timeline --json --repo OWNER/REPO <NUMBER>

# Or, if you already have the URL of the Issue / PR, paste it directly —
# the repo is inferred from the URL and --repo must NOT be set.
gh timeline --json https://github.com/OWNER/REPO/pull/<NUMBER>
gh timeline --json https://github.com/OWNER/REPO/issues/<NUMBER>
```

When the extension detects an AI agent runtime (Claude Code, Cursor, Codex,
etc.) it switches to JSON output by default, so `--json` is implicit. Pass
`--no-json` to force the human-readable text format.

If you are inside a clone of the repository, `--repo` can be omitted. GitHub
Enterprise Server URLs (`https://<ghe-host>/OWNER/REPO/pull/<NUMBER>`) are
accepted the same way.

## Output schema (JSON)

The command emits a JSON array. Each element has this shape:

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

Field meanings:

- `type` — the GraphQL `__typename` of the event, PascalCase. Examples:
  `PullRequestCommit`, `PullRequestReview`, `IssueComment`, `LabeledEvent`,
  `UnlabeledEvent`, `AssignedEvent`, `UnassignedEvent`, `ReviewRequestedEvent`,
  `HeadRefForcePushedEvent`, `BaseRefForcePushedEvent`, `ReadyForReviewEvent`,
  `ConvertToDraftEvent`, `MergedEvent`, `ClosedEvent`, `ReopenedEvent`,
  `RenamedTitleEvent`, `CrossReferencedEvent`, `ReviewDismissedEvent`,
  `MilestonedEvent`, `LockedEvent`, `PinnedEvent`, `ConnectedEvent`,
  `TransferredEvent`, `MarkedAsDuplicateEvent`, etc. Any other GraphQL
  `__typename` in the
  [`PullRequestTimelineItems`](https://docs.github.com/en/graphql/reference/unions#pullrequesttimelineitems)
  or [`IssueTimelineItems`](https://docs.github.com/en/graphql/reference/unions#issuetimelineitems)
  union may also appear; events the extension does not yet have a typed handler
  for fall through with the raw `type` and an empty `summary`.
- `actor` — login of the user who triggered the event (commit author for
  `PullRequestCommit`).
- `timestamp` — RFC 3339 UTC. Commits use `committedDate`; reviews use
  `submittedAt`; everything else uses `createdAt`.
- `summary` — short human-readable description, **truncated to 72
  characters**. For `PullRequestCommit` and `IssueComment` this is the first
  line only. Empty for events that have no meaningful one-line summary
  (e.g. `SubscribedEvent`) or are entirely unknown to the dispatcher.
- `ref` — identifiers you can pass back to `gh api` to fetch the full
  payload. Populated fields vary by event: commits expose `sha`, reviews
  expose `review_id`, comments expose `comment_id`, and most events expose
  `node_id` and `url`.

## Text output

The default for humans:

```
2026-01-02T10:00:00Z [LabeledEvent] @alice: added label bug
2026-01-02T10:05:00Z [PullRequestReview] @bob: APPROVED
2026-01-02T10:30:00Z [MergedEvent] @carol: merged deadbee into main
2026-01-02T10:31:00Z [SubscribedEvent] @dave
```

When `summary` is empty (known-but-noisy events, or future event types the
dispatcher has not seen yet), the trailing `: <summary>` segment is dropped
and the line ends after the actor.

## Drilling into truncated content

Because `summary` is truncated, use `ref` fields with `gh api` to read the
full content when needed:

```sh
# full commit message + diff
gh api repos/OWNER/REPO/commits/<ref.sha>

# full review body and inline comments
gh api repos/OWNER/REPO/pulls/<NUMBER>/reviews/<ref.review_id>
gh api repos/OWNER/REPO/pulls/<NUMBER>/reviews/<ref.review_id>/comments

# full issue comment body
gh api repos/OWNER/REPO/issues/comments/<ref.comment_id>
```

`ref.url` carries a canonical URL for events that have one — passing it to
`gh browse` (or visiting it directly) also works.

## Examples

```sh
# Quick overview as JSON
gh timeline --json --repo cli/cli 1234 | jq '.[] | {type, actor, timestamp}'

# Same thing, but starting from a URL the user shared
gh timeline --json https://github.com/cli/cli/pull/1234 \
  | jq '.[] | {type, actor, timestamp}'

# Just the reviews, newest first
gh timeline --json --repo cli/cli 1234 \
  | jq '[.[] | select(.type == "PullRequestReview")] | reverse'

# Find force pushes
gh timeline --json --repo cli/cli 1234 \
  | jq '.[] | select(.type == "HeadRefForcePushedEvent")'

# All unique event types in this PR's history
gh timeline --json --repo cli/cli 1234 | jq '[.[].type] | unique | sort'
```

## Limitations

- Truncation is fixed at 72 characters per event. There is no configuration
  knob — use the `ref` fields to get the full content.
- The command paginates the GraphQL API at 100 events per request; very
  large Issues / PRs may take a few seconds.
- A missing Issue / PR returns a non-zero exit code with a message
  containing `not found` on stderr.

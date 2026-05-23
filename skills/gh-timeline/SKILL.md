---
name: gh-timeline
description: Render a GitHub Pull Request's full event timeline in chronological order from the terminal, as JSON for AI agents.
license: MIT
compatibility:
  - claude
  - codex
  - agents
allowed_tools:
  - Bash
---

# gh-timeline

A GitHub CLI extension that prints a Pull Request's complete event timeline —
commits, reviews, comments, force pushes, labels, assignments, merges — in
chronological order, in a single call.

When you (the AI agent) need the full history of a PR, prefer `gh timeline`
over combining `gh pr view`, `gh pr review list`, and `gh api .../timeline`.

## When to use

- Investigating a PR's discussion history end-to-end
- Finding when a force-push happened and what was reviewed afterwards
- Auditing who approved / commented / requested changes and in what order
- Producing a chronological summary for a stand-up or status report

## Recommended invocation

```sh
gh timeline --json --repo OWNER/REPO <PR_NUMBER>
```

When the extension detects an AI agent runtime (Claude Code, Cursor, Codex,
etc.) it switches to JSON output by default, so `--json` is implicit. Pass
`--no-json` to force the human-readable text format.

If you are inside a clone of the repository, `--repo` can be omitted.

## Output schema (JSON)

The command emits a JSON array. Each element has this shape:

```json
{
  "type": "reviewed",
  "actor": "octocat",
  "timestamp": "2026-01-02T10:00:00Z",
  "summary": "approved",
  "ref": {
    "node_id": "PRR_kwDOA...",
    "sha": "",
    "review_id": 1234567,
    "comment_id": 0,
    "url": "https://api.github.com/repos/OWNER/REPO/pulls/123/reviews/1234567"
  }
}
```

Field meanings:

- `type` — one of `committed`, `reviewed`, `commented`, `labeled`, `unlabeled`,
  `assigned`, `unassigned`, `review_requested`, `head_ref_force_pushed`,
  `ready_for_review`, `merged`, or any other event the GitHub API may emit.
- `actor` — login of the user who triggered the event (commit author for
  `committed`).
- `timestamp` — RFC 3339 UTC.
- `summary` — short human-readable description, **truncated to 72 characters**.
  For `committed` and `commented` this is the first line only.
- `ref` — identifiers you can pass back to `gh api` to fetch the full payload.

## Drilling into truncated content

Because `summary` is truncated, use `ref` fields with `gh api` to read the
full content when needed:

```sh
# full commit message + diff
gh api repos/OWNER/REPO/commits/<ref.sha>

# full review body and inline comments
gh api repos/OWNER/REPO/pulls/<PR_NUMBER>/reviews/<ref.review_id>
gh api repos/OWNER/REPO/pulls/<PR_NUMBER>/reviews/<ref.review_id>/comments

# full issue comment body
gh api repos/OWNER/REPO/issues/comments/<ref.comment_id>
```

`ref.url` carries the canonical REST URL for events that have one — passing
it to `gh api` directly also works.

## Examples

```sh
# Quick overview as JSON
gh timeline --json --repo cli/cli 1234 | jq '.[] | {type, actor, timestamp}'

# Just the reviews, newest first
gh timeline --json --repo cli/cli 1234 \
  | jq '[.[] | select(.type=="reviewed")] | reverse'

# Find force pushes
gh timeline --json --repo cli/cli 1234 \
  | jq '.[] | select(.type=="head_ref_force_pushed")'
```

## Limitations

- Truncation is fixed at 72 characters per event. There is no configuration
  knob — use the `ref` fields to get the full content.
- The command paginates the GitHub API at 100 events per request; very large
  PRs may take a few seconds.
- A missing PR returns a non-zero exit code with a message containing
  `not found` on stderr.

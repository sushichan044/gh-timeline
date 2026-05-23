---
name: gh-timeline
description: Render a GitHub Issue or Pull Request's full event timeline in chronological order, as JSON for AI agents.
allowed-tools:
  - Bash(gh timeline *)
---

# gh-timeline

When you need an Issue or PR's complete chronological event history (commits,
reviews, comments, force pushes, labels, merges, etc.) in one call, run:

```sh
gh timeline <issue-or-pr-number-or-URL>
```

Output is JSON by default under an AI agent runtime. For the JSON schema,
flag list, drill-down via `gh api`, and worked examples, run
`gh timeline --help`.

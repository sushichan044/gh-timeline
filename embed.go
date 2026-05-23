package main

import "embed"

// skillFS embeds the agent skill bundle. The contents are surfaced to users
// via `gh timeline skills install` and to AI agents via `gh timeline --help`.
//
// Kept at the repository root because `//go:embed` patterns cannot escape
// the embedding file's directory tree.
//
//go:embed all:skills
var skillFS embed.FS

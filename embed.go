package main

import "embed"

// skillFS embeds the agent skill bundle that `gh timeline skills install`
// drops on disk. Kept at the repository root because `//go:embed` patterns
// cannot escape the embedding file's directory tree.
//
//go:embed all:skills
var skillFS embed.FS

// referenceMD is the full human+AI reference shown by `gh timeline --help`
// under an AI agent runtime. Lives outside skills/ so it never ships in the
// installed skill bundle.
//
//go:embed docs/reference.md
var referenceMD string

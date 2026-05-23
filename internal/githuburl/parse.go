// Package githuburl resolves GitHub Issue / Pull Request web URLs into a
// repository identifier and a number. Repository normalization is delegated to
// go-gh's [repository.ParseWithHost] so host, owner, and repo name handling
// stay consistent with the rest of the gh ecosystem; only the path-level
// dispatch (issues vs pull, the number suffix) lives here.
package githuburl

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// expectedPathSegments is the number of /-separated parts in a canonical
// GitHub issue or pull URL path: <owner>/<repo>/<kind>/<number>.
const expectedPathSegments = 4

// Reference identifies an Issue or Pull Request: the repository it lives in
// and its number. The repository carries the host so GitHub Enterprise Server
// URLs round-trip without losing the host.
type Reference struct {
	Repo   repository.Repository
	Number int
}

// Parse extracts a [Reference] from canonical GitHub web URLs of the form:
//
//	https://<host>/<owner>/<repo>/issues/<number>
//	https://<host>/<owner>/<repo>/pull/<number>
//
// Any host is accepted so GitHub Enterprise Server URLs work the same way.
// Fragments and query strings are ignored. The path must contain exactly four
// segments — anything else (wiki pages, /files suffix, repo root, etc.) is
// rejected so users get a clear error rather than a silent wrong fetch.
//
// Owner / repo / host normalization is handed off to
// [repository.ParseWithHost], so this function only owns the path-level
// dispatch.
func Parse(raw string) (Reference, error) {
	u, err := parseWebURL(raw)
	if err != nil {
		return Reference{}, err
	}
	owner, repoName, n, err := splitIssueOrPRPath(u.Path)
	if err != nil {
		return Reference{}, err
	}
	repo, err := repository.ParseWithHost(owner+"/"+repoName, u.Host)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid repository: %w", err)
	}
	return Reference{Repo: repo, Number: n}, nil
}

// parseWebURL parses raw and confirms it carries the bits we actually need:
// an http(s) scheme and a non-empty host. Anything else is rejected before we
// look at the path.
func parseWebURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("not a valid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, errors.New("missing host")
	}
	return u, nil
}

// splitIssueOrPRPath verifies the path matches /<owner>/<repo>/(issues|pull)/<number>
// exactly and returns the parts. The owner/repo strings are returned as-is —
// [repository.ParseWithHost] does the final normalization.
func splitIssueOrPRPath(path string) (string, string, int, error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) != expectedPathSegments {
		return "", "", 0, fmt.Errorf(
			"path must be /<owner>/<repo>/(issues|pull)/<number>, got %q", path)
	}
	owner, repoName, kind, numStr := segments[0], segments[1], segments[2], segments[3]
	if kind != "issues" && kind != "pull" {
		return "", "", 0, fmt.Errorf("path kind must be 'issues' or 'pull', got %q", kind)
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid number %q", numStr)
	}
	if n <= 0 {
		return "", "", 0, fmt.Errorf("number must be positive, got %d", n)
	}
	return owner, repoName, n, nil
}

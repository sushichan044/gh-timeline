package githuburl_test

import (
	"testing"

	"github.com/sushichan044/gh-timeline/internal/githuburl"
)

func TestParse_validInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		in        string
		wantHost  string
		wantOwner string
		wantRepo  string
		wantNum   int
	}{
		{
			name:     "issue URL on github.com",
			in:       "https://github.com/cli/cli/issues/123",
			wantHost: "github.com", wantOwner: "cli", wantRepo: "cli", wantNum: 123,
		},
		{
			name:     "pull URL on github.com",
			in:       "https://github.com/octo/demo/pull/456",
			wantHost: "github.com", wantOwner: "octo", wantRepo: "demo", wantNum: 456,
		},
		{
			name:     "GHE host",
			in:       "https://ghe.example.com/o/r/pull/7",
			wantHost: "ghe.example.com", wantOwner: "o", wantRepo: "r", wantNum: 7,
		},
		{
			name:     "URL with fragment is ignored",
			in:       "https://github.com/cli/cli/pull/456#issuecomment-1",
			wantHost: "github.com", wantOwner: "cli", wantRepo: "cli", wantNum: 456,
		},
		{
			name:     "URL with query string is ignored",
			in:       "https://github.com/cli/cli/issues/9?foo=bar",
			wantHost: "github.com", wantOwner: "cli", wantRepo: "cli", wantNum: 9,
		},
		{
			name:     "trailing slash is tolerated",
			in:       "https://github.com/cli/cli/pull/1/",
			wantHost: "github.com", wantOwner: "cli", wantRepo: "cli", wantNum: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ref, err := githuburl.Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tc.in, err)
			}
			if ref.Repo.Host != tc.wantHost ||
				ref.Repo.Owner != tc.wantOwner ||
				ref.Repo.Name != tc.wantRepo ||
				ref.Number != tc.wantNum {
				t.Errorf("Parse(%q) = {host=%q owner=%q repo=%q num=%d}, want {host=%q owner=%q repo=%q num=%d}",
					tc.in,
					ref.Repo.Host, ref.Repo.Owner, ref.Repo.Name, ref.Number,
					tc.wantHost, tc.wantOwner, tc.wantRepo, tc.wantNum)
			}
		})
	}
}

func TestParse_invalidInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
	}{
		{"missing scheme", "github.com/cli/cli/pull/1"},
		{"unsupported scheme", "ftp://github.com/cli/cli/pull/1"},
		{"non-issue/pull path", "https://github.com/cli/cli/wiki/Home"},
		{"non-numeric number", "https://github.com/cli/cli/pull/abc"},
		{"zero number", "https://github.com/cli/cli/pull/0"},
		{"negative number", "https://github.com/cli/cli/pull/-1"},
		{"path too short", "https://github.com/cli/cli"},
		{"path too long", "https://github.com/cli/cli/pull/1/files"},
		{"empty owner", "https://github.com//cli/pull/1"},
		{"empty repo", "https://github.com/cli//pull/1"},
		{"no host", "https:///cli/cli/pull/1"},
		{"plain garbage", "not a url"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ref, err := githuburl.Parse(tc.in)
			if err == nil {
				t.Fatalf("Parse(%q) = %+v, want error", tc.in, ref)
			}
		})
	}
}

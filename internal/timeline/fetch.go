package timeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Repo is the minimal repo coordinate Fetch needs.
type Repo struct {
	Owner string
	Name  string
}

// RESTClient is the subset of go-gh's REST client that Fetch uses. *api.RESTClient
// satisfies it; tests pass a fake.
type RESTClient interface {
	Request(method, path string, body io.Reader) (*http.Response, error)
}

// perPage is the upper bound the GitHub REST API allows. Using the maximum
// minimizes round trips for PRs with large timelines without changing the
// public contract.
const perPage = 100

// Fetch loads every timeline event for the given PR and returns them sorted
// chronologically (stable on equal timestamps, preserving server order).
func Fetch(client RESTClient, repo Repo, number int) ([]Event, error) {
	if repo.Owner == "" || repo.Name == "" {
		return nil, errors.New("repository owner and name are required")
	}
	if number <= 0 {
		return nil, fmt.Errorf("invalid PR number %d", number)
	}

	path := fmt.Sprintf("repos/%s/%s/issues/%d/timeline?per_page=%d",
		repo.Owner, repo.Name, number, perPage)
	var all []Event
	for {
		resp, err := client.Request(http.MethodGet, path, nil)
		if err != nil {
			var httpErr *api.HTTPError
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("PR %s/%s#%d not found", repo.Owner, repo.Name, number)
			}
			return nil, fmt.Errorf("timeline request failed: %w", err)
		}
		page, err := decodePage(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		for _, e := range page {
			all = append(all, e.normalize())
		}
		next, ok := nextPage(resp.Header.Get("Link"))
		if !ok {
			break
		}
		path = next
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all, nil
}

func decodePage(body io.Reader) ([]rawEvent, error) {
	var page []rawEvent
	if err := json.NewDecoder(body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decode timeline page: %w", err)
	}
	return page, nil
}

// linkRE matches one entry in an RFC 5988 Link header — used to find the
// `rel="next"` URL the GitHub API emits for paginated endpoints.
var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

func nextPage(linkHeader string) (string, bool) {
	for _, m := range linkRE.FindAllStringSubmatch(linkHeader, -1) {
		if len(m) >= 3 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false
}

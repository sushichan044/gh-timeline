package timeline

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/shurcooL/githubv4"
)

// GraphQLQuerier is the subset of *githubv4.Client that Fetch consumes. Tests
// pass a fake; production code wires in the real client built in cmd.
type GraphQLQuerier interface {
	Query(ctx context.Context, q any, variables map[string]any) error
}

// Fetch loads every timeline event for the given Issue or PR and returns them
// sorted chronologically (stable on equal timestamps, preserving server
// order). It paginates internally; callers see one consolidated slice.
//
// The number argument is the Issue or PR number — GitHub uses a single
// numbering space per repository, so `issueOrPullRequest(number:)` resolves
// either form. Errors surface as wrapped errors; a non-existent
// issue/PR is reported as a "not found" error.
//
//nolint:gocognit // The cursor loop interleaves first-page guard, typename dispatch, and pagination — splitting hides the data flow.
func Fetch(ctx context.Context, client GraphQLQuerier, repo Repo, number int) ([]Event, error) {
	if repo.Owner == "" || repo.Name == "" {
		return nil, errors.New("repository owner and name are required")
	}
	if number <= 0 || number > math.MaxInt32 {
		return nil, fmt.Errorf("invalid issue/PR number %d", number)
	}

	var (
		all       []Event
		cursor    *githubv4.String
		firstPage = true
	)
	for {
		var q timelineQuery
		vars := map[string]any{
			"owner": githubv4.String(repo.Owner),
			"name":  githubv4.String(repo.Name),
			// Bounded above by MaxInt32 in the input validation; safe to narrow.
			"number": githubv4.Int(int32(number)),
			"cursor": cursor,
		}
		if err := client.Query(ctx, &q, vars); err != nil {
			return nil, fmt.Errorf("timeline query failed: %w", err)
		}

		typename := string(q.Repository.IssueOrPullRequest.Typename)
		if firstPage {
			if typename == "" {
				return nil, fmt.Errorf("%s/%s#%d not found", repo.Owner, repo.Name, number)
			}
			firstPage = false
		}

		// shurcooL/githubv4 populates both inline-fragment branches from the same
		// JSON object since `timelineItems` shares the key on the wire — pick
		// whichever branch matches the actual __typename to avoid emitting each
		// node twice with one set of zero values.
		var (
			page    pageInfo
			handled bool
		)
		switch typename {
		case "PullRequest":
			prPage := q.Repository.IssueOrPullRequest.PullRequest.TimelineItems
			for _, n := range prPage.Nodes {
				all = append(all, dispatchPRNode(n))
			}
			page = prPage.PageInfo
			handled = true
		case "Issue":
			issuePage := q.Repository.IssueOrPullRequest.Issue.TimelineItems
			for _, n := range issuePage.Nodes {
				all = append(all, dispatchIssueNode(n))
			}
			page = issuePage.PageInfo
			handled = true
		}
		if !handled {
			return nil, fmt.Errorf("unexpected issueOrPullRequest typename %q", typename)
		}

		if !bool(page.HasNextPage) {
			break
		}
		nextCursor := page.EndCursor
		cursor = &nextCursor
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all, nil
}

package timeline

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	graphql "github.com/cli/shurcooL-graphql"
	"golang.org/x/sync/errgroup"
)

const (
	// timelinePageSize is GitHub's GraphQL maximum for `first:`.
	timelinePageSize = 100
	// timelineMaxConcurrency caps in-flight GraphQL queries to stay well below
	// GitHub's secondary rate limits while still parallelizing N→2 round trips.
	timelineMaxConcurrency = 10
)

// GraphQLQuerier is the subset of *api.GraphQLClient that Fetch consumes. Tests
// pass a fake; production code wires in the real client built in cmd.
type GraphQLQuerier interface {
	QueryWithContext(ctx context.Context, name string, q any, variables map[string]any) error
}

// Fetch loads every timeline event for the given Issue or PR and returns them
// sorted chronologically (stable on equal timestamps, preserving server
// order). The first request reports `totalCount`; the remaining offsets are
// fetched in parallel with bounded concurrency.
//
// The number argument is the Issue or PR number — GitHub uses a single
// numbering space per repository, so `issueOrPullRequest(number:)` resolves
// either form. Errors surface as wrapped errors; a non-existent
// issue/PR is reported as a "not found" error.
func Fetch(ctx context.Context, client GraphQLQuerier, repo Repo, number int) ([]Event, error) {
	if repo.Owner == "" || repo.Name == "" {
		return nil, errors.New("repository owner and name are required")
	}
	if number <= 0 || number > math.MaxInt32 {
		return nil, fmt.Errorf("invalid issue/PR number %d", number)
	}

	firstEvents, totalCount, typename, err := fetchTimelinePage(ctx, client, repo, number, 0)
	if err != nil {
		return nil, err
	}
	if typename == "" {
		return nil, fmt.Errorf("%s/%s#%d not found", repo.Owner, repo.Name, number)
	}

	all := firstEvents
	if totalCount > timelinePageSize {
		// totalCount fits in githubv4.Int (int32), so (totalCount-1)/100 is well
		// within int range on every supported platform.
		extraPageCount := (totalCount - 1) / timelinePageSize
		extraPages := make([][]Event, extraPageCount)

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(timelineMaxConcurrency)
		for i := range extraPageCount {
			slot := i
			offset := (i + 1) * timelinePageSize
			g.Go(func() error {
				events, pageErr := fetchExtraPage(gctx, client, repo, number, offset, typename)
				if pageErr != nil {
					return pageErr
				}
				extraPages[slot] = events
				return nil
			})
		}
		if waitErr := g.Wait(); waitErr != nil {
			return nil, waitErr
		}
		for _, page := range extraPages {
			all = append(all, page...)
		}
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all, nil
}

// fetchExtraPage fetches a non-first page and validates that the Issue/PR
// still exists and has not been converted to the other kind between requests —
// otherwise events would be silently dropped on the merge.
func fetchExtraPage(
	ctx context.Context,
	client GraphQLQuerier,
	repo Repo,
	number, offset int,
	expectedTypename string,
) ([]Event, error) {
	events, _, pageTypename, err := fetchTimelinePage(ctx, client, repo, number, offset)
	if err != nil {
		return nil, err
	}
	if pageTypename == "" {
		return nil, fmt.Errorf(
			"%s/%s#%d disappeared while fetching page (skip=%d)",
			repo.Owner,
			repo.Name,
			number,
			offset,
		)
	}
	if pageTypename != expectedTypename {
		return nil, fmt.Errorf(
			"issueOrPullRequest typename changed from %q to %q (skip=%d)",
			expectedTypename,
			pageTypename,
			offset,
		)
	}
	return events, nil
}

// fetchTimelinePage issues one timelineItems query at the given absolute
// offset. It returns the converted events plus the connection's totalCount
// and the IssueOrPullRequest typename. An empty typename signals that the
// issue/PR does not exist; the caller decides whether that is fatal.
func fetchTimelinePage(
	ctx context.Context,
	client GraphQLQuerier,
	repo Repo,
	number, skip int,
) ([]Event, int, string, error) {
	// Validate inputs locally so the int → int32 narrowing for the GraphQL
	// variables is recognised as bounded by static analysis (gosec G115).
	if number <= 0 || number > math.MaxInt32 {
		return nil, 0, "", fmt.Errorf("invalid issue/PR number %d", number)
	}
	if skip < 0 || skip > math.MaxInt32 {
		return nil, 0, "", fmt.Errorf("invalid skip offset %d", skip)
	}

	var q timelineQuery
	vars := map[string]any{
		"owner":  graphql.String(repo.Owner),
		"name":   graphql.String(repo.Name),
		"number": graphql.Int(number),
		"skip":   graphql.Int(skip),
	}
	if err := client.QueryWithContext(ctx, "TimelineQuery", &q, vars); err != nil {
		return nil, 0, "", fmt.Errorf("timeline query failed (skip=%d): %w", skip, err)
	}
	typename := q.Repository.IssueOrPullRequest.Typename
	switch typename {
	case "":
		return nil, 0, "", nil
	case "PullRequest":
		page := q.Repository.IssueOrPullRequest.PullRequest.TimelineItems
		events := make([]Event, 0, len(page.Nodes))
		for _, n := range page.Nodes {
			events = append(events, dispatchPRNode(n))
		}
		return events, int(page.TotalCount), typename, nil
	case "Issue":
		page := q.Repository.IssueOrPullRequest.Issue.TimelineItems
		events := make([]Event, 0, len(page.Nodes))
		for _, n := range page.Nodes {
			events = append(events, dispatchIssueNode(n))
		}
		return events, int(page.TotalCount), typename, nil
	default:
		return nil, 0, typename, fmt.Errorf("unexpected issueOrPullRequest typename %q", typename)
	}
}

package github

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var client *github.Client

// ErrNotFound is not found error.
var ErrNotFound = errors.New("not found")

// ConnectOption specifies optional parameter to connect github.
type ConnectOption struct {
	AccessToken string
}

// Connect creates github client.
func Connect(op *ConnectOption) {
	var tc *http.Client
	if op != nil && op.AccessToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: op.AccessToken},
		)
		tc = oauth2.NewClient(context.Background(), ts)
	}

	client = github.NewClient(tc)
}

// Disconnect releases github client.
func Disconnect() {
	client = nil
}

// GetOwner gets owner info.
func GetOwner(ctx context.Context, name string) (*github.User, error) {
	owner, res, err := client.Users.Get(ctx, name)
	if res != nil && res.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return owner, nil
}

// LatestRelease fetches latest release of specified repository.
func LatestRelease(ctx context.Context, owner string, name string) (*github.RepositoryRelease, error) {
	release, res, err := client.Repositories.GetLatestRelease(ctx, owner, name)
	if res != nil && res.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return release, nil
}

// LatestCommit fetches latest tag of specified repository.
func LatestCommit(ctx context.Context, owner string, name string) (*github.RepositoryCommit, error) {
	commits, res, err := client.Repositories.ListCommits(ctx, owner, name, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 1,
		},
	})
	if res != nil && res.StatusCode == 404 || len(commits) == 0 {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return commits[0], nil
}

// LatestIssue fetches latest issue of specified repository.
func LatestIssue(ctx context.Context, owner string, name string) (*github.Issue, error) {
	issues, res, err := client.Issues.ListByRepo(ctx, owner, name, &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 1,
		},
	})
	if res != nil && res.StatusCode == 404 || len(issues) == 0 {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return issues[0], nil
}

// LatestPRIssue fetches latest pr of specified repository.
// PR is every pull request is an issue. see https://godoc.org/github.com/google/go-github/github/issues.go?s=780:2350#L16
func LatestPRIssue(ctx context.Context, owner string, name string) (*github.Issue, error) {
	page := 0

	for {
		issues, res, err := client.Issues.ListByRepo(ctx, owner, name, &github.IssueListByRepoOptions{
			State: "all",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 20,
			},
		})
		if res != nil && res.StatusCode == 404 || len(issues) == 0 {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			// if PullRequestLinks is not nil, this is a pull request.
			if issue.PullRequestLinks != nil {
				return issue, nil
			}
		}

		page++
	}
}

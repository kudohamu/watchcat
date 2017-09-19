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

// ConnectOption ...
type ConnectOption struct {
	AccessToken string
}

// Connect ...
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

// Disconnect ...
func Disconnect() {
	client = nil
}

// LatestRelease ...
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

// LatestCommit ...
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

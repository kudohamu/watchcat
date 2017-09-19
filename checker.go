package watchcat

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/kudohamu/watchcat/internal/lmdb"
	version "github.com/mcuadros/go-version"
)

// ReleaseChecker represents checker for latest release.
type ReleaseChecker struct {
	repo      *RepoConfig
	notifiers notifiers
}

// CommitChecker represents checker for latest commit.
type CommitChecker struct {
	repo      *RepoConfig
	notifiers notifiers
}

// Run checks latest release.
func (rc *ReleaseChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  rc.repo.Owner,
		Name:   rc.repo.Name,
		Target: TargetReleases,
	}
	if err := repo.Read(); err != nil {
		rc.notifiers.Error(err)
		return err
	}

	// fetch latest release.
	client := github.NewClient(nil)
	info, res, err := client.Repositories.GetLatestRelease(context.Background(), repo.Owner, repo.Name)
	if res.StatusCode == 404 {
		return nil
	}
	if err != nil {
		rc.notifiers.Error(err)
		return err
	}

	prev := repo.Current
	if repo.Current == "" || version.CompareSimple(repo.Current, info.GetTagName()) < 0 {
		repo.Current = info.GetTagName()
		if err := repo.Write(); err != nil {
			rc.notifiers.Error(err)
			return err
		}

		ni := &NotificationInfo{
			Owner:     repo.Owner,
			AvatarURL: info.Author.GetAvatarURL(),
			RepoName:  repo.Name,
			Current:   repo.Current,
			Prev:      prev,
			Link:      info.GetHTMLURL(),
			Body:      info.GetBody(),
			Target:    repo.Target[:len(repo.Target)-1],
		}
		rc.notifiers.Notify(ni)
	}
	return nil
}

// Run checks latest commit.
func (c *CommitChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetCommits,
	}
	if err := repo.Read(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	// fetch latest commit.
	client := github.NewClient(nil)
	info, _, err := client.Repositories.ListCommits(context.Background(), repo.Owner, repo.Name, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 1,
		},
	})
	if err != nil {
		c.notifiers.Error(err)
		return err
	}

	commit := info[0]
	// has new commit?
	if repo.Current == commit.GetSHA() {
		return nil
	}

	prev := repo.Current
	repo.Current = commit.GetSHA()
	if err := repo.Write(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	ni := &NotificationInfo{
		Owner:     repo.Owner,
		AvatarURL: commit.Author.GetAvatarURL(),
		RepoName:  repo.Name,
		Current:   repo.Current,
		Prev:      prev,
		Link:      commit.GetHTMLURL(),
		Body:      commit.Commit.GetMessage(),
		Target:    repo.Target[:len(repo.Target)-1],
	}
	c.notifiers.Notify(ni)

	return nil
}

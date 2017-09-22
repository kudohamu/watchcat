package watchcat

import (
	"context"
	"strconv"

	"github.com/kudohamu/watchcat/internal/github"
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

// IssueChecker represents checker for latest issue.
type IssueChecker struct {
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

	release, err := github.LatestRelease(context.Background(), repo.Owner, repo.Name)
	if err != nil {
		if err != github.ErrNotFound {
			rc.notifiers.Error(err)
		}
		return err
	}

	// has new release?
	if repo.Current != "" && version.CompareSimple(repo.Current, release.GetTagName()) >= 0 {
		return nil
	}

	prev := repo.Current
	repo.Current = release.GetTagName()
	if err := repo.Write(); err != nil {
		rc.notifiers.Error(err)
		return err
	}

	ni := &NotificationInfo{
		Owner:     repo.Owner,
		AvatarURL: rc.repo.avatarURL,
		RepoName:  repo.Name,
		Current:   repo.Current,
		Prev:      prev,
		Link:      release.GetHTMLURL(),
		Body:      release.GetBody(),
		Target:    repo.Target[:len(repo.Target)-1],
	}
	rc.notifiers.Notify(ni)
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

	commit, err := github.LatestCommit(context.Background(), repo.Owner, repo.Name)
	if err != nil {
		if err != github.ErrNotFound {
			c.notifiers.Error(err)
		}
		return err
	}

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
		AvatarURL: c.repo.avatarURL,
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

// Run checks latest issue.
func (c *IssueChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetIssues,
	}
	if err := repo.Read(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	issue, err := github.LatestIssue(context.Background(), repo.Owner, repo.Name)
	if err != nil {
		if err != github.ErrNotFound {
			c.notifiers.Error(err)
		}
		return err
	}

	// has new issue?
	current, err := strconv.Atoi(repo.Current)
	if err == nil && current >= issue.GetID() {
		return nil
	}

	prev := repo.Current
	repo.Current = strconv.Itoa(issue.GetID())
	if err := repo.Write(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	ni := &NotificationInfo{
		Owner:     repo.Owner,
		AvatarURL: c.repo.avatarURL,
		RepoName:  repo.Name,
		Current:   issue.GetTitle(),
		Prev:      prev,
		Link:      issue.GetHTMLURL(),
		Body:      issue.GetBody(),
		Target:    repo.Target[:len(repo.Target)-1],
	}
	c.notifiers.Notify(ni)

	return nil
}

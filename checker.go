package watchcat

import (
	"context"
	"fmt"
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

// PRChecker repositories checker for latest pr.
type PRChecker struct {
	repo      *RepoConfig
	notifiers notifiers
}

// TagChecker repositories checker for latest tag.
type TagChecker struct {
	repo      *RepoConfig
	notifiers notifiers
}

// Run checks latest release.
func (rc *ReleaseChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  rc.repo.Owner,
		Name:   rc.repo.Name,
		Target: TargetRelease,
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
		Title:     release.GetTagName(),
		Body:      release.GetBody(),
		Target:    repo.Target,
	}
	rc.notifiers.Notify(ni)
	return nil
}

// Run checks latest commit.
func (c *CommitChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetCommit,
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
		Title:     commit.GetSHA(),
		Body:      commit.Commit.GetMessage(),
		Target:    repo.Target,
	}
	c.notifiers.Notify(ni)

	return nil
}

// Run checks latest issue.
func (c *IssueChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetIssue,
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
	current, err := strconv.ParseInt(repo.Current, 10, 64)
	if err == nil && current >= issue.GetID() {
		return nil
	}

	prev := repo.Current
	repo.Current = strconv.FormatInt(issue.GetID(), 10)
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
		Link:      issue.GetHTMLURL(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		Target:    repo.Target,
	}
	c.notifiers.Notify(ni)

	return nil
}

// Run checks latest pr.
func (c *PRChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetPR,
	}
	if err := repo.Read(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	pr, err := github.LatestPRIssue(context.Background(), repo.Owner, repo.Name)
	if err != nil {
		if err != github.ErrNotFound {
			c.notifiers.Error(err)
		}
		return err
	}

	// has new pr?
	current, err := strconv.ParseInt(repo.Current, 10, 64)
	if err == nil && current >= pr.GetID() {
		return nil
	}

	prev := repo.Current
	repo.Current = strconv.FormatInt(pr.GetID(), 10)
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
		Link:      pr.PullRequestLinks.GetHTMLURL(),
		Title:     pr.GetTitle(),
		Body:      pr.GetBody(),
		Target:    repo.Target,
	}
	c.notifiers.Notify(ni)

	return nil
}

// Run checks latest tag.
func (c *TagChecker) Run() error {
	repo := &lmdb.Repo{
		Owner:  c.repo.Owner,
		Name:   c.repo.Name,
		Target: TargetTag,
	}
	if err := repo.Read(); err != nil {
		c.notifiers.Error(err)
		return err
	}

	tag, err := github.LatestTag(context.Background(), repo.Owner, repo.Name)
	if err != nil {
		if err != github.ErrNotFound {
			c.notifiers.Error(err)
		}
		return err
	}

	// has new tag?
	if repo.Current != "" && version.CompareSimple(repo.Current, tag.GetName()) >= 0 {
		return nil
	}

	prev := repo.Current
	repo.Current = tag.GetName()
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
		Link:      fmt.Sprintf("https://github.com/%s/%s/tags", repo.Owner, repo.Name),
		Title:     tag.GetName(),
		Body:      "",
		Target:    repo.Target,
	}
	c.notifiers.Notify(ni)

	return nil
}

package watchcat

import (
	"context"

	"github.com/google/go-github/github"
	version "github.com/mcuadros/go-version"
)

// ReleaseChecker represents checker for latest release.
type ReleaseChecker struct {
	repo      *RepoConfig
	notifiers notifiers
}

// Run checks latest release.
func (rc *ReleaseChecker) Run() error {
	repo := &Repo{
		Owner:    rc.repo.Owner,
		RepoName: rc.repo.Name,
		Target:   TargetReleases,
	}
	if err := repo.Read(); err != nil {
		rc.notifiers.Error(err)
		return err
	}

	// fetch latest release.
	client := github.NewClient(nil)
	info, _, err := client.Repositories.GetLatestRelease(context.Background(), repo.Owner, repo.RepoName)
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
			RepoName:  repo.RepoName,
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

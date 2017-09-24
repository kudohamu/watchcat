// Package watchcat is library for watching github activities.
package watchcat

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kudohamu/petelgeuse"
	"github.com/kudohamu/watchcat/internal/github"
	"github.com/kudohamu/watchcat/internal/lmdb"
	homedir "github.com/mitchellh/go-homedir"
)

// watching targets.
const (
	TargetRelease = "release"
	TargetCommit  = "commit"
	TargetIssue   = "issue"
	TargetPR      = "pr"
	TargetTag     = "tag"
)

// Watcher represents watcher for github some activities.
type Watcher struct {
	configPath  string
	ticker      *time.Ticker
	notifiers   notifiers
	worker      *petelgeuse.Manager
	interval    string
	accessToken string
}

// Config represents cofiguration of watching targets.
type Config struct {
	Repos []*RepoConfig `toml:"repos"`
}

// RepoConfig represents target repository to watch.
type RepoConfig struct {
	Owner     string   `toml:"owner"`
	Name      string   `toml:"name"`
	Targets   []string `toml:"targets"`
	avatarURL string
}

// New creates new watchcat instance.
func New(confingPath string, interval string, accessToken string) *Watcher {
	worker := petelgeuse.New(&petelgeuse.Option{
		WorkerSize: 10,
		QueueSize:  1000,
	})

	return &Watcher{
		configPath:  confingPath,
		worker:      worker,
		notifiers:   notifiers{},
		interval:    interval,
		accessToken: accessToken,
	}
}

// Watch starts to watch repositories.
func (w *Watcher) Watch() error {
	w.worker.Start()
	if err := lmdb.Connect(); err != nil {
		return err
	}
	if w.accessToken == "" {
		github.Connect(nil)
	} else {
		github.Connect(&github.ConnectOption{
			AccessToken: w.accessToken,
		})
	}

	defer func() {
		github.Disconnect()
		lmdb.Disconnect()
		w.worker.StopImmediately()
	}()

	config, err := readConfig(w.configPath)
	if err != nil {
		return err
	}

	w.check(config.Repos)

	interval, err := time.ParseDuration(w.interval)
	if err != nil {
		return fmt.Errorf("invalid interval: %s", w.interval)
	}
	w.ticker = time.NewTicker(interval)
	defer w.ticker.Stop()

	stopC := make(chan os.Signal, 1)
	signal.Notify(stopC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case <-w.ticker.C:
			config, err := readConfig(w.configPath)
			if err != nil {
				continue
			}
			w.check(config.Repos)
		case <-stopC:
			return nil
		}
	}
}

// AddNotifier adds notifier.
func (w *Watcher) AddNotifier(n Notifier) {
	w.notifiers = append(w.notifiers, n)
}

func (w *Watcher) check(repos []*RepoConfig) {
	for _, repo := range repos {
		avatarURL, err := fetchAvatarURL(repo.Owner)
		if err == nil {
			repo.avatarURL = avatarURL
		}

		for _, target := range repo.Targets {
			switch target {
			case TargetRelease:
				w.worker.Add(&ReleaseChecker{
					repo:      repo,
					notifiers: w.notifiers,
				})
			case TargetCommit:
				w.worker.Add(&CommitChecker{
					repo:      repo,
					notifiers: w.notifiers,
				})
			case TargetIssue:
				w.worker.Add(&IssueChecker{
					repo:      repo,
					notifiers: w.notifiers,
				})
			case TargetPR:
				w.worker.Add(&PRChecker{
					repo:      repo,
					notifiers: w.notifiers,
				})
			case TargetTag:
				w.worker.Add(&TagChecker{
					repo:      repo,
					notifiers: w.notifiers,
				})
			}
		}
	}
}

func readConfig(path string) (*Config, error) {
	if strings.HasPrefix(path, "https://") {
		return readConfigFromURL(path)
	} else if strings.HasPrefix(path, "http://") {
		return readConfigFromURL(path)
	} else if strings.HasPrefix(path, "file://") {
		return readConfigFromFilePath(path)
	}
	return nil, fmt.Errorf("invalid configration type")
}

func readConfigFromURL(url string) (*Config, error) {
	client := http.Client{Timeout: time.Duration(20 * time.Second)}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("could not read url: %s", url)
	}

	var config Config
	if _, err := toml.DecodeReader(res.Body, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func readConfigFromFilePath(fpath string) (*Config, error) {
	var config Config
	fp := strings.Replace(fpath, "file://", "", 1)
	if fp[:2] == "~/" {
		hd, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		fp = filepath.Join(hd, fp[2:])
	}
	if _, err := toml.DecodeFile(fp, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func fetchAvatarURL(ownerName string) (string, error) {
	cache := &lmdb.Owner{
		Name: ownerName,
	}
	// expiration time of cache is one day.
	if err := cache.Read(); err == nil && time.Now().Before(cache.CachedAt.Add(24*time.Hour)) {
		return cache.AvatarURL, nil
	}

	owner, err := github.GetOwner(context.Background(), ownerName)
	if err != nil {
		return "", err
	}
	cache = &lmdb.Owner{
		Name:      ownerName,
		AvatarURL: owner.GetAvatarURL(),
		CachedAt:  time.Now(),
	}
	cache.Write()

	return owner.GetAvatarURL(), nil
}

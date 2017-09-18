// Package watchcat is library for watching github activities.
package watchcat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kudohamu/petelgeuse"
)

// watching targets.
const (
	TargetReleases = "releases"
)

// Watcher represents watcher for github some activities.
type Watcher struct {
	configPath string
	ticker     *time.Ticker
	notifiers  notifiers
	worker     *petelgeuse.Manager
	interval   string
}

// Config represents cofiguration of watching targets.
type Config struct {
	Repos []*RepoConfig `json:"repos"`
}

// RepoConfig represents target repository to watch.
type RepoConfig struct {
	Owner   string   `json:"owner"`
	Name    string   `json:"name"`
	Targets []string `json:"targets"`
}

// New creates new watchcat instance.
func New(confingPath string, interval string) *Watcher {
	worker := petelgeuse.New(&petelgeuse.Option{
		WorkerSize: 10,
		QueueSize:  1000,
	})

	return &Watcher{
		configPath: confingPath,
		worker:     worker,
		notifiers:  notifiers{},
		interval:   interval,
	}
}

// Watch starts to watch repositories.
func (w *Watcher) Watch() error {
	w.worker.Start()
	if err := connect(); err != nil {
		return err
	}
	defer func() {
		disconnect()
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
		for _, target := range repo.Targets {
			switch target {
			case TargetReleases:
				w.worker.Add(&ReleaseChecker{
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
	if err := json.NewDecoder(res.Body).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

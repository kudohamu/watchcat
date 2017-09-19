package watchcat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// Notifier is interface of notification.
type Notifier interface {
	Notify(*NotificationInfo) error
	Error(error) error
}

// NotificationInfo is notification payload.
type NotificationInfo struct {
	Owner     string
	AvatarURL string
	RepoName  string
	Target    string
	Current   string
	Prev      string
	Body      string
	Link      string
}

// StdNotifier handles notifications to stdout.
type StdNotifier struct{}

// SlackNotifier handles notifications to slack.
type SlackNotifier struct {
	WebhookURL string
}

type notifiers []Notifier

var notificationColors = map[string]string{
	"release": "#3EBB3E",
	"commit":  "#CBBE34",
	"error":   "danger",
}

// Notify notifies to stdout.
func (*StdNotifier) Notify(info *NotificationInfo) error {
	log.Printf("(%s/%s) new %s: %s\n", info.Owner, info.RepoName, info.Target, info.Link)

	return nil
}

// Error notifies error to stdout.
func (*StdNotifier) Error(err error) error {
	log.Printf("error: %s\n", err.Error())

	return nil
}

// Notify notifies to slack.
func (n *SlackNotifier) Notify(info *NotificationInfo) error {
	data, err := json.Marshal(map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fallback":    fmt.Sprintf("(%s/%s) new %s: %s", info.Owner, info.RepoName, info.Target, info.Current),
				"author_name": fmt.Sprintf("%s/%s", info.Owner, info.RepoName),
				"author_link": fmt.Sprintf("https://github.com/%s/%s", info.Owner, info.RepoName),
				"author_icon": info.AvatarURL,
				"title":       fmt.Sprintf("new %s: %s", info.Target, info.Current),
				"title_link":  info.Link,
				"text":        info.Body,
				"color":       notificationColors[info.Target],
				"mrkdwn_in":   []string{"text"},
			},
		},
	})
	if err != nil {
		return err
	}

	res, err := http.PostForm(n.WebhookURL, url.Values{"payload": {string(data)}})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify")
	}

	return nil
}

// Error notifies error to slack.
func (n *SlackNotifier) Error(err error) error {
	data, err := json.Marshal(map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fallback": "failed",
				"title":    "failed",
				"text":     err.Error(),
				"color":    notificationColors["error"],
			},
		},
	})
	if err != nil {
		return err
	}

	res, err := http.PostForm(n.WebhookURL, url.Values{"payload": {string(data)}})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify")
	}

	return nil
}

// Notify fires all notifiers' Notify.
func (ns notifiers) Notify(info *NotificationInfo) error {
	for _, n := range ns {
		n.Notify(info)
	}
	return nil
}

// Error fires all notifiers' Error.
func (ns notifiers) Error(err error) error {
	for _, n := range ns {
		n.Error(err)
	}
	return nil
}

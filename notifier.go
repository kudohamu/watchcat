package watchcat

import "log"

// Notifier is interface of notification.
type Notifier interface {
	Notify(*NotificationInfo) error
	Error(error) error
}

// NotificationInfo is notification payload.
type NotificationInfo struct {
	Owner    string
	RepoName string
	Target   string
	Current  string
	Link     string
}

// StdNotifier handles notifications to stdout.
type StdNotifier struct{}

type notifiers []Notifier

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

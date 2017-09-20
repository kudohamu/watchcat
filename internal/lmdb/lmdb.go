package lmdb

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var conn *bolt.DB

// Repo is the LMDB store to store current repository state.
type Repo struct {
	Owner   string
	Name    string
	Target  string
	Current string
}

// Owner is the LMDB store to cache owner's avatar.
type Owner struct {
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatarUrl"`
	CachedAt  time.Time `json:"cachedAt"`
}

const timeFormat = "2006-01-02 15:04:05 -0700"

var bktOwer = []byte("owner")
var bktRepo = []byte("repo")

// Connect connects to lmdb.
func Connect() error {
	var err error
	usr, err := user.Current()
	if err != nil {
		return err
	}
	dir := path.Join(usr.HomeDir, ".config", "watchcat")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	p := path.Join(dir, "watchcat.db")

	conn, err = bolt.Open(p, 0755, nil)
	if err != nil {
		return err
	}

	// create buckets
	return conn.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bktRepo); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bktOwer); err != nil {
			return err
		}

		return nil
	})
}

// Disconnect disconnects from lmdb.
func Disconnect() error {
	if err := conn.Close(); err != nil {
		return err
	}
	conn = nil

	return nil
}

// Read reads stored current target information of repository.
func (repo *Repo) Read() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bktRepo)

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.Name, repo.Target)
		value := bkt.Get([]byte(key))
		repo.Current = string(value)

		return nil
	})
}

// Write stores target information of repository.
func (repo *Repo) Write() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bktRepo)

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.Name, repo.Target)
		if err := bkt.Put([]byte(key), []byte(repo.Current)); err != nil {
			return err
		}

		return nil
	})
}

func (o *Owner) Read() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bktOwer)

		avatarURL := bkt.Get([]byte(fmt.Sprintf("%s/%s", o.Name, "avatar")))
		o.AvatarURL = string(avatarURL)

		cachedAt := bkt.Get([]byte(fmt.Sprintf("%s/%s", o.Name, "cachedAt")))
		if len(cachedAt) == 0 {
			return errors.New("not found")
		}
		cAt, err := time.Parse(timeFormat, string(cachedAt))
		if err != nil {
			return err
		}
		o.CachedAt = cAt

		return nil
	})
}

func (o *Owner) Write() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bktOwer)

		if err := bkt.Put([]byte(fmt.Sprintf("%s/%s", o.Name, "avatar")), []byte(o.AvatarURL)); err != nil {
			return err
		}
		if err := bkt.Put([]byte(fmt.Sprintf("%s/%s", o.Name, "cachedAt")), []byte(o.CachedAt.Format(timeFormat))); err != nil {
			return err
		}

		return nil
	})
}

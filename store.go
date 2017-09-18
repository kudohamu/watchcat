package watchcat

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/boltdb/bolt"
)

var conn *bolt.DB

// Repo represents stored current target information of repository.
type Repo struct {
	Owner    string
	RepoName string
	Target   string
	Current  string
}

func connect() error {
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

	return nil
}

func disconnect() error {
	if err := conn.Close(); err != nil {
		return err
	}
	conn = nil

	return nil
}

// Read reads stored current target information of repository.
func (repo *Repo) Read() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte("repo"))
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.RepoName, repo.Target)
		value := bkt.Get([]byte(key))
		repo.Current = string(value)

		return nil
	})
}

// Write stores target information of repository.
func (repo *Repo) Write() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("repo"))

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.RepoName, repo.Target)
		if err := bkt.Put([]byte(key), []byte(repo.Current)); err != nil {
			return err
		}

		return nil
	})
}

package lmdb

import (
	"fmt"
	"os"
	"os/user"
	"path"

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

	return nil
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
		bkt, err := tx.CreateBucketIfNotExists([]byte("repo"))
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.Name, repo.Target)
		value := bkt.Get([]byte(key))
		repo.Current = string(value)

		return nil
	})
}

// Write stores target information of repository.
func (repo *Repo) Write() error {
	return conn.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("repo"))

		key := fmt.Sprintf("%s/%s/%s", repo.Owner, repo.Name, repo.Target)
		if err := bkt.Put([]byte(key), []byte(repo.Current)); err != nil {
			return err
		}

		return nil
	})
}

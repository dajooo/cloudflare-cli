package db

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"go.etcd.io/bbolt"
)

var (
	db   *bbolt.DB
	once sync.Once
)

var (
	ConfigBucket = []byte("config")
	CacheBucket  = []byte("cache")
)

func Open() (*bbolt.DB, error) {
	var err error
	once.Do(func() {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			err = homeErr
			return
		}

		dbPath := filepath.Join(home, ".cloudflare-cli")
		if err = os.MkdirAll(dbPath, 0700); err != nil {
			return
		}

		database, openErr := bbolt.Open(filepath.Join(dbPath, "cf.db"), 0600, nil)
		if openErr != nil {
			err = openErr
			return
		}
		db = database

		err = db.Update(func(tx *bbolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists(ConfigBucket); err != nil {
				return err
			}
			if _, err := tx.CreateBucketIfNotExists(CacheBucket); err != nil {
				return err
			}
			return nil
		})
	})

	return db, err
}

func Get(bucket, key []byte) ([]byte, error) {
	database, err := Open()
	if err != nil {
		return nil, err
	}

	var value []byte
	err = database.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket not found")
		}
		// b.Get returns a direct reference; copy it to a new slice.
		val := b.Get(key)
		if val != nil {
			value = append([]byte(nil), val...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return value, nil
}

func Set(bucket, key, value []byte) error {
	database, err := Open()
	if err != nil {
		return err
	}

	return database.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket not found")
		}
		return b.Put(key, value)
	})
}

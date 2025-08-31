package db

import (
	"encoding/json"
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
	ConfigBucket    = []byte("config")
	CacheBucket     = []byte("cache")
	CacheTagsBucket = []byte("cache_tags")
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
			_, err := tx.CreateBucketIfNotExists(ConfigBucket)
			if err != nil {
				return err
			}
			_, err = tx.CreateBucketIfNotExists(CacheBucket)
			if err != nil {
				return err
			}
			_, err = tx.CreateBucketIfNotExists(CacheTagsBucket)
			return err
		})
	})
	return db, err
}

func Set(bucket, key, value []byte) error {
	database, err := Open()
	if err != nil {
		return err
	}
	return database.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if value == nil {
			return b.Delete(key)
		}
		return b.Put(key, value)
	})
}

func Get(bucket, key []byte) ([]byte, error) {
	database, err := Open()
	if err != nil {
		return nil, err
	}
	var value []byte
	err = database.View(func(tx *bbolt.Tx) error {
		val := tx.Bucket(bucket).Get(key)
		if val != nil {
			value = append([]byte(nil), val...)
		}
		return nil
	})
	return value, err
}

func AddTagsToKey(key string, tags []string) error {
	db, err := Open()
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CacheTagsBucket)
		for _, tag := range tags {
			var keys []string
			existingKeysBytes := b.Get([]byte(tag))
			if existingKeysBytes != nil {
				if err := json.Unmarshal(existingKeysBytes, &keys); err != nil {
					return err
				}
			}
			keys = append(keys, key)
			newKeysBytes, err := json.Marshal(keys)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(tag), newKeysBytes); err != nil {
				return err
			}
		}
		return nil
	})
}

func InvalidateTags(tags []string) error {
	db, err := Open()
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		tagsBucket := tx.Bucket(CacheTagsBucket)
		cacheBucket := tx.Bucket(CacheBucket)
		keysToDelete := make(map[string]struct{})

		for _, tag := range tags {
			keysBytes := tagsBucket.Get([]byte(tag))
			if keysBytes != nil {
				var keys []string
				if err := json.Unmarshal(keysBytes, &keys); err == nil {
					for _, key := range keys {
						keysToDelete[key] = struct{}{}
					}
				}
				if err := tagsBucket.Delete([]byte(tag)); err != nil {
					return err
				}
			}
		}

		for key := range keysToDelete {
			if err := cacheBucket.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	})
}

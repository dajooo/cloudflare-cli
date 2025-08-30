package cloudflare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type CacheData struct {
	Identifiers map[string]CacheEntry `json:"identifiers"`
}

var (
	cache     *CacheData
	cachePath string
	once      sync.Once
	mu        sync.Mutex
)

func initCache() {
	home, err := os.UserHomeDir()
	if err != nil {
		cache = &CacheData{Identifiers: make(map[string]CacheEntry)}
		return
	}
	cachePath = filepath.Join(home, ".cloudflare-cli/cache.json")
	loadCache()
}

func loadCache() {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(cachePath)
	if err != nil {
		cache = &CacheData{Identifiers: make(map[string]CacheEntry)}
		return
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		cache = &CacheData{Identifiers: make(map[string]CacheEntry)}
	}
}

func saveCache() error {
	mu.Lock()
	defer mu.Unlock()

	if cachePath == "" {
		return nil
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0600)
}

func GetID(key string) (string, bool) {
	once.Do(initCache)

	mu.Lock()
	defer mu.Unlock()

	entry, found := cache.Identifiers[key]
	if !found {
		return "", false
	}
	return entry.ID, true
}

func SetID(key, id string) {
	once.Do(initCache)

	mu.Lock()
	if cache.Identifiers == nil {
		cache.Identifiers = make(map[string]CacheEntry)
	}
	cache.Identifiers[key] = CacheEntry{
		ID:        id,
		Timestamp: time.Now(),
	}
	mu.Unlock()

	_ = saveCache()
}

func ZoneCacheKey(zoneName string) string {
	return fmt.Sprintf("zone:%s", zoneName)
}

func DNSRecordCacheKey(zoneID, recordName string) string {
	return fmt.Sprintf("dns:%s:%s", zoneID, recordName)
}

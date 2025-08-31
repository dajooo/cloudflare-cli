package cloudflare

import (
	"fmt"

	"dario.lol/cf/internal/db"
)

func GetID(key string) (string, bool) {
	val, err := db.Get(db.CacheBucket, []byte(key))
	if err != nil || val == nil {
		return "", false
	}
	return string(val), true
}

func SetID(key, id string) {
	// Errors are ignored for cache writes to not disrupt user operations.
	_ = db.Set(db.CacheBucket, []byte(key), []byte(id))
}

func ZoneCacheKey(zoneIdentifier string) string {
	return fmt.Sprintf("zone:%s", zoneIdentifier)
}

func DNSRecordCacheKey(zoneID, recordName string) string {
	return fmt.Sprintf("dns:%s:%s", zoneID, recordName)
}

func DNSRecordCacheKeyByID(recordID string) string {
	return fmt.Sprintf("dns_by_id:%s", recordID)
}

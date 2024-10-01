package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// Cache type
type Cache struct {
	Cache   *ristretto.Cache[string, interface{}]
	TTL     time.Duration
	Enabled bool
}

var CacheInstance Cache

func InitCache(enabled bool, TTLString string) Cache {
	var cacheInstance Cache
	var cacheErr error

	if TTLString == "" {
		TTLString = "1m"
	}

	cacheInstance.Enabled = enabled

	cacheInstance.TTL, cacheErr = time.ParseDuration(TTLString)
	if cacheErr != nil {
		panic(cacheErr)
	}

	cacheInstance.Cache, cacheErr = ristretto.NewCache(&ristretto.Config[string, interface{}]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 28, // maximum cost of cache (256mb).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if cacheErr != nil {
		panic(cacheErr)
	}

	return cacheInstance
}

package libs

import (
	"time"

	"log/slog"

	"github.com/wasilak/cloudpile/cache"
)

func Runner() {

	ticker := time.NewTicker(cache.CacheInstance.TTL)

	slog.Debug("Initial cache refresh...")

	Run([]string{}, cache.CacheInstance, true)

	slog.Debug("Cache refresh done", "next_in", cache.CacheInstance.TTL)

	go func() {
		for range ticker.C {
			slog.Debug("Refreshing cache...")
			Run([]string{}, cache.CacheInstance, true)
			slog.Debug("Cache refresh done", "next_in", cache.CacheInstance.TTL)
		}
	}()
}

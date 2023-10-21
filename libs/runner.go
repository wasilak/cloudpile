package libs

import (
	"context"
	"time"

	"log/slog"

	"github.com/wasilak/cloudpile/cache"
)

func Runner(ctx context.Context) {

	ticker := time.NewTicker(cache.CacheInstance.GetTTL(ctx))

	slog.Debug("Initial cache refresh...")

	Run([]string{}, cache.CacheInstance, true)

	slog.Debug("Cache refresh done", "next_in", cache.CacheInstance.GetTTL(ctx))

	go func() {
		for range ticker.C {
			Run([]string{}, cache.CacheInstance, true)
		}
	}()
}

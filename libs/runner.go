package libs

import (
	"context"
	"time"

	"log/slog"
)

func Runner(ctx context.Context) {

	ticker := time.NewTicker(CacheInstance.GetConfig().TTL)

	slog.Debug("Initial cache refresh...")

	Run(ctx, []string{}, true)

	slog.Debug("Cache refresh done", "next_in", CacheInstance.GetConfig().TTL)

	go func() {
		for range ticker.C {
			Run(ctx, []string{}, true)
		}
	}()
}

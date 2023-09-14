package libs

import (
	"time"

	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/cache"
	"golang.org/x/exp/slog"
)

func Runner() {

	cache.CacheInstance = cache.InitCache(viper.GetBool("cache.enabled"), viper.GetString("cache.TTL"))

	ticker := time.NewTicker(cache.CacheInstance.TTL)

	slog.Debug("Initial cache refresh...")

	Run([]string{}, cache.CacheInstance, true)

	slog.Debug("Cache refresh done")

	go func() {

		for range ticker.C {
			slog.Debug("Refreshing cache...")
			Run([]string{}, cache.CacheInstance, true)
			slog.Debug("Cache refresh done")
		}
	}()

	defer ticker.Stop()
}

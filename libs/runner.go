package libs

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func Runner() {

	CacheInstance = InitCache(viper.GetBool("cache.enabled"), viper.GetString("cache.TTL"))

	ticker := time.NewTicker(CacheInstance.TTL)

	slog.Debug("Initial cache refresh...")

	Describe(viper.GetStringSlice("aws.regions"), []string{}, viper.GetStringSlice("aws.iam_role_arn"), viper.GetStringMapString("aws.account_aliasses"), CacheInstance, true)

	slog.Debug("Cache refresh done")

	go func() {

		for range ticker.C {
			slog.Debug("Refreshing cache...")
			Describe(viper.GetStringSlice("aws.regions"), []string{}, viper.GetStringSlice("aws.iam_role_arn"), viper.GetStringMapString("aws.account_aliasses"), CacheInstance, true)
			slog.Debug("Cache refresh done")
		}
	}()

	defer ticker.Stop()
}

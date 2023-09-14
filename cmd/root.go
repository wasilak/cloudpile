package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/libs"
	"github.com/wasilak/cloudpile/web"
	"github.com/wasilak/loggergo"
)

var (
	rootCmd = &cobra.Command{
		Use:   libs.AppName,
		Short: "Cross account AWS resources directory",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(ctx)
		},
		Run: func(cmd *cobra.Command, args []string) {
			loggerConfig := loggergo.LoggerGoConfig{
				Level:  viper.GetString("loglevel"),
				Format: viper.GetString("logformat"),
			}

			_, err := loggergo.LoggerInit(loggerConfig)
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}

			if viper.GetBool("cache.enabled") {
				cache.CacheInstance = cache.InitCache(viper.GetBool("cache.enabled"), viper.GetString("cache.TTL"))
				libs.Runner()
			}

			web.Web()
		},
	}
	ctx = context.Background()
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVar(&libs.CfgFile, "config", "", "config file (default is $HOME/."+libs.AppName+"/config.yml)")
	rootCmd.PersistentFlags().StringVar(&libs.Listen, "listen", "127.0.0.1:3000", "listen address")

	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))

	cobra.OnInitialize(libs.InitConfig)

	rootCmd.AddCommand(versionCmd)
}

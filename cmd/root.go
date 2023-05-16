package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/libs"
	"golang.org/x/exp/slog"
)

var (
	rootCmd = &cobra.Command{
		Use:   libs.AppName,
		Short: "Cross account AWS resources directory",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(ctx)
		},
		Run: func(cmd *cobra.Command, args []string) {
			libs.InitLogging(viper.GetString("loglevel"), viper.GetString("logformat"))

			slog.Debug(fmt.Sprintf("%+v", viper.AllSettings()))

			if viper.GetBool("cache.enabled") {
				libs.Runner()
			}

			libs.Web()
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
	cobra.OnInitialize(libs.InitConfig)

	rootCmd.PersistentFlags().StringVar(&libs.CfgFile, "config", "", "config file (default is $HOME/."+libs.AppName+"/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&libs.CacheEnabled, "cacheEnabled", false, "cache enabled")
	rootCmd.PersistentFlags().StringVar(&libs.Listen, "listen", "127.0.0.1:3000", "listen address")

	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("cacheEnabled", rootCmd.PersistentFlags().Lookup("cacheEnabled"))

	rootCmd.AddCommand(versionCmd)
}

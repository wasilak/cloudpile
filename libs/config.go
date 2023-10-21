package libs

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	AppName      = "cloudpile"
	CfgFile      string
	Listen       string
	CacheEnabled bool
	AWSConfigs   []AWSConfig
)

type AWSConfig struct {
	Type         string   `mapstructure:"type"`
	IAMRoleARN   string   `mapstructure:"iam_role_arn"`
	Profile      string   `mapstructure:"profile"`
	AccountAlias string   `mapstructure:"account_alias"`
	Regions      []string `mapstructure:"regions"`
	Resources    []string `mapstructure:"resources"`
}

func InitConfig() {
	godotenv.Load()

	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")

	viper.SetEnvPrefix(AppName)

	viper.SetDefault("logformat", "plain")

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.SetConfigType("yaml")
		viper.SetConfigName(AppName)
		viper.AddConfigPath(home)
		viper.AddConfigPath("./")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("%+v\n", err)
	}

	viper.UnmarshalKey("aws", &AWSConfigs)

	viper.SetDefault("cache_type", "memory")
	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("cache_expire", "1h")

	if strings.ToLower(viper.GetString("loglevel")) == "debug" {
		log.Printf("%+v", viper.AllSettings())
		log.Printf("%+v", AWSConfigs)
	}
}

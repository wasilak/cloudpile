package main

import (
	"embed"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/libs"
)

var err error

//go:embed views
var views embed.FS

//go:embed assets/*
var assets embed.FS

var cacheInstance libs.Cache

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func getEmbededViews() fs.FS {
	fsys, err := fs.Sub(views, "views")
	if err != nil {
		panic(err)
	}

	return fsys
}

func getEmbededAssets() http.FileSystem {
	fsys, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}

func main() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")

	// using standard library "flag" package
	flag.Bool("verbose", false, "verbose")
	flag.Bool("cacheEnabled", false, "cache enabled")
	flag.String("listen", "127.0.0.1:3000", "listen address")
	flag.String("config", "./", "path to cloudpile.yml")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("CLOUDPILE")
	viper.AutomaticEnv()

	viper.SetConfigName("cloudpile")               // name of config file (without extension)
	viper.SetConfigType("yaml")                    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(viper.GetString("config")) // path to look for the config file in
	viperErr := viper.ReadInConfig()               // Find and read the config file

	if viperErr != nil { // Handle errors reading the config file
		log.Fatal(viperErr)
		panic(viperErr)
	}

	libs.AwsRegions = viper.GetStringSlice("aws.regions")
	libs.AwsRoles = viper.GetStringSlice("aws.iam_role_arn")
	libs.AccountAliasses = viper.GetStringMapString("aws.account_aliasses")

	if err != nil { // Handle errors reading the config file
		log.Fatal(err)
		panic(err)
	}

	log.Debug(viper.AllSettings())

	if viper.GetBool("debug") {
		log.SetLevel(log.DEBUG)
	}

	if viper.GetBool("cache.enabled") {

		cacheInstance = libs.InitCache(viper.GetBool("cache.enabled"), viper.GetString("cache.TTL"))

		libs.CacheInstance = cacheInstance

		ticker := time.NewTicker(cacheInstance.TTL)

		log.Debug("Initial cache refresh...")

		libs.Describe(libs.AwsRegions, []string{}, libs.AwsRoles, libs.AccountAliasses, cacheInstance, true)

		log.Debug("Cache refresh done")

		go func() {

			for range ticker.C {
				log.Debug("Refreshing cache...")
				libs.Describe(libs.AwsRegions, []string{}, libs.AwsRoles, libs.AccountAliasses, cacheInstance, true)
				log.Debug("Cache refresh done")
			}
		}()

		defer ticker.Stop()

	}

	e := echo.New()

	e.Use(middleware.Gzip())

	e.HideBanner = true

	if viper.GetBool("verbose") {
		e.Logger.SetLevel(log.DEBUG)
	}

	e.Debug = viper.GetBool("debug")

	t := &Template{
		templates: template.Must(template.ParseFS(getEmbededViews(), "*.html")),
	}

	e.Renderer = t

	e.Use(middleware.Logger())

	// Enable metrics middleware
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	assetHandler := http.FileServer(getEmbededAssets())
	e.GET("/public/assets/*", echo.WrapHandler(http.StripPrefix("/public/assets/", assetHandler)))

	e.GET("/", libs.MainRoute)
	e.GET("/list", libs.ListRoute)
	e.GET("/api/list", libs.ApiListRoute)
	e.GET("/search", libs.SearchRoute)
	e.GET("/search/", libs.SearchRoute)
	e.GET("/api/search/", libs.ApiSearchRoute)
	e.GET("/api/search/:id", libs.ApiSearchRoute)
	e.GET("/api/config/", libs.ApiConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}

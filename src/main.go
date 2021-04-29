package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/libs"
)

var awsRegions []string
var awsRoles []string
var accountAliasses map[string]string
var sess *session.Session
var verbose bool

//go:embed views
var views embed.FS

//go:embed assets/*
var assets embed.FS

func mainRoute(c *fiber.Ctx) error {
	return c.Render("views/main", fiber.Map{})
}

func searchRoute(c *fiber.Ctx) error {
	ids := strings.Split(strings.Replace(c.Query("id"), "%2C", ",", -1), ",")

	ids = libs.Deduplicate(ids)

	items := libs.Describe(awsRegions, ids, awsRoles, sess, accountAliasses, verbose)

	if verbose {
		log.Println(c.Query("id"))
		log.Println(ids)
		log.Println(items)
	}

	return c.Render("views/search", fiber.Map{
		"Items": items,
		"IDs":   strings.Join(ids, ","),
	})
}

func apiSearchRoute(c *fiber.Ctx) error {
	ids := strings.Split(strings.Replace(c.Params("id"), "%2C", ",", -1), ",")

	items := libs.Describe(awsRegions, ids, awsRoles, sess, accountAliasses, verbose)

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

	if verbose {
		log.Println(c.Params("id"))
		log.Println(ids)
		log.Println(items)
	}

	return json.NewEncoder(c.Response().BodyWriter()).Encode(items)
}

func main() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")

	// using standard library "flag" package
	flag.Bool("verbose", false, "verbose")
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

	awsRegions = viper.GetStringSlice("aws.regions")
	awsRoles = viper.GetStringSlice("aws.iam_role_arn")
	accountAliasses = viper.GetStringMapString("aws.account_aliasses")
	verbose = viper.GetBool("verbose")

	if verbose == true {
		log.Println(viper.AllSettings())
	}

	engine := html.NewFileSystem(http.FS(views), ".html")

	// Debug will print each template that is parsed, good for debugging
	engine.Debug(true)

	// engine.Layout("content")

	app := fiber.New(fiber.Config{
		Views:                 engine,
		DisableStartupMessage: false,
	})

	app.Use(compress.New())

	app.Use("/public", filesystem.New(filesystem.Config{
		Root: http.FS(assets),
	}))

	// Reload the templates on each render, good for development
	if verbose == true {
		engine.Reload(true) // Optional. Default: false
	}

	sess = session.Must(session.NewSession())

	appLogger := logger.New(logger.Config{
		Format: `${pid} ${locals:requestid} ${status} - ${method} ${path}​ ${query}​ ${queryParams}​` + "\n",
	})
	app.Use(appLogger)

	app.Get("/", mainRoute)
	app.Get("/search", searchRoute)
	app.Get("/api/search/:id", apiSearchRoute)

	app.Listen(viper.GetString("listen"))
}

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gofiber/fiber/v2"
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

func searchRoute(c *fiber.Ctx) error {
	ids := strings.Split(strings.Replace(c.Query("id"), "%2C", ",", -1), ",")
	
	ids = libs.Deduplicate(ids)

	items := libs.Describe(awsRegions, ids, awsRoles, sess, accountAliasses, verbose)

	if verbose {
		log.Println(c.Query("id"))
		log.Println(ids)
		log.Println(items)
	}

	return c.Render("index", fiber.Map{
            "Items": items,
            "IDs": strings.Join(ids, ","),
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
	flag.String("listen", ":3000", "listen address")
	flag.String("config", "./", "path to cloudpile.yml")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("CLOUDPILE")
	viper.AutomaticEnv()

	viper.SetConfigName("cloudpile")                 // name of config file (without extension)
	viper.SetConfigType("yaml")                   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(viper.GetString("config"))  // path to look for the config file in
	viperErr := viper.ReadInConfig() // Find and read the config file
	
	if viperErr != nil {             // Handle errors reading the config file
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

	engine := html.NewFileSystem(rice.MustFindBox("views").HTTPBox(), ".html")
	
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/", "./public")

	// Reload the templates on each render, good for development
	if verbose == true {
		engine.Reload(true) // Optional. Default: false
	}

	// Debug will print each template that is parsed, good for debugging
	// engine.Debug(true) // Optional. Default: false

	sess = session.Must(session.NewSession())

	// appLogger := logger.New(logger.Config{
	// 	Format: `{"pid":${pid}, "timestamp":"${time}", "status":${status}, "latency":"${latency}", "method":"${method}", "path":"${path}"}` + "\n",
	// })
	// app.Use(appLogger)

	app.Get("/", searchRoute)
	app.Get("/api/search/:id", apiSearchRoute)

	app.Listen(viper.GetString("listen"))
}

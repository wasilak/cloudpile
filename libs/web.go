package libs

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

//go:embed views
var views embed.FS

//go:embed assets
var assets embed.FS

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

func Web() {
	e := echo.New()

	e.Use(middleware.Gzip())

	e.HideBanner = true

	e.Debug = viper.GetBool("debug")

	t := &Template{
		templates: template.Must(template.ParseFS(getEmbededViews(), "*.html")),
	}

	e.Renderer = t

	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.Recover())

	// Enable metrics middleware
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	assetHandler := http.FileServer(getEmbededAssets())
	e.GET("/public/assets/*", echo.WrapHandler(http.StripPrefix("/public/assets/", assetHandler)))

	e.GET("/", MainRoute)
	e.GET("/list", ListRoute)
	e.GET("/api/list", ApiListRoute)
	e.GET("/search", SearchRoute)
	e.GET("/search/", SearchRoute)
	e.GET("/api/search/", ApiSearchRoute)
	e.GET("/api/search/:id", ApiSearchRoute)
	e.GET("/api/config/", ApiConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}

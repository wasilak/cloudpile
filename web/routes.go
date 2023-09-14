package web

import (
	"net/http"
	"strings"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/wasilak/cloudpile/cache"
	"github.com/wasilak/cloudpile/libs"
	"github.com/wasilak/cloudpile/resources"
)

func MainRoute(c echo.Context) error {
	var tempalateData interface{}
	return c.Render(http.StatusOK, "main", tempalateData)
}

func ApiConfigRoute(c echo.Context) error {
	return c.JSON(http.StatusOK, viper.GetStringMap("aws"))
}

func SearchRoute(c echo.Context) error {
	var ids []string

	ids = strings.Split(strings.Replace(c.QueryParam("id"), "%2C", ",", -1), ",")
	ids = libs.RemoveEmptyStrings(ids)
	ids = libs.Deduplicate(ids)

	slog.Debug("QueryDebug", "QueryParam('id')", c.QueryParam("id"), "ids", slog.AnyValue(ids))

	tempalateData := map[string]string{
		"IDs": strings.Join(ids, ","),
	}

	return c.Render(http.StatusOK, "search", tempalateData)
}

func ListRoute(c echo.Context) error {
	var tempalateData interface{}
	return c.Render(http.StatusOK, "list", tempalateData)
}

func ApiSearchRoute(c echo.Context) error {
	var ids []string

	ids = strings.Split(strings.Replace(c.Param("id"), "%2C", ",", -1), ",")
	ids = libs.RemoveEmptyStrings(ids)
	ids = libs.Deduplicate(ids)

	slog.Debug("QueryDebug", "QueryParam('id')", c.QueryParam("id"), "ids", slog.AnyValue(ids))

	var items []resources.Item
	if len(ids) > 0 {
		var err error
		items, err = libs.Run(ids, cache.CacheInstance, !viper.GetBool("cache.enabled"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, items)
}

func ApiListRoute(c echo.Context) error {
	var ids []string

	items, err := libs.Run(ids, cache.CacheInstance, viper.GetBool("cache.enabled"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, items)
}

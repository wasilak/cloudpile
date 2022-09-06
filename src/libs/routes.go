package libs

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

var AwsRegions []string
var AwsRoles []string
var AccountAliasses map[string]string
var CacheInstance Cache

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
	ids = RemoveEmptyStrings(ids)
	ids = Deduplicate(ids)

	log.Debug(c.QueryParam("id"))
	log.Debug(ids)

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
	ids = RemoveEmptyStrings(ids)
	ids = Deduplicate(ids)

	log.Debug(c.QueryParam("id"))
	log.Debug(ids)

	var items Items

	log.Debug(c.QueryParam("id"))
	log.Debug(ids)

	if len(ids) > 0 {
		items = Describe(AwsRegions, ids, AwsRoles, AccountAliasses, CacheInstance, false)
	}

	return c.JSON(http.StatusOK, items)
}

func ApiListRoute(c echo.Context) error {
	var ids []string
	var items interface{}

	items = Describe(AwsRegions, ids, AwsRoles, AccountAliasses, CacheInstance, false)

	return c.JSON(http.StatusOK, items)
}

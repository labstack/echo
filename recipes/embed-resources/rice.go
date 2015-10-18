package main

import (
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	// the file server for rice. "app" is the folder where the files come from.
	assetHandler := http.FileServer(rice.MustFindBox("app").HTTPBox())
	// serves the index.html from rice
	e.Get("/", func(c *echo.Context) error {
		assetHandler.ServeHTTP(c.Response().Writer(), c.Request())
		return nil
	})
	// servers other static files
	e.Get("/static/*", func(c *echo.Context) error {
		http.StripPrefix("/static/", assetHandler).
			ServeHTTP(c.Response().Writer(), c.Request())
		return nil
	})
	e.Run(":3000")
}

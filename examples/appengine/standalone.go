// +build !appengine,!appenginevm

package main

import (
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

func createMux() *echo.Echo {
	e := echo.New()

	// app engine provides this functionality so it's only included in standalone
	e.Use(mw.Recover())
	e.Use(mw.Logger())
	e.Use(mw.Gzip())

	return e
}

func init() {
	// static files are served by the app engine frontend/CDN based on the
	// configuration in the app.yaml file so we can offload this work but
	// need to serve them ourselves when running standalone

	// Serve index file
	e.Index("public/index.html")

	// Serve favicon
	e.Favicon("public/favicon.ico")

	// Serve static files
	e.Static("/scripts", "public/scripts")
}

// for standalone mode we run echo server ourselves
func main() {
	e.Run(":8080")
}

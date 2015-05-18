package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Debug mode
	e.Debug(true)

	//------------
	// Middleware
	//------------

	// Recover
	e.Use(mw.Recover())

	// Logger
	e.Use(mw.Logger())

	// Basic auth
	e.Use(mw.BasicAuth(func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}))

	//-------
	// Slash
	//-------

	e.Use(mw.StripTrailingSlash())

	// or

	//	e.Use(mw.RedirectToSlash())

	// Gzip
	e.Use(mw.Gzip())

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323")
}

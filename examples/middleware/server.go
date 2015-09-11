package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Debug mode
	e.Debug()

	//------------
	// Middleware
	//------------

	// Logger
	e.Use(mw.Logger())

	// Recover
	e.Use(mw.Recover())

	// Basic auth
	e.Use(mw.BasicAuth(func(usr, pwd string) bool {
		if usr == "joe" && pwd == "secret" {
			return true
		}
		return false
	}))

	// Gzip
	e.Use(mw.Gzip())

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323")
}

package main

import (
	"net/http"

	"github.com/trafficstars/echo"
	"github.com/trafficstars/echo/engine/standard"
	"github.com/trafficstars/echo/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	// Start server
	e.Run(standard.New(":1323"))
}

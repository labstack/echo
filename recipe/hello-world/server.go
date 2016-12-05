package main

import (
	"net/http"

	"gopkg.in/echo.v2"
	"gopkg.in/echo.v2/engine/standard"
	"gopkg.in/echo.v2/middleware"
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

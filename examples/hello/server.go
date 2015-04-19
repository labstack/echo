package main

import (
	"net/http"

	"github.com/labstack/echo"
)

// Handler
func hello(c *echo.Context) {
	c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(echo.Logger)

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":4444")
}

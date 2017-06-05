package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func init() {
	// Group with no parent middleware
	g := e.Group("/files", func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	g.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Your files!")
	})
}

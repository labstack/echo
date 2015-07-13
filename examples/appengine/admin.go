package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func init() {
	//-------
	// Group
	//-------

	// Group with parent middleware
	a := e.Group("/admin")
	a.Use(func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	a.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Welcome admin!")
	})
}

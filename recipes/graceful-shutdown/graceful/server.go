package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/tylerb/graceful"
)

func main() {
	// Setup
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Sue sews rose on slow joe crows nose")
	})

	graceful.ListenAndServe(e.Server(":1323"), 5*time.Second)
}

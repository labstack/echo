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
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Sue sews rose on slow joe crows nose")
	})
	e.Server.Addr = ":1323"

	// Serve it like a boss
	graceful.ListenAndServe(e.Server, 5*time.Second)
}

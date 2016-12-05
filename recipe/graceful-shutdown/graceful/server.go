package main

import (
	"net/http"
	"time"

	"gopkg.in/echo.v2"
	"gopkg.in/echo.v2/engine/standard"
	"github.com/tylerb/graceful"
)

func main() {
	// Setup
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Sue sews rose on slow joe crows nose")
	})
	std := standard.New(":1323")
	std.SetHandler(e)
	graceful.ListenAndServe(std.Server, 5*time.Second)
}

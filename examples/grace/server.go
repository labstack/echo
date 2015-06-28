package main

import (
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
)

func main() {
	// Setup
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		c.String(http.StatusOK, "Six sick bricks tick")
		return nil
	})

	// Use github.com/facebookgo/grace/gracehttp
	gracehttp.Serve(e.Server(":1323"))
}

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
		return c.String(http.StatusOK, "Six sick bricks tick")
	})

	gracehttp.Serve(e.Server(":1323"))
}

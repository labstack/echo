package main

import (
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/trafficstars/echo"
	"github.com/trafficstars/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Six sick bricks tick")
	})
	std := standard.New(":1323")
	std.SetHandler(e)
	gracehttp.Serve(std.Server)
}

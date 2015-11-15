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

	// Get the http.Server
	s := e.Server(":1323")

	// HTTP2 is currently enabled by default in echo.New(). To override TLS handshake errors
	// you will need to override the TLSConfig for the server so it does not attempt to validate
	// the connection using TLS as required by HTTP2
	s.TLSConfig = nil

	// Serve it like a boss
	gracehttp.Serve(s)
}

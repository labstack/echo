// +build appengine

package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func createMux() *echo.Echo {
	e := echo.New()
	// note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
	// app engine has it's own "main" wrapper - we just need to hook echo into the default handler
	http.Handle("/", e)
	return e
}

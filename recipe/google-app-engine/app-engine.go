// +build appengine

package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"net/http"
)

func createMux() *echo.Echo {
	e := echo.New()

	// note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
	// app engine has it's own "main" wrapper - we just need to hook echo into the default handler
	s := standard.New("")
	s.SetHandler(e)
	http.Handle("/", s)

	return e
}

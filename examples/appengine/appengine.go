// +build appengine appenginevm

package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func createMux() *echo.Echo {
	e := echo.New()

	// note: we don't need to provide the middleware or static handlers
	// for the app-engine version - that's taken care of by the platform

	return e
}

// app engine has it's own "main" - we just need to hook echo into the root
func init() {
	http.Handle("/", e)
}

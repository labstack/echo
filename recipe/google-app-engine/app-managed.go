// +build appenginevm

package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"google.golang.org/appengine"
	"net/http"
	"runtime"
)

func createMux() *echo.Echo {
	e := echo.New()

	// note: we don't need to provide the middleware or static handlers
	// for the appengine vm version - that's taken care of by the platform

	return e
}

func main() {
	// the appengine package provides a convenient method to handle the health-check requests
	// and also run the app on the correct port. We just need to add Echo to the default handler
	s := standard.New(":8080")
	s.SetHandler(e)
	http.Handle("/", s)
	appengine.Main()
}

// +build appenginevm

package main

import (
	"github.com/labstack/echo"
	"google.golang.org/appengine"
	"net/http"
	"runtime"
)

func createMux() *echo.Echo {
	// we're in a container on a Google Compute Engine instance so are not sandboxed anymore ...
	runtime.GOMAXPROCS(runtime.NumCPU())

	e := echo.New()

	// note: we don't need to provide the middleware or static handlers
	// for the appengine vm version - that's taken care of by the platform

	return e
}

func main() {
	// the appengine package provides a convenient method to handle the health-check requests
	// and also run the app on the correct port. We just need to add Echo to the default handler
	http.Handle("/", e)
	appengine.Main()
}

package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	// the file server for rice. "app" is the folder where the files come from.
	assetHandler := http.FileServer(rice.MustFindBox("app").HTTPBox())
	// serves the index.html from rice
	e.GET("/", standard.WrapHandler(assetHandler))

	// servers other static files
	e.GET("/static/*", standard.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	panic(e.Start(":3000"))
}

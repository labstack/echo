// +build !appengine,!appenginevm

package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func createMux() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	e.Index("public/index.html")
	e.Static("/public", "public")

	return e
}

func main() {
	e.Run(":8080")
}

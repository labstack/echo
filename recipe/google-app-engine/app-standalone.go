// +build !appengine,!appenginevm

package main

import (
	"gopkg.in/echo.v2"
	"gopkg.in/echo.v2/engine/standard"
	"gopkg.in/echo.v2/middleware"
)

func createMux() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	e.Use(middleware.Static("public"))

	return e
}

func main() {
	e.Run(standard.New(":8080"))
}

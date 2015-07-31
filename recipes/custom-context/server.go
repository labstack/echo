package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

type myCtx struct {
	echo.Context
}
func (c myCtx) HelloWorld() error {
	return c.String(http.StatusOK, "Hello, World!")
}

func helloWorld(c echo.Context) error {
	return c.(myCtx).HelloWorld()
}

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return h(myCtx{c})
		}
	})
	e.Get("/", helloWorld)

	e.Run(":1323")
}
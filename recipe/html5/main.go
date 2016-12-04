package main

import (
	"path"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const staticRoot = "./public"

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.File(path.Join(staticRoot, c.Param("*"))) == echo.ErrNotFound {
				return c.File(path.Join(staticRoot, "/"))
			}

			return h(c)
		}
	})

	e.Static("/", staticRoot)
	e.Logger.Fatal(e.Start(":1323"))
}

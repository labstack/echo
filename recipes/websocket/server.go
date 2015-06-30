package main

import (
	"io"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.WebSocket("/ws", func(c *echo.Context) error {
		io.Copy(c.Socket(), c.Socket())
		return nil
	})
	e.Run(":1323")
}

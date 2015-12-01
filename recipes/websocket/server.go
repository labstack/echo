package main

import (
	"fmt"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

func main() {
	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Static("/", "public")
	e.WebSocket("/ws", func(c *echo.Context) (err error) {
		ws := c.Socket()
		msg := ""

		for {
			if err = websocket.Message.Send(ws, "Hello, Client!"); err != nil {
				return
			}
			if err = websocket.Message.Receive(ws, &msg); err != nil {
				return
			}
			fmt.Println(msg)
		}
	})

	e.Run(":1323")
}

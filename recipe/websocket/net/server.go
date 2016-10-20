package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

func hello() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		for {
			// Write
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				log.Fatal(err)
			}

			// Read
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", msg)
		}
	})
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Static("../public"))
	e.GET("/ws", standard.WrapHandler(hello()))
	e.Run(standard.New(":1323"))
}

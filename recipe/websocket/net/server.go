package main

import (
	"fmt"
	"log"

	"gopkg.in/echo.v2"
	"gopkg.in/echo.v2/engine/standard"
	"gopkg.in/echo.v2/middleware"
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

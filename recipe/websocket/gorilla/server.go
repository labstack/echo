package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

var (
	upgrader = websocket.Upgrader{}
)

func hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		for {
			// Write
			err := c.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
			if err != nil {
				log.Fatal(err)
			}

			// Read
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", msg)
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Static("../public"))
	e.GET("/ws", standard.WrapHandler(http.HandlerFunc(hello())))
	e.Run(standard.New(":1323"))
}

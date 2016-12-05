package main

import (
	"math/rand"
	"net/http"
	"time"

	"gopkg.in/echo.v2"
	"gopkg.in/echo.v2/engine/standard"
	"gopkg.in/echo.v2/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Static("public"))

	// JSONP
	e.GET("/jsonp", func(c echo.Context) error {
		callback := c.QueryParam("callback")
		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Sent via JSONP"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)
		return c.JSONP(http.StatusOK, callback, &content)
	})

	// Start server
	e.Run(standard.New(":1323"))
}

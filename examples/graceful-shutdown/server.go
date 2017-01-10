package main

import (
	"time"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	// Setting up the termination timeout to 30 seconds.
	e.ShutdownTimeout = 30 * time.Second

	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(200, "OK")
	})

	e.Logger.Fatal(e.Start(":1323"))
}

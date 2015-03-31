package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
	"github.com/mattn/go-colorable"
)

func Logger(h echo.HandlerFunc) echo.HandlerFunc {
	log.SetOutput(colorable.NewColorableStdout())
	return echo.HandlerFunc(func(c *echo.Context) {
		start := time.Now()
		h(c)
		end := time.Now()
		col := color.Green
		m := c.Request.Method
		p := c.Request.URL.Path
		s := c.Response.Status()

		switch {
		case s >= 500:
			col = color.Red
		case s >= 400:
			col = color.Yellow
		case s >= 300:
			col = color.Cyan
		}

		log.Printf("%s %s %s %s", m, p, col(s), end.Sub(start))
	})
}

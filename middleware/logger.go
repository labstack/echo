package middleware

import (
	"log"
	"time"

	"github.com/labstack/bolt"
	"github.com/labstack/gommon/color"
)

func Logger() bolt.HandlerFunc {
	return func(c *bolt.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		co := color.Green
		m := c.Request.Method
		p := c.Request.URL.Path
		s := c.Response.Status()

		switch {
		case s >= 500:
			co = color.Red
		case s >= 400:
			co = color.Yellow
		case s >= 300:
			co = color.Cyan
		}

		log.Printf("%s %s %s %s", m, p, co(s), end.Sub(start))
	}
}

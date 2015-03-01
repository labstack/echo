package middleware

import (
	"log"
	"time"

	"github.com/labstack/bolt"
	"labstack.com/common/utils"
)

func Logger() bolt.HandlerFunc {
	return func(c *bolt.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		color := utils.Green
		m := c.Request.Method
		p := c.Request.URL.Path
		s := c.Response.Status()

		switch {
		case s >= 500:
			color = utils.Red
		case s >= 400:
			color = utils.Yellow
		case s >= 300:
			color = utils.Cyan
		}

		log.Printf("%s %s %s %s", m, p, color(s), end.Sub(start))
	}
}

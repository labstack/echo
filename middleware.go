package echo

import (
	"log"
	"time"

	"labstack.com/gommon/color"
)

func Logger(h HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		start := time.Now()
		if err := h(c); err != nil {
			c.Error(err)
		}
		end := time.Now()
		m := c.Request.Method
		p := c.Request.URL.Path
		n := c.Response.Status()
		col := color.Green

		switch {
		case n >= 500:
			col = color.Red
		case n >= 400:
			col = color.Yellow
		case n >= 300:
			col = color.Cyan
		}

		log.Printf("%s %s %s %s", m, p, col(n), end.Sub(start))
		return nil
	}
}

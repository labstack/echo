package middleware

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

func parseIP(s string) string {
	if s != "" {
		if len(strings.Split(s, ".")) == 4 {
			// ipv4
			s = strings.Join(strings.Split(s, ":")[:1], ":")
		} else {
			// ipv6
			if s[0] == '[' {
				// ipv6 with port
				s = strings.Join(strings.Split(s[1:], "]")[:1], "")
			}
		}
		if ip := net.ParseIP(s); ip != nil {
			return ip.String()
		}
	}
	return "unknown"
}

func Logger() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {

			remoteAddr := c.Request().RemoteAddr
			if realIP := c.Request().Header.Get("X-Real-IP"); realIP != "" {
				remoteAddr = realIP
			}
			if realIP := c.Request().Header.Get("X-Forwarded-For"); realIP != "" {
				remoteAddr = realIP
			}

			start := time.Now()
			if err := h(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			method := c.Request().Method
			path := c.Request().URL.Path
			if path == "" {
				path = "/"
			}
			size := c.Response().Size()

			n := c.Response().Status()
			code := color.Green(n)
			switch {
			case n >= 500:
				code = color.Red(n)
			case n >= 400:
				code = color.Yellow(n)
			case n >= 300:
				code = color.Cyan(n)
			}

			log.Printf("%s %s %s %s %s %d", parseIP(remoteAddr), method, path, code, stop.Sub(start), size)
			return nil
		}
	}
}

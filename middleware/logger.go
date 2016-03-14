package middleware

import (
	"net"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

type (
	LoggerConfig struct {
	}
)

var (
	DefaultLoggerConfig = LoggerConfig{}
)

func Logger() echo.MiddlewareFunc {
	return LoggerFromConfig(DefaultLoggerConfig)
}

func LoggerFromConfig(config LoggerConfig) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			logger := c.Logger()

			remoteAddr := req.RemoteAddress()
			if ip := req.Header().Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header().Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			start := time.Now()
			if err := next.Handle(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			method := req.Method()
			path := req.URL().Path()
			if path == "" {
				path = "/"
			}
			size := res.Size()

			n := res.Status()
			code := color.Green(n)
			switch {
			case n >= 500:
				code = color.Red(n)
			case n >= 400:
				code = color.Yellow(n)
			case n >= 300:
				code = color.Cyan(n)
			}

			logger.Printf("%s %s %s %s %s %d", remoteAddr, method, path, code, stop.Sub(start), size)
			return nil
		})
	}
}

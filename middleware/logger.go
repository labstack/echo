package middleware

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
	"github.com/valyala/fasttemplate"
)

type (
	LoggerConfig struct {
		Format   string
		template *fasttemplate.Template
	}
)

var (
	DefaultLoggerConfig = LoggerConfig{
		Format: "time=${time_rfc3339}, remote_ip=${remote_ip}, method=${method}, path=${path}, status=${status}, response_time=${response_time}, size=${size}\n",
	}
)

func Logger() echo.MiddlewareFunc {
	return LoggerFromConfig(DefaultLoggerConfig)
}

func LoggerFromConfig(config LoggerConfig) echo.MiddlewareFunc {
	config.template = fasttemplate.New(config.Format, "${", "}")

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			remoteAddr := req.RemoteAddress()
			output := c.Logger().Output()

			if ip := req.Header().Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header().Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			start := time.Now()
			if err := next.Handle(c); err != nil {
				return err
			}
			stop := time.Now()
			method := []byte(req.Method())
			path := req.URL().Path()
			if path == "" {
				path = "/"
			}
			took := stop.Sub(start)
			size := strconv.FormatInt(res.Size(), 10)

			n := res.Status()
			status := color.Green(n)
			switch {
			case n >= 500:
				status = color.Red(n)
			case n >= 400:
				status = color.Yellow(n)
			case n >= 300:
				status = color.Cyan(n)
			}

			_, err = config.template.ExecuteFunc(output, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_rfc3339":
					return w.Write([]byte(time.Now().Format(time.RFC3339)))
				case "remote_ip":
					return w.Write([]byte(remoteAddr))
				case "method":
					return w.Write(method)
				case "path":
					return w.Write([]byte(path))
				case "status":
					return w.Write([]byte(status))
				case "response_time":
					return w.Write([]byte(took.String()))
				case "size":
					return w.Write([]byte(size))
				default:
					return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
				}
			})
			return
		})
	}
}

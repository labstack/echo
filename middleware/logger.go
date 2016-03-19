package middleware

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
	"github.com/mattn/go-isatty"
	"github.com/valyala/fasttemplate"
)

type (
	LoggerConfig struct {
		Format   string
		Output   io.Writer
		template *fasttemplate.Template
		color    *color.Color
	}
)

var (
	DefaultLoggerConfig = LoggerConfig{
		Format: "time=${time_rfc3339}, remote_ip=${remote_ip}, method=${method}, path=${path}, status=${status}, response_time=${response_time}, size=${size}\n",
		color:  color.New(),
		Output: os.Stdout,
	}
)

func Logger() echo.MiddlewareFunc {
	return LoggerFromConfig(DefaultLoggerConfig)
}

func LoggerFromConfig(config LoggerConfig) echo.MiddlewareFunc {
	config.template = fasttemplate.New(config.Format, "${", "}")
	config.color = color.New()
	if w, ok := config.Output.(*os.File); ok && !isatty.IsTerminal(w.Fd()) {
		config.color.Disable()
	}

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
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

			_, err = config.template.ExecuteFunc(config.Output, func(w io.Writer, tag string) (int, error) {
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

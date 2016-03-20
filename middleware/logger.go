package middleware

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/color"
	"github.com/mattn/go-isatty"
	"github.com/valyala/fasttemplate"
)

type (
	// LoggerConfig defines config for logger middleware.
	//
	LoggerConfig struct {
		// Format is the log format.
		//
		// Example "${remote_id} ${status}"
		// Available tags:
		// - time_rfc3339
		// - remote_ip
		// - method
		// - path
		// - status
		// - response_time
		// - response_size
		Format string

		// Output is the writer where logs are written. Default is `os.Stdout`.
		Output io.Writer

		template *fasttemplate.Template
		color    *color.Color
	}
)

var (
	// DefaultLoggerConfig is the default logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Format: "time=${time_rfc3339}, remote_ip=${remote_ip}, method=${method}, path=${path}, status=${status}, response_time=${response_time}, response_size=${response_size} bytes\n",
		color:  color.New(),
		Output: os.Stdout,
	}
)

func remoteAddress(request engine.Request) string {
	header := request.Header()

	if ip := header.Get(echo.XRealIP); ip != "" {
		return ip
	}

	if ip := header.Get(echo.XForwardedFor); ip != "" {
		return ip
	}

	ip, _, _ := net.SplitHostPort(request.RemoteAddress())

	return ip
}

func colorizeStatusCode(code int) string {
	switch {
	case code >= 500:
		return color.Red(code)
	case code >= 400:
		return color.Yellow(code)
	case code >= 300:
		return color.Cyan(code)
	default:
		return color.Green(code)
	}
}

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerFromConfig(DefaultLoggerConfig)
}

// LoggerFromConfig returns a logger middleware from config.
// See `Logger()`.
func LoggerFromConfig(config LoggerConfig) echo.MiddlewareFunc {
	config.template = fasttemplate.New(config.Format, "${", "}")
	config.color = color.New()

	if w, ok := config.Output.(*os.File); ok && !isatty.IsTerminal(w.Fd()) {
		config.color.Disable()
	}

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			start := time.Now()

			if err := next.Handle(c); err != nil {
				return err
			}

			stop := time.Now()

			_, err := config.template.ExecuteFunc(config.Output, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_rfc3339":
					return w.Write([]byte(time.Now().Format(time.RFC3339)))
				case "remote_ip":
					return w.Write([]byte(remoteAddress(req)))
				case "method":
					return w.Write([]byte(req.Method()))
				case "path":
					path := req.URL().Path()

					if path == "" {
						path = "/"
					}

					return w.Write([]byte(path))
				case "status":
					return w.Write([]byte(colorizeStatusCode(res.Status())))
				case "response_time":
					return w.Write([]byte(stop.Sub(start).String()))
				case "response_size":
					return w.Write([]byte(strconv.FormatInt(res.Size(), 10)))
				default:
					return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
				}
			})

			return err
		})
	}
}

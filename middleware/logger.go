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
	isatty "github.com/mattn/go-isatty"
	"github.com/valyala/fasttemplate"
)

type (
	// LoggerConfig defines the config for logger middleware.
	LoggerConfig struct {
		// Format is the log format which can be constructed using the following tags:
		//
		// - time_rfc3339
		// - remote_ip
		// - uri
		// - method
		// - path
		// - status
		// - response_time
		// - response_size
		//
		// Example "${remote_id} ${status}"
		//
		// Optional with default value as `DefaultLoggerConfig.Format`.
		Format string

		// Output is the writer where logs are written.
		// Optional with default value as `DefaultLoggerConfig.Output`.
		Output io.Writer

		template *fasttemplate.Template
		color    *color.Color
	}
)

var (
	// DefaultLoggerConfig is the default logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Format: "time=${time_rfc3339}, remote_ip=${remote_ip}, method=${method}, " +
			"uri=${uri}, status=${status}, took=${response_time}, sent=${response_size} bytes\n",
		color:  color.New(),
		Output: os.Stdout,
	}
)

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerFromConfig(DefaultLoggerConfig)
}

// LoggerFromConfig returns a logger middleware from config.
// See `Logger()`.
func LoggerFromConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.color = color.New()
	if w, ok := config.Output.(*os.File); ok && !isatty.IsTerminal(w.Fd()) {
		config.color.Disable()
	}

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) (err error) {
			rq := c.Request()
			rs := c.Response()
			start := time.Now()
			if err = next.Handle(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			_, err = config.template.ExecuteFunc(config.Output, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_rfc3339":
					return w.Write([]byte(time.Now().Format(time.RFC3339)))
				case "remote_ip":
					ra := rq.RemoteAddress()
					if ip := rq.Header().Get(echo.XRealIP); ip != "" {
						ra = ip
					} else if ip = rq.Header().Get(echo.XForwardedFor); ip != "" {
						ra = ip
					} else {
						ra, _, _ = net.SplitHostPort(ra)
					}
					return w.Write([]byte(ra))
				case "uri":
					return w.Write([]byte(rq.URI()))
				case "method":
					return w.Write([]byte(rq.Method()))
				case "path":
					p := rq.URL().Path()
					if p == "" {
						p = "/"
					}
					return w.Write([]byte(p))
				case "status":
					n := rs.Status()
					s := color.Green(n)
					switch {
					case n >= 500:
						s = color.Red(n)
					case n >= 400:
						s = color.Yellow(n)
					case n >= 300:
						s = color.Cyan(n)
					}
					return w.Write([]byte(s))
				case "response_time":
					return w.Write([]byte(stop.Sub(start).String()))
				case "response_size":
					return w.Write([]byte(strconv.FormatInt(rs.Size(), 10)))
				default:
					return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
				}
			})
			return
		})
	}
}

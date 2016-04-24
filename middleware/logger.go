package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
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
		// Optional, with default value as `DefaultLoggerConfig.Format`.
		Format string

		// Output is the writer where logs are written.
		// Optional with default value as os.Stdout.
		Output io.Writer

		template   *fasttemplate.Template
		color      *color.Color
		bufferPool sync.Pool
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
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a logger middleware from config.
// See `Logger()`.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.color = color.New()
	if w, ok := config.Output.(*os.File); !ok || !isatty.IsTerminal(w.Fd()) {
		config.color.Disable()
	}
	config.bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			buf := config.bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer config.bufferPool.Put(buf)

			_, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_rfc3339":
					return w.Write([]byte(time.Now().Format(time.RFC3339)))
				case "remote_ip":
					ra := req.RemoteAddress()
					if ip := req.Header().Get(echo.HeaderXRealIP); ip != "" {
						ra = ip
					} else if ip = req.Header().Get(echo.HeaderXForwardedFor); ip != "" {
						ra = ip
					} else {
						ra, _, _ = net.SplitHostPort(ra)
					}
					return w.Write([]byte(ra))
				case "uri":
					return w.Write([]byte(req.URI()))
				case "method":
					return w.Write([]byte(req.Method()))
				case "path":
					p := req.URL().Path()
					if p == "" {
						p = "/"
					}
					return w.Write([]byte(p))
				case "status":
					n := res.Status()
					s := config.color.Green(n)
					switch {
					case n >= 500:
						s = config.color.Red(n)
					case n >= 400:
						s = config.color.Yellow(n)
					case n >= 300:
						s = config.color.Cyan(n)
					}
					return w.Write([]byte(s))
				case "response_time":
					return w.Write([]byte(stop.Sub(start).String()))
				case "response_size":
					return w.Write([]byte(strconv.FormatInt(res.Size(), 10)))
				default:
					return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
				}
			})
			if err == nil {
				config.Output.Write(buf.Bytes())
			}
			return
		}
	}
}

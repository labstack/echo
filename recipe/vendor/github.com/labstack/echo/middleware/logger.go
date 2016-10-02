package middleware

import (
	"bytes"
	"io"
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
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Log format which can be constructed using the following tags:
		//
		// - time_rfc3339
		// - id (Request ID - Not implemented)
		// - remote_ip
		// - uri
		// - host
		// - method
		// - path
		// - referer
		// - user_agent
		// - status
		// - latency (In microseconds)
		// - latency_human (Human readable)
		// - bytes_in (Bytes received)
		// - bytes_out (Bytes sent)
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		Format string `json:"format"`

		// Output is a writer where logs are written.
		// Optional. Default value os.Stdout.
		Output io.Writer

		template   *fasttemplate.Template
		color      *color.Color
		bufferPool sync.Pool
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper: defaultSkipper,
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
		Output: os.Stdout,
		color:  color.New(),
	}
)

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
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
			if config.Skipper(c) {
				return next(c)
			}

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
					ra := req.RealIP()
					return w.Write([]byte(ra))
				case "host":
					return w.Write([]byte(req.Host()))
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
				case "referer":
					return w.Write([]byte(req.Referer()))
				case "user_agent":
					return w.Write([]byte(req.UserAgent()))
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
				case "latency":
					l := stop.Sub(start).Nanoseconds() / 1000
					return w.Write([]byte(strconv.FormatInt(l, 10)))
				case "latency_human":
					return w.Write([]byte(stop.Sub(start).String()))
				case "bytes_in":
					b := req.Header().Get(echo.HeaderContentLength)
					if b == "" {
						b = "0"
					}
					return w.Write([]byte(b))
				case "bytes_out":
					return w.Write([]byte(strconv.FormatInt(res.Size(), 10)))
				}
				return 0, nil
			})
			if err == nil {
				config.Output.Write(buf.Bytes())
			}
			return
		}
	}
}

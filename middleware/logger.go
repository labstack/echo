package middleware

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/labstack/echo"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Availabe logger fields:
		//
		// - time (Unix time)
		// - id (Request ID - Not implemented)
		// - remote_ip
		// - uri
		// - host
		// - method
		// - path
		// - referer
		// - user_agent
		// - status
		// - latency (In nanosecond)
		// - latency_human (Human readable)
		// - bytes_in (Bytes received)
		// - bytes_out (Bytes sent)
		// - header:<name>
		// - query:<name>
		// - form:<name>

		// Optional. Default value DefaultLoggerConfig.Fields.
		Fields []string `json:"fields"`

		// Output is a writer where logs are written.
		// Optional. Default value os.Stdout.
		Output io.Writer
	}

	loggerFields struct {
		// ID string `json:"id,omitempty"` (Request ID - Not implemented)
		Time         int64             `json:"time,omitempty"`
		RemoteIP     string            `json:"remote_ip,omitempty"`
		URI          string            `json:"uri,omitempty"`
		Host         string            `json:"host,omitempty"`
		Method       string            `json:"method,omitempty"`
		Path         string            `json:"path,omitempty"`
		Referer      string            `json:"referer,omitempty"`
		UserAgent    string            `json:"user_agent,omitempty"`
		Status       int               `json:"status,omitempty"`
		Latency      time.Duration     `json:"latency,omitempty"`
		LatencyHuman string            `json:"latency_human,omitempty"`
		BytesIn      int64             `json:"bytes_in,omitempty"`
		BytesOut     int64             `json:"bytes_out,omitempty"`
		Header       map[string]string `json:"header,omitempty"`
		Form         map[string]string `json:"form,omitempty"`
		Query        map[string]string `json:"query,omitempty"`
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper: defaultSkipper,
		Fields: []string{
			"time",
			"remote_ip",
			"host",
			"method",
			"uri",
			"status",
			"latency",
			"latency_human",
			"bytes_in",
			"bytes_out",
		},
		Output: os.Stdout,
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
	if len(config.Fields) == 0 {
		config.Fields = DefaultLoggerConfig.Fields
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
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
			fields := &loggerFields{
				Header: make(map[string]string),
				Form:   make(map[string]string),
				Query:  make(map[string]string),
			}

			for _, f := range config.Fields {
				switch f {
				case "time":
					fields.Time = time.Now().Unix()
				case "remote_ip":
					fields.RemoteIP = c.RealIP()
				case "host":
					fields.Host = req.Host
				case "uri":
					fields.URI = req.RequestURI
				case "method":
					fields.Method = req.Method
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					fields.Path = p
				case "referer":
					fields.Referer = req.Referer()
				case "user_agent":
					fields.UserAgent = req.UserAgent()
				case "status":
					// n := res.Status
					// s := config.color.Green(n)
					// switch {
					// case n >= 500:
					// 	s = config.color.Red(n)
					// case n >= 400:
					// 	s = config.color.Yellow(n)
					// case n >= 300:
					// 	s = config.color.Cyan(n)
					// }
					// return w.Write([]byte(s))
					fields.Status = res.Status
				case "latency":
					fields.Latency = stop.Sub(start)
				case "latency_human":
					fields.LatencyHuman = stop.Sub(start).String()
				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					l, _ := strconv.ParseInt(cl, 10, 64)
					fields.BytesIn = l
				case "bytes_out":
					fields.BytesOut = res.Size
				default:
					switch {
					case strings.HasPrefix(f, "header:"):
						k := f[7:]
						fields.Header[k] = c.Request().Header.Get(k)
					case strings.HasPrefix(f, "form:"):
						k := f[5:]
						fields.Form[k] = c.Request().Header.Get(k)
					case strings.HasPrefix(f, "query:"):
						k := f[6:]
						fields.Query[k] = c.Request().Header.Get(k)
					}
				}
			}

			// Write
			enc := json.NewEncoder(config.Output)
			return enc.Encode(fields)
		}
	}
}

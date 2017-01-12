package middleware

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/labstack/echo"
	"github.com/labstack/echo/db"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Availabe logger fields:
		//
		// - time
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

		// Output is where logs are written.
		// Optional. Default value &Stream{os.Stdout}.
		Output db.Logger
	}

	// Stream implements `db.Logger`.
	Stream struct {
		io.Writer
	}
)

// LogRequest encodes `db.Request` into a stream.
func (s *Stream) Log(r *db.Request) error {
	enc := json.NewEncoder(s)
	return enc.Encode(r)
}

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
		Output: &Stream{os.Stdout},
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
			request := &db.Request{
				Header: make(map[string]string),
				Query:  make(map[string]string),
				Form:   make(map[string]string),
			}

			for _, f := range config.Fields {
				switch f {
				case "time":
					t := time.Now()
					request.Time = &t
				case "remote_ip":
					request.RemoteIP = c.RealIP()
				case "host":
					request.Host = req.Host
				case "uri":
					request.URI = req.RequestURI
				case "method":
					request.Method = req.Method
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					request.Path = p
				case "referer":
					request.Referer = req.Referer()
				case "user_agent":
					request.UserAgent = req.UserAgent()
				case "status":
					request.Status = res.Status
				case "latency":
					request.Latency = stop.Sub(start)
				case "latency_human":
					request.LatencyHuman = stop.Sub(start).String()
				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					l, _ := strconv.ParseInt(cl, 10, 64)
					request.BytesIn = l
				case "bytes_out":
					request.BytesOut = res.Size
				default:
					switch {
					case strings.HasPrefix(f, "header:"):
						k := f[7:]
						request.Header[k] = c.Request().Header.Get(k)
					case strings.HasPrefix(f, "query:"):
						k := f[6:]
						request.Query[k] = c.QueryParam(k)
					case strings.HasPrefix(f, "form:"):
						k := f[5:]
						request.Form[k] = c.FormValue(k)
					}
				}
			}

			// Write
			return config.Output.Log(request)
		}
	}
}

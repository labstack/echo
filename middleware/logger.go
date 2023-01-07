package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/valyala/fasttemplate"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Tags to construct the logger format.
	//
	// - time_unix
	// - time_unix_milli
	// - time_unix_micro
	// - time_unix_nano
	// - time_rfc3339
	// - time_rfc3339_nano
	// - time_custom
	// - id (Request ID)
	// - remote_ip
	// - uri
	// - host
	// - method
	// - path
	// - route
	// - protocol
	// - referer
	// - user_agent
	// - status
	// - error
	// - latency (In nanoseconds)
	// - latency_human (Human readable)
	// - bytes_in (Bytes received)
	// - bytes_out (Bytes sent)
	// - header:<NAME>
	// - query:<NAME>
	// - form:<NAME>
	// - custom (see CustomTagFunc field)
	//
	// Example "${remote_ip} ${status}"
	//
	// Optional. Default value DefaultLoggerConfig.Format.
	Format string

	// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
	CustomTimeFormat string

	// CustomTagFunc is function called for `${custom}` tag to output user implemented text by writing it to buf.
	// Make sure that outputted text creates valid JSON string with other logged tags.
	// Optional.
	CustomTagFunc func(c echo.Context, buf *bytes.Buffer) (int, error)

	// Output is a writer where logs in JSON format are written.
	// Optional. Default destination `echo.Logger.Infof()`
	Output io.Writer

	template *fasttemplate.Template
	pool     *sync.Pool
}

// DefaultLoggerConfig is the default Logger middleware config.
var DefaultLoggerConfig = LoggerConfig{
	Skipper: DefaultSkipper,
	Format: `{"time":"${time_rfc3339_nano}","level":"INFO","id":"${id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	CustomTimeFormat: "2006-01-02 15:04:05.00000",
}

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config or panics on invalid configuration.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts LoggerConfig to middleware or returns an error for invalid configuration
func (config LoggerConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			start := time.Now()
			err := next(c)
			if err != nil {
				// When global error handler writes the error to the client the Response gets "committed". This state can be
				// checked with `c.Response().Committed` field.
				c.Error(err)
			}
			stop := time.Now()

			buf := config.pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer config.pool.Put(buf)

			_, tmplErr := config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "custom":
					if config.CustomTagFunc == nil {
						return 0, nil
					}
					return config.CustomTagFunc(c, buf)
				case "time_unix":
					return buf.WriteString(strconv.FormatInt(stop.Unix(), 10))
				case "time_unix_milli":
					return buf.WriteString(strconv.FormatInt(stop.UnixMilli(), 10))
				case "time_unix_micro":
					return buf.WriteString(strconv.FormatInt(stop.UnixMicro(), 10))
				case "time_unix_nano":
					return buf.WriteString(strconv.FormatInt(stop.UnixNano(), 10))
				case "time_rfc3339":
					return buf.WriteString(stop.Format(time.RFC3339))
				case "time_rfc3339_nano":
					return buf.WriteString(stop.Format(time.RFC3339Nano))
				case "time_custom":
					return buf.WriteString(stop.Format(config.CustomTimeFormat))
				case "id":
					id := req.Header.Get(echo.HeaderXRequestID)
					if id == "" {
						id = res.Header().Get(echo.HeaderXRequestID)
					}
					return buf.WriteString(id)
				case "remote_ip":
					return buf.WriteString(c.RealIP())
				case "host":
					return buf.WriteString(req.Host)
				case "uri":
					return buf.WriteString(req.RequestURI)
				case "method":
					return buf.WriteString(req.Method)
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "route":
					return buf.WriteString(c.Path())
				case "protocol":
					return buf.WriteString(req.Proto)
				case "referer":
					return buf.WriteString(req.Referer())
				case "user_agent":
					return buf.WriteString(req.UserAgent())
				case "status":
					status := res.Status
					if err != nil {
						var httpErr *echo.HTTPError
						if errors.As(err, &httpErr) {
							status = httpErr.Code
						}
					}
					return buf.WriteString(strconv.Itoa(status))
				case "error":
					if err != nil {
						// Error may contain invalid JSON e.g. `"`
						b, _ := json.Marshal(err.Error())
						b = b[1 : len(b)-1]
						return buf.Write(b)
					}
				case "latency":
					l := stop.Sub(start)
					return buf.WriteString(strconv.FormatInt(int64(l), 10))
				case "latency_human":
					return buf.WriteString(stop.Sub(start).String())
				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					return buf.WriteString(cl)
				case "bytes_out":
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				default:
					switch {
					case strings.HasPrefix(tag, "header:"):
						return buf.Write([]byte(c.Request().Header.Get(tag[7:])))
					case strings.HasPrefix(tag, "query:"):
						return buf.Write([]byte(c.QueryParam(tag[6:])))
					case strings.HasPrefix(tag, "form:"):
						return buf.Write([]byte(c.FormValue(tag[5:])))
					case strings.HasPrefix(tag, "cookie:"):
						cookie, cookieErr := c.Cookie(tag[7:])
						if cookieErr == nil {
							return buf.Write([]byte(cookie.Value))
						}
					}
				}
				return 0, nil
			})
			if tmplErr != nil {
				if err != nil {
					return fmt.Errorf("error in middleware chain and also failed to create log from template: %v: %w", tmplErr, err)
				}
				return fmt.Errorf("failed to create log from template: %w", tmplErr)
			}

			if config.Output != nil {
				if _, lErr := config.Output.Write(buf.Bytes()); lErr != nil {
					return lErr
				}
			} else {
				if _, lErr := c.Echo().Logger.Write(buf.Bytes()); lErr != nil {
					return lErr
				}
			}
			return err
		}
	}, nil
}

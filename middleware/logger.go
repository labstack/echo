package middleware

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/valyala/fasttemplate"
	"go.uber.org/zap"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Tags to construct the logger format.
		//
		// - time_unix
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
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		Format string `yaml:"format"`

		// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
		CustomTimeFormat string `yaml:"custom_time_format"`

		LoggerHandler func(c echo.Context, m map[string]string)

		template *fasttemplate.Template
	}
)

// NOTE: optionaly here can be handler to diffrent log libraries, or it can be left
// for user to define.
func NewZapLoggerHandler(l *zap.Logger) func(c echo.Context, m map[string]string) {
	return func(c echo.Context, m map[string]string) {
		l.Info("request log", zap.String("id", m["id"]) /*TODO: ....*/)
	}
}

func NewStdLoggerHandler(l *log.Logger) func(c echo.Context, m map[string]string) {
	return func(c echo.Context, m map[string]string) {
		l.Printf("request log, id=%s", m["id"])
	}
}

// func NewLogrusLoggerHandler(l *log.Logger) func(c echo.Context, m map[string]string) {}
// func NewLog15LoggerHandler(l *log.Logger) func(c echo.Context, m map[string]string) {}
// etc.

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper: DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}
)

// Logger returns a middleware that logs HTTP requests.
func Logger(handler func(c echo.Context, m map[string]string)) echo.MiddlewareFunc {
	c := DefaultLoggerConfig
	c.LoggerHandler = handler
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

	config.template = fasttemplate.New(config.Format, "${", "}")

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
			m := make(map[string]string)

			var buf bytes.Buffer
			if _, err = config.template.ExecuteFunc(&buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_unix":
					m["time_unix"] = strconv.FormatInt(time.Now().Unix(), 10)
				case "id":
					id := req.Header.Get(echo.HeaderXRequestID)
					if id == "" {
						id = res.Header().Get(echo.HeaderXRequestID)
					}
					m["id"] = id
				case "latency":
					l := stop.Sub(start)
					m["latency"] = strconv.FormatInt(int64(l), 10)
				}
				// TODO:
				return 0, nil
			}); err != nil {
				return
			}

			config.LoggerHandler(c, m)
			return
		}
	}
}

package middleware

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
)

type (
	// GzipConfig defines the config for Gzip middleware.
	GzipConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Gzip compression level.
		// Optional. Default value -1.
		Level int `json:"level"`
	}

	gzipResponseWriter struct {
		engine.Response
		io.Writer
	}
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Skipper: defaultSkipper,
		Level:   -1,
	}
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithConfig return Gzip middleware from config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultGzipConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}

	pool := gzipPool(config)
	scheme := "gzip"

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header().Get(echo.HeaderAcceptEncoding), scheme) {
				rw := res.Writer()
				gw := pool.Get().(*gzip.Writer)
				gw.Reset(rw)
				defer func() {
					if res.Size() == 0 {
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.SetWriter(rw)
						res.Header().Del(echo.HeaderContentEncoding)
						gw.Reset(ioutil.Discard)
					}
					gw.Close()
					pool.Put(gw)
				}()
				g := gzipResponseWriter{Response: res, Writer: gw}
				res.Header().Set(echo.HeaderContentEncoding, scheme)
				res.SetWriter(g)
			}
			return next(c)
		}
	}
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get(echo.HeaderContentType) == "" {
		g.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	return g.Writer.Write(b)
}

func gzipPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(ioutil.Discard, config.Level)
			return w
		},
	}
}

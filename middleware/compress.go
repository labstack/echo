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
	// GzipConfig defines the config for gzip middleware.
	GzipConfig struct {
		// Level is the gzip level.
		// Optional with default value as `DefaultGzipConfig.Level`.
		Level int
	}

	gzipResponseWriter struct {
		engine.Response
		io.Writer
	}
)

var (
	// DefaultGzipConfig is the default gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Level: -1,
	}
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipFromConfig(DefaultGzipConfig)
}

// GzipFromConfig return gzip middleware from config.
// See `Gzip()`.
func GzipFromConfig(config GzipConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}

	pool := gzipPool(config)
	scheme := "gzip"

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			rs := c.Response()
			rs.Header().Add(echo.Vary, echo.AcceptEncoding)
			if strings.Contains(c.Request().Header().Get(echo.AcceptEncoding), scheme) {
				rw := rs.Writer()
				gw := pool.Get().(*gzip.Writer)
				gw.Reset(rw)
				defer func() {
					if rs.Size() == 0 {
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						rs.SetWriter(rw)
						rs.Header().Del(echo.ContentEncoding)
						gw.Reset(ioutil.Discard)
					}
					gw.Close()
					pool.Put(gw)
				}()
				g := gzipResponseWriter{Response: rs, Writer: gw}
				rs.Header().Set(echo.ContentEncoding, scheme)
				rs.SetWriter(g)
			}
			return next.Handle(c)
		})
	}
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get(echo.ContentType) == "" {
		g.Header().Set(echo.ContentType, http.DetectContentType(b))
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

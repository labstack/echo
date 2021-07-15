package middleware

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v5"
)

const (
	gzipScheme = "gzip"
)

// GzipConfig defines the config for Gzip middleware.
type GzipConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Gzip compression level.
	// Optional. Default value -1.
	Level int
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteBody bool
}

// Gzip returns a middleware which compresses HTTP response using gzip compression scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipWithConfig(GzipConfig{})
}

// GzipWithConfig returns a middleware which compresses HTTP response using gzip compression scheme.
func GzipWithConfig(config GzipConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts GzipConfig to middleware or returns an error for invalid configuration
func (config GzipConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Level < -2 || config.Level > 9 { // these are consts: gzip.HuffmanOnly and gzip.BestCompression
		return nil, errors.New("invalid gzip level")
	}
	if config.Level == 0 {
		config.Level = -1
	}

	pool := gzipCompressPool(config)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), gzipScheme) {
				res.Header().Set(echo.HeaderContentEncoding, gzipScheme) // Issue #806
				i := pool.Get()
				w, ok := i.(*gzip.Writer)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}
				rw := res.Writer
				w.Reset(rw)
				grw := &gzipResponseWriter{Writer: w, ResponseWriter: rw}
				defer func() {
					if !grw.wroteBody {
						if res.Header().Get(echo.HeaderContentEncoding) == gzipScheme {
							res.Header().Del(echo.HeaderContentEncoding)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.Writer = rw
						w.Reset(ioutil.Discard)
					}
					w.Close()
					pool.Put(w)
				}()
				res.Writer = grw
			}
			return next(c)
		}
	}, nil
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	w.Writer.(*gzip.Writer).Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func gzipCompressPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, err := gzip.NewWriterLevel(ioutil.Discard, config.Level)
			if err != nil {
				return err
			}
			return w
		},
	}
}

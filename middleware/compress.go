package middleware

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

type (
	// GzipConfig defines the config for Gzip middleware.
	GzipConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Gzip compression level.
		// Optional. Default value -1.
		Level int `yaml:"level"`

		// Length threshold before gzip compression
		// is applied. Optional. Default value 0
		MinLength int
	}

	gzipResponseWriter struct {
		io.Writer
		http.ResponseWriter
		wroteBody         bool
		minLength         int
		minLengthExceeded bool
		buffer            bytes.Buffer
		code              int
	}
)

const (
	gzipScheme = "gzip"
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Skipper:   DefaultSkipper,
		Level:     -1,
		MinLength: 0,
	}
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithConfig return Gzip middleware with config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultGzipConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}
	if config.MinLength < 0 {
		config.MinLength = DefaultGzipConfig.MinLength
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
				i := pool.Get()
				w, ok := i.(*gzip.Writer)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}
				rw := res.Writer
				w.Reset(rw)
				grw := &gzipResponseWriter{Writer: w, ResponseWriter: rw, minLength: config.MinLength}
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
					} else if !grw.minLengthExceeded {
						// Write uncompressed response
						res.Writer = rw
						grw.ResponseWriter.WriteHeader(grw.code)
						grw.buffer.WriteTo(rw)
						w.Reset(io.Discard)
					}
					w.Close()
					pool.Put(w)
				}()
				res.Writer = grw
			}
			return next(c)
		}
	}
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444

	// Delay writing of the header until we know if we'll actually compress the response
	w.code = code
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true

	if !w.minLengthExceeded {
		n, err := w.buffer.Write(b)

		if w.buffer.Len() >= w.minLength {
			w.minLengthExceeded = true

			// The minimum length is exceeded, add Content-Encoding header and write the header
			w.Header().Set(echo.HeaderContentEncoding, gzipScheme) // Issue #806
			w.ResponseWriter.WriteHeader(w.code)

			return w.Writer.Write(w.buffer.Bytes())
		} else {
			return n, err
		}
	}

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

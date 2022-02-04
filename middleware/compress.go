package middleware

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

type (
	compressConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Compression level.
		// Optional. Default value -1.
		Level int `yaml:"level"`
	}

	// GzipConfig defines the config for Gzip middleware.
	GzipConfig compressConfig
	// DeflateConfig defines the config for Deflate middleware.
	DeflateConfig compressConfig

	compressResponseWriter struct {
		io.Writer
		http.ResponseWriter
		wroteBody bool
	}

	resetWriteCloser interface {
		Reset(w io.Writer)
		io.WriteCloser
	}

	flusher interface {
		Flush() error
	}
)

const (
	gzipScheme    = "gzip"
	deflateScheme = "deflate"
)

var (
	defaultConfig = compressConfig{
		Skipper: DefaultSkipper,
		Level:   -1,
	}
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultGzipConfig = GzipConfig(defaultConfig)
	// DefaultDeflateConfig is the default Deflate middleware config.
	DefaultDeflateConfig = DeflateConfig(defaultConfig)
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

// Deflate returns a middleware which compresses HTTP response using deflate(zlib) compression
func Deflate() echo.MiddlewareFunc {
	return DeflateWithConfig(DefaultDeflateConfig)
}

// GzipWithConfig return Gzip middleware with config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) echo.MiddlewareFunc {
	return compressWithConfig(compressConfig(config), gzipScheme)
}

// DeflateWithConfig return Deflate middleware with config.
// See: `Deflate()`.
func DeflateWithConfig(config DeflateConfig) echo.MiddlewareFunc {
	return compressWithConfig(compressConfig(config), deflateScheme)
}

func compressWithConfig(config compressConfig, encoding string) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = defaultConfig.Level
	}

	var pool sync.Pool
	switch encoding {
	case gzipScheme:
		pool = gzipCompressPool(config)
	case deflateScheme:
		pool = deflateCompressPool(config)
	default:
		panic("echo: either gzip or deflate is currently supported")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), encoding) {
				res.Header().Set(echo.HeaderContentEncoding, encoding) // Issue #806
				i := pool.Get()
				w, ok := i.(resetWriteCloser)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}
				rw := res.Writer
				w.Reset(rw)
				grw := &compressResponseWriter{Writer: w, ResponseWriter: rw}
				defer func() {
					if !grw.wroteBody {
						if res.Header().Get(echo.HeaderContentEncoding) == encoding {
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
	}
}

func (w *compressResponseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444
	w.ResponseWriter.WriteHeader(code)
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true
	return w.Writer.Write(b)
}

func (w *compressResponseWriter) Flush() {
	w.Writer.(flusher).Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *compressResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *compressResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func gzipCompressPool(config compressConfig) sync.Pool {
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

func deflateCompressPool(config compressConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, err := zlib.NewWriterLevel(ioutil.Discard, config.Level)
			if err != nil {
				return err
			}
			return w
		},
	}
}

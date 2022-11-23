package middleware

import (
	"bufio"
	"compress/gzip"
	"errors"
	"github.com/andybalholm/brotli"
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
	}

	// BrotliConfig defines the config for Brotli middleware.
	BrotliConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Brotli compression level.
		// Optional. Default value -1.
		Level int `yaml:"level"`
	}

	responseWriter struct {
		io.Writer
		http.ResponseWriter
		wroteBody bool
	}
)

const (
	gzipScheme   = "gzip"
	brotliScheme = "br"
	defaultLevel = 6
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Skipper: DefaultSkipper,
		Level:   defaultLevel,
	}
	// DefaultBrotliConfig is the default Brotli middleware config.
	DefaultBrotliConfig = BrotliConfig{
		Skipper: DefaultSkipper,
		Level:   defaultLevel,
	}

	unsupportedCompressEncodingErr = errors.New("unsupported compress content encoding")
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithConfig return Gzip middleware with config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) echo.MiddlewareFunc {
	return CompressWithConfig(gzipScheme, config.Skipper, config.Level)
}

// Brotli returns a middleware which compresses HTTP response using brotli compression
// scheme.
func Brotli() echo.MiddlewareFunc {
	return BrotliWithConfig(DefaultBrotliConfig)
}

// BrotliWithConfig return Brotli middleware with config.
// See: `Brotli()`.
func BrotliWithConfig(config BrotliConfig) echo.MiddlewareFunc {
	return CompressWithConfig(brotliScheme, config.Skipper, config.Level)
}

// CompressWithConfig return Gzip, Brotli middleware according config.
func CompressWithConfig(encode string, skipper Skipper, level int) echo.MiddlewareFunc {
	// Defaults
	if encode == "" {
		encode = gzipScheme
	}
	if skipper == nil {
		skipper = DefaultSkipper
	}
	if level == 0 {
		level = defaultLevel
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), encode) {
				res.Header().Set(echo.HeaderContentEncoding, encode) // Issue #806
				rw := res.Writer
				w, err := acquireCompressWriter(encode, level, rw)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
				}
				grw := &responseWriter{Writer: w, ResponseWriter: rw}
				defer func() {
					if !grw.wroteBody {
						if res.Header().Get(echo.HeaderContentEncoding) == encode {
							res.Header().Del(echo.HeaderContentEncoding)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.Writer = rw
					}
					releaseCompressWriter(encode, w, !grw.wroteBody)
				}()
				res.Writer = grw
			}
			return next(c)
		}
	}
}

func (w *responseWriter) WriteHeader(code int) {
	w.Header().Del(echo.HeaderContentLength) // Issue #444
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true
	return w.Writer.Write(b)
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.Writer.(*gzip.Writer); ok {
		flusher.Flush()
	}
	if flusher, ok := w.Writer.(*brotli.Writer); ok {
		flusher.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

var (
	gzipWriterPool   sync.Pool
	brotliWriterPool sync.Pool
)

func acquireCompressWriter(encode string, level int, w io.Writer) (io.WriteCloser, error) {
	switch encode {
	case brotliScheme:
		v := brotliWriterPool.Get()
		if v == nil {
			bw := brotli.NewWriterLevel(w, level)
			return bw, nil
		}
		bw := v.(*brotli.Writer)
		bw.Reset(w)
		return bw, nil
	case gzipScheme:
		v := gzipWriterPool.Get()
		if v == nil {
			gw, err := gzip.NewWriterLevel(w, level)
			return gw, err
		}
		gw := v.(*gzip.Writer)
		gw.Reset(w)
		return gw, nil
	}
	return nil, unsupportedCompressEncodingErr
}

func releaseCompressWriter(encode string, w io.Closer, reset bool) error {
	var err error
	switch encode {
	case brotliScheme:
		if reset {
			w.(*brotli.Writer).Reset(ioutil.Discard)
		}
		err = w.Close()
		brotliWriterPool.Put(w)
	case gzipScheme:
		if reset {
			w.(*gzip.Writer).Reset(ioutil.Discard)
		}
		err = w.Close()
		gzipWriterPool.Put(w)
	}
	return err
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
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

	// Length threshold before gzip compression is applied.
	// Optional. Default value 0.
	//
	// Most of the time you will not need to change the default. Compressing
	// a short response might increase the transmitted data because of the
	// gzip format overhead. Compressing the response will also consume CPU
	// and time on the server and the client (for decompressing). Depending on
	// your use case such a threshold might be useful.
	//
	// See also:
	// https://webmasters.stackexchange.com/questions/31750/what-is-recommended-minimum-object-size-for-gzip-performance-benefits
	MinLength int
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteHeader       bool
	wroteBody         bool
	minLength         int
	minLengthExceeded bool
	passthrough       bool
	buffer            *bytes.Buffer
	code              int
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
	if config.MinLength < 0 {
		config.MinLength = 0
	}

	pool := gzipCompressPool(config)
	bpool := bufferPool()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			res := c.Response()
			res.Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			var succeeded bool
			if acceptsGzip(c.Request().Header.Values(echo.HeaderAcceptEncoding)) {
				i := pool.Get()
				w, ok := i.(*gzip.Writer)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, "invalid pool object")
				}
				rw := res
				w.Reset(rw)
				buf := bpool.Get().(*bytes.Buffer)
				buf.Reset()

				grw := &gzipResponseWriter{
					Writer:         w,
					ResponseWriter: rw,
					minLength:      config.MinLength,
					buffer:         buf,
				}
				c.SetResponse(grw)
				defer func() {
					// There are different reasons for cases when we have not yet written response to the client and now need to do so.
					// a) handler response had only response code and no response body (ala 404 or redirects etc). Response code need to be written now.
					// b) body is shorter than our minimum length threshold and being buffered currently and needs to be written
					// Responses already encoded by the handler pass through unchanged.
					if grw.passthrough || (succeeded && grw.passThroughIfEncoded()) {
						c.SetResponse(rw)
						w.Reset(io.Discard)
					} else if !grw.wroteBody && !grw.minLengthExceeded {
						// Nothing was compressed, so no coding applies to the body that follows.
						res.Header().Del(echo.HeaderContentEncoding)
						if !succeeded {
							// The error handler writes its own body.
							res.Header().Del(echo.HeaderContentLength)
						}
						if grw.wroteHeader {
							rw.WriteHeader(grw.code)
						}
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						c.SetResponse(rw)
						w.Reset(io.Discard)
					} else if !grw.minLengthExceeded {
						// Write uncompressed response
						c.SetResponse(rw)
						if grw.wroteHeader {
							grw.ResponseWriter.WriteHeader(grw.code)
						}
						_, _ = grw.buffer.WriteTo(rw)
						w.Reset(io.Discard)
					}
					_ = w.Close()
					bpool.Put(buf)
					pool.Put(w)
				}()
			}
			err := next(c)
			succeeded = err == nil
			return err
		}
	}, nil
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		// Informational responses precede the final status and are never compressed.
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if w.passthrough {
		w.ResponseWriter.WriteHeader(code)
		return
	}

	w.wroteHeader = true

	// Delay writing of the header until we know if we'll actually compress the response
	w.code = code
	w.passThroughIfEncoded()
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.passThroughIfEncoded() {
		return w.ResponseWriter.Write(b)
	}

	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}
	w.wroteBody = true

	if !w.minLengthExceeded {
		n, err := w.buffer.Write(b)

		if w.buffer.Len() >= w.minLength {
			w.minLengthExceeded = true

			// The minimum length is exceeded, add Content-Encoding header and write the header
			w.startCompression()

			if _, err := w.Writer.Write(w.buffer.Bytes()); err != nil {
				return 0, err
			}
			// Report only the caller's bytes; the buffer also holds earlier writes.
			return len(b), nil
		}

		return n, err
	}

	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if w.passThroughIfEncoded() {
		// echo.Response.Flush commits the status if nothing was written yet.
		_ = http.NewResponseController(w.ResponseWriter).Flush()
		return
	}

	if !w.minLengthExceeded {
		// Enforce compression because we will not know how much more data will come
		w.minLengthExceeded = true
		w.startCompression()

		_, _ = w.Writer.Write(w.buffer.Bytes())
	}

	if gw, ok := w.Writer.(*gzip.Writer); ok {
		gw.Flush()
	}
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

// startCompression commits gzip headers once the response is known to be compressed.
func (w *gzipResponseWriter) startCompression() {
	w.Header().Set(echo.HeaderContentEncoding, gzipScheme) // Issue #806
	w.Header().Del(echo.HeaderContentLength)               // Issue #444
	if w.wroteHeader {
		w.ResponseWriter.WriteHeader(w.code)
	}
}

// passThroughIfEncoded switches to the underlying writer before gzip handles the body.
func (w *gzipResponseWriter) passThroughIfEncoded() bool {
	if w.passthrough {
		return true
	}
	if w.wroteBody || w.minLengthExceeded {
		return false
	}
	if !hasContentCoding(w.Header().Values(echo.HeaderContentEncoding)) {
		return false
	}
	w.passthrough = true
	if w.wroteHeader {
		w.ResponseWriter.WriteHeader(w.code)
	}
	return true
}

// hasContentCoding reports whether Content-Encoding lists any coding other than identity.
func hasContentCoding(headers []string) bool {
	for _, header := range headers {
		for coding := range strings.SplitSeq(header, ",") {
			coding = strings.TrimSpace(coding)
			if coding != "" && !strings.EqualFold(coding, "identity") {
				return true
			}
		}
	}
	return false
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *gzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func gzipCompressPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() any {
			w, err := gzip.NewWriterLevel(io.Discard, config.Level)
			if err != nil {
				return err
			}
			return w
		},
	}
}

func bufferPool() sync.Pool {
	return sync.Pool{
		New: func() any {
			b := &bytes.Buffer{}
			return b
		},
	}
}

func acceptsGzip(headers []string) bool {
	var gzipAllowed, wildcardAllowed, wildcardRefused bool
	for _, header := range headers {
		for entry := range strings.SplitSeq(header, ",") {
			coding, weight, hasWeight := strings.Cut(entry, ";")
			coding = strings.TrimSpace(coding)
			switch {
			case strings.EqualFold(coding, gzipScheme), strings.EqualFold(coding, "x-gzip"):
				// RFC 9110, section 8.4.1.3: x-gzip is equivalent to gzip.
				// An explicit refusal or invalid weight takes precedence over
				// wildcards and over any duplicate entry allowing gzip.
				if hasWeight && !positiveQualityValue(weight) {
					return false
				}
				gzipAllowed = true
			case coding == "*":
				if hasWeight && !positiveQualityValue(weight) {
					wildcardRefused = true
				} else {
					wildcardAllowed = true
				}
			}
		}
	}
	return gzipAllowed || (wildcardAllowed && !wildcardRefused)
}

// positiveQualityValue matches the qvalue grammar in RFC 9110, section 12.4.2.
func positiveQualityValue(weight string) bool {
	name, value, ok := strings.Cut(strings.TrimSpace(weight), "=")
	if !ok || !strings.EqualFold(name, "q") {
		return false
	}
	whole, fraction, _ := strings.Cut(value, ".")
	if (whole != "0" && whole != "1") || len(fraction) > 3 {
		return false
	}
	positive := whole == "1"
	for i := range len(fraction) {
		digit := fraction[i]
		if digit < '0' || digit > '9' || (whole == "1" && digit != '0') {
			return false
		}
		positive = positive || digit != '0'
	}
	return positive
}

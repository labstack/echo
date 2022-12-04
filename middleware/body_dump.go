package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/labstack/echo/v5"
)

// BodyDumpConfig defines the config for BodyDump middleware.
type BodyDumpConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Handler receives request and response payload.
	// Required.
	Handler BodyDumpHandler
}

// BodyDumpHandler receives the request and response payload.
type BodyDumpHandler func(c echo.Context, reqBody []byte, resBody []byte)

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// BodyDump returns a BodyDump middleware.
//
// BodyDump middleware captures the request and response payload and calls the
// registered handler.
func BodyDump(handler BodyDumpHandler) echo.MiddlewareFunc {
	return BodyDumpWithConfig(BodyDumpConfig{Handler: handler})
}

// BodyDumpWithConfig returns a BodyDump middleware with config.
// See: `BodyDump()`.
func BodyDumpWithConfig(config BodyDumpConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts BodyDumpConfig to middleware or returns an error for invalid configuration
func (config BodyDumpConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Handler == nil {
		return nil, errors.New("echo body-dump middleware requires a handler function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Request
			reqBody := []byte{}
			if c.Request().Body != nil {
				reqBody, _ = io.ReadAll(c.Request().Body)
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

			// Response
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer

			err := next(c)

			// Callback
			config.Handler(c, reqBody, resBody.Bytes())

			return err
		}
	}, nil
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

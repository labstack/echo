// +build go1.13

package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"time"
)

type (
	// TimeoutConfig defines the config for Timeout middleware.
	TimeoutConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// OnTimeoutRouteErrorHandler is an error handler that is executed for error that was returned from wrapped route after
		// request timeouted and we already had sent the error code (503) and message response to the client.
		// NB: do not write headers/body inside this handler. The response has already been sent to the client and response writer
		// will not accept anything no more. If you want to know what actual route middleware timeouted use `c.Path()`
		OnTimeoutRouteErrorHandler func(err error, c echo.Context)

		// ErrorMessage is written to response on timeout in addition to http.StatusServiceUnavailable (503) status code
		// It can be used to define a custom timeout error message
		// DEPRECATED: do not use. Use `DefaultTimeoutErrorHandler` instead
		ErrorMessage string

		// DefaultTimeoutErrorHandler is an error handler that is executed when handler context timeouts or is cancelled.
		// Overwrite this with our custom implementation when want to send your own custom error code etc.
		// NB: OnTimeoutRouteErrorHandler will still be called when handler function eventually finishes.
		DefaultTimeoutErrorHandler func(contextError error, c echo.Context) error

		// Timeout configures a timeout for the middleware, defaults to 0 for no timeout
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		Timeout time.Duration
	}
)

var (
	// DefaultTimeoutConfig is the default Timeout middleware config.
	DefaultTimeoutConfig = TimeoutConfig{
		Skipper: DefaultSkipper,
		Timeout: 0,
	}
)

// Timeout returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Timeout() echo.MiddlewareFunc {
	return TimeoutWithConfig(DefaultTimeoutConfig)
}

// TimeoutWithConfig returns a Timeout middleware with config.
// See: `Timeout()`.
func TimeoutWithConfig(config TimeoutConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTimeoutConfig.Skipper
	}
	if config.ErrorMessage != "" && config.DefaultTimeoutErrorHandler == nil {
		config.DefaultTimeoutErrorHandler = func(contextError error, c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, config.ErrorMessage)
		}
	}
	if config.DefaultTimeoutErrorHandler == nil {
		config.DefaultTimeoutErrorHandler = func(contextError error, c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, contextError.Error())
		}
	}
	if config.OnTimeoutRouteErrorHandler == nil {
		config.OnTimeoutRouteErrorHandler = func(err error, c echo.Context) {
			if c.Logger() == nil {
				return
			}
			c.Logger().Errorf("handler for path %v returned an error after request timeouted", c.Path())
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.Timeout == 0 {
				return next(c)
			}

			ctx, cancelCtx := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancelCtx()

			r := c.Request().WithContext(ctx)

			tw := &timeoutWriter{
				w:   c.Response(),
				h:   make(map[string][]string),
				req: r,
			}
			originalWriter := c.Response().Writer
			c.Response().Writer = tw

			done := make(chan error)
			panicChan := make(chan interface{})
			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicChan <- p
					}
				}()

				err := next(c)
				// NB: when timeout occurs this functions returns after timeout response has already been sent to the client
				// only thing we can do with error is to log it as we can not send anything to client anymore
				select {
				case <-ctx.Done():
					config.OnTimeoutRouteErrorHandler(err, c)
				default:
					done <- err
				}
				close(done)
			}()

			select {
			case p := <-panicChan:
				// restore writer for context so error handler can write response to request
				c.Response().Writer = originalWriter

				panic(p)
			case err := <-done: // where are here before timeout has been exceeded
				tw.mu.Lock()
				defer tw.mu.Unlock()

				// restore writer for context
				c.Response().Writer = originalWriter
				dst := c.Response().Header()
				for k, vv := range tw.h {
					dst[k] = vv
				}
				// as handler could call `c.JSON` etc methods that write to response and therefore commit request we need unlock
				// response and set values from timeoutWriter had wrapped
				if c.Response().Committed {
					c.Response().Committed = false
					c.Response().WriteHeader(tw.code)
					c.Response().Committed = true
				}
				if tw.wbuf.Len() != 0 {
					c.Response().Write(tw.wbuf.Bytes())
				}

				return err
			case <-ctx.Done(): // where are here after timeout has been exceeded or context was cancelled
				tw.mu.Lock()
				defer tw.mu.Unlock()

				// restore writer for context so error handler can write response to request
				c.Response().Writer = originalWriter

				tw.timedOut = true
				return config.DefaultTimeoutErrorHandler(ctx.Err(), c)
			}
		}
	}
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = errors.New("http: Handler timeout")

// NOTE: this is copy of http.timeoutWriter
type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer
	req  *http.Request

	logger echo.Logger

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

// Push implements the Pusher interface.
func (tw *timeoutWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := tw.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.writeHeaderLocked(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) writeHeaderLocked(code int) {
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}

	switch {
	case tw.timedOut:
		return
	case tw.wroteHeader:
		if tw.req != nil {
			if tw.logger != nil {
				tw.logger.Errorf("http: superfluous response.WriteHeader call: %v", tw.req.RequestURI)
			} else {
				fmt.Printf("http: superfluous response.WriteHeader call: %v", tw.req.RequestURI)
			}
		}
	default:
		tw.wroteHeader = true
		tw.code = code
	}
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.writeHeaderLocked(code)
}

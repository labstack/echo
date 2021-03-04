// +build go1.13

package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type (
	// TimeoutConfig defines the config for Timeout middleware.
	TimeoutConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// ErrorMessage is written to response on timeout in addition to http.StatusServiceUnavailable (503) status code
		// It can be used to define a custom timeout error message
		ErrorMessage string

		// OnTimeoutRouteErrorHandler is an error handler that is executed for error that was returned from wrapped route after
		// request timeouted and we already had sent the error code (503) and message response to the client.
		// NB: do not write headers/body inside this handler. The response has already been sent to the client and response writer
		// will not accept anything no more. If you want to know what actual route middleware timeouted use `c.Path()`
		OnTimeoutRouteErrorHandler func(err error, c echo.Context)

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
		Skipper:      DefaultSkipper,
		Timeout:      0,
		ErrorMessage: "",
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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.Timeout == 0 {
				return next(c)
			}

			handlerWrapper := echoHandlerFuncWrapper{
				ctx:        c,
				handler:    next,
				errChan:    make(chan error, 1),
				errHandler: config.OnTimeoutRouteErrorHandler,
			}
			handler := http.TimeoutHandler(handlerWrapper, config.Timeout, config.ErrorMessage)
			handler.ServeHTTP(c.Response().Writer, c.Request())

			select {
			case err := <-handlerWrapper.errChan:
				return err
			default:
				return nil
			}
		}
	}
}

type echoHandlerFuncWrapper struct {
	ctx        echo.Context
	handler    echo.HandlerFunc
	errHandler func(err error, c echo.Context)
	errChan    chan error
}

func (t echoHandlerFuncWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// replace writer with TimeoutHandler custom one. This will guarantee that
	// `writes by h to its ResponseWriter will return ErrHandlerTimeout.`
	originalWriter := t.ctx.Response().Writer
	t.ctx.Response().Writer = rw

	err := t.handler(t.ctx)
	if ctxErr := r.Context().Err(); ctxErr == context.DeadlineExceeded {
		if err != nil && t.errHandler != nil {
			t.errHandler(err, t.ctx)
		}
		return // on timeout we can not send handler error to client because `http.TimeoutHandler` has already sent headers
	}
	// we restore original writer only for cases we did not timeout. On timeout we have already sent response to client
	// and should not anymore send additional headers/data
	// so on timeout writer stays what http.TimeoutHandler uses and prevents writing headers/body
	t.ctx.Response().Writer = originalWriter
	if err != nil {
		t.errChan <- err
	}
}

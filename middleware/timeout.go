// +build go1.13

package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"time"
)

type (
	// TimeoutConfig defines the config for Timeout middleware.
	TimeoutConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// ErrorHandler defines a function which is executed for a timeout
		// It can be used to define a custom timeout error
		ErrorHandler TimeoutErrorHandlerWithContext
		// Timeout configures a timeout for the middleware, defaults to 0 for no timeout
		Timeout time.Duration
	}

	// TimeoutErrorHandlerWithContext is an error handler that is used with the timeout middleware so we can
	// handle the error as we see fit
	TimeoutErrorHandlerWithContext func(error, echo.Context) error
)

var (
	// DefaultTimeoutConfig is the default Timeout middleware config.
	DefaultTimeoutConfig = TimeoutConfig{
		Skipper:      DefaultSkipper,
		Timeout:      0,
		ErrorHandler: nil,
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

			ctx, cancel := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancel()

			// this does a deep clone of the context, wondering if there is a better way to do this?
			c.SetRequest(c.Request().Clone(ctx))

			done := make(chan error, 1)
			go func() {
				// This goroutine will keep running even if this middleware times out and
				// will be stopped when ctx.Done() is called down the next(c) call chain
				done <- next(c)
			}()

			select {
			case <-ctx.Done():
				if config.ErrorHandler != nil {
					return config.ErrorHandler(ctx.Err(), c)
				}
				return ctx.Err()
			case err := <-done:
				return err
			}
		}
	}
}

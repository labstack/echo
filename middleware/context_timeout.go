package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/echo/v4"
)

// Simply return or wrap the error that context passed method returns.
// Note: Injects ContextWithTimeout to c.Request().Context(). Does not work with kong running processes without context.
//
// e.GET("/", func(c echo.Context) error {
// 	sleepTime := time.Duration(2 * time.Second)
//
// 	if err := sleepWithContext(c.Request().Context(), sleepTime); err != nil {
// 		return fmt.Errorf("%w: execution error", err)
// 	}
//
// 	return c.String(http.StatusOK, "Hello, World!")
// })
//

// ContextTimeoutConfig defines the config for ContextTimeout middleware.
type ContextTimeoutConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// ErrorHandler is a function when error aries in middeware execution.
	ErrorHandler func(err error, c echo.Context) error

	// Timeout configures a timeout for the middleware, defaults to 0 for no timeout
	Timeout time.Duration
}

var (
	// DefaultContextTimeoutErrorHandler is default error handler of ContextTimeout middleware.
	DefaultContextTimeoutErrorHandler = func(err error, c echo.Context) error {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return echo.ErrServiceUnavailable
			}
			return err
		}
		return nil
	}

	// DefaultContextTimeoutConfig is the default ContextTimeoutConfig middleware config.
	DefaultContextTimeoutConfig = ContextTimeoutConfig{
		Skipper:      DefaultSkipper,
		Timeout:      0,
		ErrorHandler: DefaultContextTimeoutErrorHandler,
	}
)

// ContextTimeout returns a middleware which returns error (503 Service Unavailable error) to client
// when underlying method returns context.DeadlineExceeded error.
func ContextTimeout() echo.MiddlewareFunc {
	return ContextTimeoutWithConfig(DefaultContextTimeoutConfig)
}

// ContextTimeoutWithConfig returns a Timeout middleware with config.
func ContextTimeoutWithConfig(config ContextTimeoutConfig) echo.MiddlewareFunc {
	return config.ToMiddleware()
}

// ToMiddleware converts Config to middleware.
func (config ContextTimeoutConfig) ToMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if (config.Skipper != nil && config.Skipper(c)) || config.Timeout == 0 {
				return next(c)
			}

			timeoutContext, cancel := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(timeoutContext))

			err := next(c)
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(err, c)
				} else {
					return DefaultContextTimeoutErrorHandler(err, c)
				}
			}
			return nil
		}
	}
}

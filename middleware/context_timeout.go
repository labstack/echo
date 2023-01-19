package middleware

import (
	"context"
	"errors"
	"net/http"
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

	// ErrorMessage is written to response on timeout in addition to http.StatusServiceUnavailable (503) status code
	// It can be used to define a custom timeout error message
	ErrorMessage string

	// Timeout configures a timeout for the middleware, defaults to 0 for no timeout
	Timeout time.Duration
}

var (
	// DefaultContextTimeoutConfig is the default ContextTimeoutConfig middleware config.
	DefaultContextTimeoutConfig = ContextTimeoutConfig{
		Skipper:      DefaultSkipper,
		Timeout:      0,
		ErrorMessage: "",
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
	if config.Skipper == nil {
		config.Skipper = DefaultTimeoutConfig.Skipper
	}

	suResponse := echo.Map{"message": http.StatusText(http.StatusServiceUnavailable)}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.Timeout == 0 {
				return next(c)
			}

			timeoutContext, cancel := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancel()

			timeoutRequest := c.Request().WithContext(timeoutContext)

			c.SetRequest(timeoutRequest)

			err := next(c)

			if err == nil {
				err = c.Request().Context().Err()
			}

			if err != nil && !c.Response().Committed {
				if errors.Is(err, context.DeadlineExceeded) {
					c.Logger().Error("http: Handler timeout")
					if config.ErrorMessage == "" {
						return c.JSON(http.StatusServiceUnavailable, suResponse)
					} else {
						return c.String(http.StatusServiceUnavailable, config.ErrorMessage)
					}
				}

				return err
			}
			return nil

		}
	}
}

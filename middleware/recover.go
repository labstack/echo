// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v5"
)

// RecoverConfig defines the config for Recover middleware.
type RecoverConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Size of the stack to be printed.
	// Optional. Default value 4KB.
	StackSize int

	// DisableStackAll disables formatting stack traces of all other goroutines
	// into buffer after the trace for the current goroutine.
	// Optional. Default value false.
	DisableStackAll bool

	// DisablePrintStack disables printing stack trace.
	// Optional. Default value as false.
	DisablePrintStack bool
}

// DefaultRecoverConfig is the default Recover middleware config.
var DefaultRecoverConfig = RecoverConfig{
	Skipper:           DefaultSkipper,
	StackSize:         4 << 10, // 4 KB
	DisableStackAll:   false,
	DisablePrintStack: false,
}

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover() echo.MiddlewareFunc {
	return RecoverWithConfig(DefaultRecoverConfig)
}

// RecoverWithConfig returns a Recovery middleware with config or panics on invalid configuration.
func RecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts RecoverConfig to middleware or returns an error for invalid configuration
func (config RecoverConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					if r == http.ErrAbortHandler {
						panic(r)
					}
					tmpErr, ok := r.(error)
					if !ok {
						tmpErr = fmt.Errorf("%v", r)
					}
					if !config.DisablePrintStack {
						stack := make([]byte, config.StackSize)
						length := runtime.Stack(stack, !config.DisableStackAll)
						tmpErr = &PanicStackError{Stack: stack[:length], Err: tmpErr}
					}
					err = tmpErr
				}
			}()
			return next(c)
		}
	}, nil
}

// PanicStackError is an error type that wraps an error along with its stack trace.
// It is returned when config.DisablePrintStack is set to false.
type PanicStackError struct {
	Stack []byte
	Err   error
}

func (e *PanicStackError) Error() string {
	return fmt.Sprintf("[PANIC RECOVER] %s %s", e.Err.Error(), e.Stack)
}

func (e *PanicStackError) Unwrap() error {
	return e.Err
}

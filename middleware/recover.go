package middleware

import (
	"fmt"
	"runtime"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

type (
	// RecoverConfig defines the config for recover middleware.
	RecoverConfig struct {
		// StackSize is the stack size to be printed.
		// Optional, with default value as 4 KB.
		StackSize int

		// DisableStackAll disables formatting stack traces of all other goroutines
		// into buffer after the trace for the current goroutine.
		// Optional, with default value as false.
		DisableStackAll bool

		// DisablePrintStack disables printing stack trace.
		// Optional, with default value as false.
		DisablePrintStack bool
	}
)

var (
	// DefaultRecoverConfig is the default recover middleware config.
	DefaultRecoverConfig = RecoverConfig{
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover() echo.MiddlewareFunc {
	return RecoverWithConfig(DefaultRecoverConfig)
}

// RecoverWithConfig returns a recover middleware from config.
// See `Recover()`.
func RecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.StackSize == 0 {
		config.StackSize = DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					switch r := r.(type) {
					case error:
						err = r
					default:
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						c.Logger().Printf("[%s] %s %s", color.Red("PANIC RECOVER"), err, stack[:length])
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

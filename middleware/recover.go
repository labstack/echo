package middleware

import (
	"fmt"
	"runtime"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

type (
	// RecoverConfig defines config for recover middleware.
	RecoverConfig struct {
		// StackSize is the stack size to be printed.
		StackSize int

		// StackAll is a flag to format stack traces of all other goroutines into
		// buffer after the trace for the current goroutine, or not.
		StackAll bool

		// PrintStack is a flag to print stack or not.
		PrintStack bool
	}
)

var (
	// DefaultRecoverConfig is the default recover middleware config.
	DefaultRecoverConfig = RecoverConfig{
		StackSize:  4 << 10, // 4 KB
		StackAll:   true,
		PrintStack: true,
	}
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover() echo.MiddlewareFunc {
	return RecoverFromConfig(DefaultRecoverConfig)
}

// RecoverFromConfig returns a recover middleware from config.
// See `Recover()`.
func RecoverFromConfig(config RecoverConfig) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
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
					length := runtime.Stack(stack, config.StackAll)
					if config.PrintStack {
						c.Logger().Printf("[%s] %s %s", color.Red("PANIC RECOVER"), err, stack[:length])
					}
					c.Error(err)
				}
			}()
			return next.Handle(c)
		})
	}
}

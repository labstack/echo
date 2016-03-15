package middleware

import (
	"fmt"
	"runtime"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

type (
	RecoverConfig struct {
		StackSize  int
		StackAll   bool
		PrintStack bool
	}
)

var (
	DefaultRecoverConfig = RecoverConfig{
		StackSize:  4 << 10, // 4 KB
		StackAll:   true,
		PrintStack: true,
	}
)

func Recover() echo.MiddlewareFunc {
	return RecoverWithConfig(DefaultRecoverConfig)
}

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func RecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
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
						c.Logger().Printf("%s|%s", color.Red("PANIC RECOVER"), stack[:length])
					}
					c.Error(err)
				}
			}()
			return next.Handle(c)
		})
	}
}

package middleware

import (
	"errors"

	"github.com/labstack/echo"
)

type (
	RecoverOptions struct {
	}
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover(options ...*RecoverOptions) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		// TODO: Provide better stack trace
		// - `https://github.com/go-errors/errors`
		// - `https://github.com/docker/libcontainer/tree/master/stacktrace`
		return echo.HandlerFunc(func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					e := ""
					switch r := r.(type) {
					case string:
						e = r
					case error:
						e = r.Error()
					default:
						e = "unknown error"
					}
					c.Error(errors.New("panic recover|" + e))
				}
			}()
			return next.Handle(c)
		})
	}
}

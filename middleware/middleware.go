package middleware

import "github.com/labstack/echo"

type (
	Middleware interface {
		Process(echo.HandlerFunc) echo.HandlerFunc
	}

	MiddlewareFunc func(echo.HandlerFunc) echo.HandlerFunc
)

func (f MiddlewareFunc) Process(h echo.HandlerFunc) echo.HandlerFunc {
	return f(h)
}

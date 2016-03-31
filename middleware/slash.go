package middleware

import (
	"github.com/labstack/echo"
)

// AddTrailingSlash returns a root level (before router) middleware which adds a
// trailing slash to the request `URL#Path`.
//
// Usage `Echo#Pre(AddTrailingSlash())`
func AddTrailingSlash() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			url := c.Request().URL()
			path := url.Path()
			if path != "/" && path[len(path)-1] != '/' {
				url.SetPath(path + "/")
			}
			return next.Handle(c)
		})
	}
}

// RemoveTrailingSlash returns a root level (before router) middleware which removes
// a trailing slash from the request URI.
//
// Usage `Echo#Pre(RemoveTrailingSlash())`
func RemoveTrailingSlash() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			url := c.Request().URL()
			path := url.Path()
			l := len(path) - 1
			if path != "/" && path[l] == '/' {
				url.SetPath(path[:l])
			}
			return next.Handle(c)
		})
	}
}

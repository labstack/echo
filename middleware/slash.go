package middleware

import "github.com/labstack/echo"

// StripTrailingSlash returns a middleware which removes trailing slash from request
// path.
func StripTrailingSlash() echo.HandlerFunc {
	return func(c *echo.Context) error {
		p := c.Request().URL.Path
		l := len(p)
		if p[l-1] == '/' {
			c.Request().URL.Path = p[:l-1]
		}
		return nil
	}
}

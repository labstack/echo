package middleware

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	RedirectToSlashOptions struct {
		Code int
	}
)

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

// RedirectToSlash returns a middleware which redirects requests without trailing
// slash path to trailing slash path.
func RedirectToSlash(opts ...RedirectToSlashOptions) echo.HandlerFunc {
	code := http.StatusMovedPermanently

	if len(opts) > 0 {
		o := opts[0]
		code = o.Code
	}

	return func(c *echo.Context) error {
		p := c.Request().URL.Path
		l := len(p)
		if p[l-1] != '/' {
			c.Redirect(code, p+"/")
		}
		return nil
	}
}

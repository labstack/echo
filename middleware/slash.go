package middleware

import (
	"github.com/labstack/echo"
	"net/http"
)

type (
	RedirectToSlashOptions struct {
		Code int
	}
)

// StripTrailingSlash removes trailing slash from request path.
func StripTrailingSlash() echo.HandlerFunc {
	return func(c *echo.Context) *echo.HTTPError {
		p := c.Request.URL.Path
		l := len(p)
		if p[l-1] == '/' {
			c.Request.URL.Path = p[:l-1]
		}
		return nil
	}
}

// RedirectToSlash redirects requests without trailing slash path to trailing slash
// path, with .
func RedirectToSlash(opts ...RedirectToSlashOptions) echo.HandlerFunc {
	code := http.StatusMovedPermanently

	for _, o := range opts {
		if o.Code != 0 {
			code = o.Code
		}
	}

	return func(c *echo.Context) (he *echo.HTTPError) {
		p := c.Request.URL.Path
		l := len(p)
		if p[l-1] != '/' {
			c.Redirect(code, p+"/")
		}
		return nil
	}
}

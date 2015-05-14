package middleware

import "github.com/labstack/echo"

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
// path, with status code.
func RedirectToSlash(code int) echo.HandlerFunc {
	return func(c *echo.Context) (he *echo.HTTPError) {
		p := c.Request.URL.Path
		l := len(p)
		if p[l-1] != '/' {
			c.Redirect(code, p+"/")
		}
		return nil
	}
}

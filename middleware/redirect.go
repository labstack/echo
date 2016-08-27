package middleware

import (
	"net/http"

	"github.com/labstack/echo"
)

// HTTPSRedirect redirects HTTP requests to HTTPS.
// For example, http://labstack.com will be redirect to https://labstack.com.
func HTTPSRedirect() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			host := req.Host()
			uri := req.URI()
			if !req.IsTLS() {
				return c.Redirect(http.StatusMovedPermanently, "https://"+host+uri)
			}
			return next(c)
		}
	}
}

// HTTPSWWWRedirect redirects HTTP requests to WWW HTTPS.
// For example, http://labstack.com will be redirect to https://www.labstack.com.
func HTTPSWWWRedirect() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			host := req.Host()
			uri := req.URI()
			if !req.IsTLS() && host[:3] != "www" {
				return c.Redirect(http.StatusMovedPermanently, "https://www."+host+uri)
			}
			return next(c)
		}
	}
}

// WWWRedirect redirects non WWW requests to WWW.
// For example, http://labstack.com will be redirect to http://www.labstack.com.
func WWWRedirect() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			scheme := req.Scheme()
			host := req.Host()
			if host[:3] != "www" {
				uri := req.URI()
				return c.Redirect(http.StatusMovedPermanently, scheme+"://www."+host+uri)
			}
			return next(c)
		}
	}
}

// NonWWWRedirect redirects WWW request to non WWW.
// For example, http://www.labstack.com will be redirect to http://labstack.com.
func NonWWWRedirect() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			scheme := req.Scheme()
			host := req.Host()
			if host[:3] == "www" {
				uri := req.URI()
				return c.Redirect(http.StatusMovedPermanently, scheme+"://"+host[4:]+uri)
			}
			return next(c)
		}
	}
}

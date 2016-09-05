package middleware

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	// RedirectConfig defines the config for Redirect middleware.
	RedirectConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Status code to be used when redirecting the request.
		// Optional. Default value http.StatusMovedPermanently.
		Code int `json:"code"`
	}
)

var (
	// DefaultRedirectConfig is the default Redirect middleware config.
	DefaultRedirectConfig = RedirectConfig{
		Skipper: defaultSkipper,
		Code:    http.StatusMovedPermanently,
	}
)

// HTTPSRedirect redirects HTTP requests to HTTPS.
// For example, http://labstack.com will be redirect to https://labstack.com.
//
// Usage `Echo#Pre(HTTPSRedirect())`
func HTTPSRedirect() echo.MiddlewareFunc {
	return HTTPSRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSRedirectWithConfig returns a HTTPSRedirect middleware with config.
// See `HTTPSRedirect()`.
func HTTPSRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			host := req.Host()
			uri := req.URI()
			if !req.IsTLS() {
				return c.Redirect(config.Code, "https://"+host+uri)
			}
			return next(c)
		}
	}
}

// HTTPSWWWRedirect redirects HTTP requests to WWW HTTPS.
// For example, http://labstack.com will be redirect to https://www.labstack.com.
//
// Usage `Echo#Pre(HTTPSWWWRedirect())`
func HTTPSWWWRedirect() echo.MiddlewareFunc {
	return HTTPSWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSWWWRedirectWithConfig returns a HTTPSRedirect middleware with config.
// See `HTTPSWWWRedirect()`.
func HTTPSWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

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
//
// Usage `Echo#Pre(WWWRedirect())`
func WWWRedirect() echo.MiddlewareFunc {
	return WWWRedirectWithConfig(DefaultRedirectConfig)
}

// WWWRedirectWithConfig returns a HTTPSRedirect middleware with config.
// See `WWWRedirect()`.
func WWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

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

// NonWWWRedirect redirects WWW requests to non WWW.
// For example, http://www.labstack.com will be redirect to http://labstack.com.
//
// Usage `Echo#Pre(NonWWWRedirect())`
func NonWWWRedirect() echo.MiddlewareFunc {
	return NonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// NonWWWRedirectWithConfig returns a HTTPSRedirect middleware with config.
// See `NonWWWRedirect()`.
func NonWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

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

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

const (
	www = "www"
)

var (
	// DefaultRedirectConfig is the default Redirect middleware config.
	DefaultRedirectConfig = RedirectConfig{
		Skipper: DefaultSkipper,
		Code:    http.StatusMovedPermanently,
	}
)

// HTTPSRedirect redirects http requests to https.
// For example, http://labstack.com will be redirect to https://labstack.com.
//
// Usage `Echo#Pre(HTTPSRedirect())`
func HTTPSRedirect() echo.MiddlewareFunc {
	return HTTPSRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSRedirectWithConfig returns an HTTPSRedirect middleware with config.
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
			host := req.Host
			uri := req.RequestURI
			if !c.IsTLS() {
				return c.Redirect(config.Code, "https://"+host+uri)
			}
			return next(c)
		}
	}
}

// HTTPSWWWRedirect redirects http requests to https www.
// For example, http://labstack.com will be redirect to https://www.labstack.com.
//
// Usage `Echo#Pre(HTTPSWWWRedirect())`
func HTTPSWWWRedirect() echo.MiddlewareFunc {
	return HTTPSWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
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
			host := req.Host
			uri := req.RequestURI
			if !c.IsTLS() && host[:3] != www {
				return c.Redirect(config.Code, "https://www."+host+uri)
			}
			return next(c)
		}
	}
}

// HTTPSNonWWWRedirect redirects http requests to https non www.
// For example, http://www.labstack.com will be redirect to https://labstack.com.
//
// Usage `Echo#Pre(HTTPSNonWWWRedirect())`
func HTTPSNonWWWRedirect() echo.MiddlewareFunc {
	return HTTPSNonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSNonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSNonWWWRedirect()`.
func HTTPSNonWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
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
			host := req.Host
			uri := req.RequestURI
			if !c.IsTLS() {
				if host[:3] == www {
					return c.Redirect(config.Code, "https://"+host[4:]+uri)
				}
				return c.Redirect(config.Code, "https://"+host+uri)
			}
			return next(c)
		}
	}
}

// WWWRedirect redirects non www requests to www.
// For example, http://labstack.com will be redirect to http://www.labstack.com.
//
// Usage `Echo#Pre(WWWRedirect())`
func WWWRedirect() echo.MiddlewareFunc {
	return WWWRedirectWithConfig(DefaultRedirectConfig)
}

// WWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
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
			scheme := c.Scheme()
			host := req.Host
			if host[:3] != www {
				uri := req.RequestURI
				return c.Redirect(config.Code, scheme+"://www."+host+uri)
			}
			return next(c)
		}
	}
}

// NonWWWRedirect redirects www requests to non www.
// For example, http://www.labstack.com will be redirect to http://labstack.com.
//
// Usage `Echo#Pre(NonWWWRedirect())`
func NonWWWRedirect() echo.MiddlewareFunc {
	return NonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// NonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
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
			scheme := c.Scheme()
			host := req.Host
			if host[:3] == www {
				uri := req.RequestURI
				return c.Redirect(config.Code, scheme+"://"+host[4:]+uri)
			}
			return next(c)
		}
	}
}

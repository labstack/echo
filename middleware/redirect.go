package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// RedirectConfig defines the config for Redirect middleware.
type RedirectConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper

	// Status code to be used when redirecting the request.
	// Optional. Default value http.StatusMovedPermanently.
	Code int

	redirect redirectLogic
}

// redirectLogic represents a function that given a scheme, host and uri
// can both: 1) determine if redirect is needed (will set ok accordingly) and
// 2) return the appropriate redirect url.
type redirectLogic func(scheme, host, uri string) (ok bool, url string)

const www = "www."

// RedirectHTTPSConfig is the HTTPS Redirect middleware config.
var RedirectHTTPSConfig = RedirectConfig{redirect: redirectHTTPS}

// RedirectHTTPSWWWConfig is the HTTPS WWW Redirect middleware config.
var RedirectHTTPSWWWConfig = RedirectConfig{redirect: redirectHTTPSWWW}

// RedirectNonHTTPSWWWConfig is the non HTTPS WWW Redirect middleware config.
var RedirectNonHTTPSWWWConfig = RedirectConfig{redirect: redirectNonHTTPSWWW}

// RedirectWWWConfig is the WWW Redirect middleware config.
var RedirectWWWConfig = RedirectConfig{redirect: redirectWWW}

// RedirectNonWWWConfig is the non WWW Redirect middleware config.
var RedirectNonWWWConfig = RedirectConfig{redirect: redirectNonWWW}

// HTTPSRedirect redirects http requests to https.
// For example, http://labstack.com will be redirect to https://labstack.com.
//
// Usage `Echo#Pre(HTTPSRedirect())`
func HTTPSRedirect() echo.MiddlewareFunc {
	return HTTPSRedirectWithConfig(RedirectHTTPSConfig)
}

// HTTPSRedirectWithConfig returns a HTTPS redirect middleware with config or panics on invalid configuration.
func HTTPSRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	config.redirect = redirectHTTPS
	return toMiddlewareOrPanic(config)
}

// HTTPSWWWRedirect redirects http requests to https www.
// For example, http://labstack.com will be redirect to https://www.labstack.com.
//
// Usage `Echo#Pre(HTTPSWWWRedirect())`
func HTTPSWWWRedirect() echo.MiddlewareFunc {
	return HTTPSWWWRedirectWithConfig(RedirectHTTPSWWWConfig)
}

// HTTPSWWWRedirectWithConfig returns a HTTPS WWW redirect middleware with config or panics on invalid configuration.
func HTTPSWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	config.redirect = redirectHTTPSWWW
	return toMiddlewareOrPanic(config)
}

// HTTPSNonWWWRedirect redirects http requests to https non www.
// For example, http://www.labstack.com will be redirect to https://labstack.com.
//
// Usage `Echo#Pre(HTTPSNonWWWRedirect())`
func HTTPSNonWWWRedirect() echo.MiddlewareFunc {
	return HTTPSNonWWWRedirectWithConfig(RedirectNonHTTPSWWWConfig)
}

// HTTPSNonWWWRedirectWithConfig returns a HTTPS Non-WWW redirect middleware with config or panics on invalid configuration.
func HTTPSNonWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	config.redirect = redirectNonHTTPSWWW
	return toMiddlewareOrPanic(config)
}

// WWWRedirect redirects non www requests to www.
// For example, http://labstack.com will be redirect to http://www.labstack.com.
//
// Usage `Echo#Pre(WWWRedirect())`
func WWWRedirect() echo.MiddlewareFunc {
	return WWWRedirectWithConfig(RedirectWWWConfig)
}

// WWWRedirectWithConfig returns a WWW redirect middleware with config or panics on invalid configuration.
func WWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	config.redirect = redirectWWW
	return toMiddlewareOrPanic(config)
}

// NonWWWRedirect redirects www requests to non www.
// For example, http://www.labstack.com will be redirect to http://labstack.com.
//
// Usage `Echo#Pre(NonWWWRedirect())`
func NonWWWRedirect() echo.MiddlewareFunc {
	return NonWWWRedirectWithConfig(RedirectNonWWWConfig)
}

// NonWWWRedirectWithConfig returns a Non-WWW redirect middleware with config or panics on invalid configuration.
func NonWWWRedirectWithConfig(config RedirectConfig) echo.MiddlewareFunc {
	config.redirect = redirectNonWWW
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts RedirectConfig to middleware or returns an error for invalid configuration
func (config RedirectConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Code == 0 {
		config.Code = http.StatusMovedPermanently
	}
	if config.redirect == nil {
		return nil, errors.New("redirectConfig is missing redirect function")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req, scheme := c.Request(), c.Scheme()
			host := req.Host
			if ok, url := config.redirect(scheme, host, req.RequestURI); ok {
				return c.Redirect(config.Code, url)
			}

			return next(c)
		}
	}, nil
}

var redirectHTTPS = func(scheme, host, uri string) (bool, string) {
	if scheme != "https" {
		return true, "https://" + host + uri
	}
	return false, ""
}

var redirectHTTPSWWW = func(scheme, host, uri string) (bool, string) {
	if scheme != "https" && !strings.HasPrefix(host, www) {
		return true, "https://www." + host + uri
	}
	return false, ""
}

var redirectNonHTTPSWWW = func(scheme, host, uri string) (ok bool, url string) {
	if scheme != "https" {
		host = strings.TrimPrefix(host, www)
		return true, "https://" + host + uri
	}
	return false, ""
}

var redirectWWW = func(scheme, host, uri string) (bool, string) {
	if !strings.HasPrefix(host, www) {
		return true, scheme + "://www." + host + uri
	}
	return false, ""
}

var redirectNonWWW = func(scheme, host, uri string) (bool, string) {
	if strings.HasPrefix(host, www) {
		return true, scheme + "://" + host[4:] + uri
	}
	return false, ""
}

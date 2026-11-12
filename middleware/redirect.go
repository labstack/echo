// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// RedirectConfig defines the config for Redirect middleware.
type RedirectConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper

	// Status code to be used when redirecting the request.
	// Optional. Default value http.StatusMovedPermanently.
	Code int `yaml:"code"`
}

// redirectLogic represents a function that given a scheme, host and uri
// can both: 1) determine if redirect is needed (will set ok accordingly) and
// 2) return the appropriate redirect url.
type redirectLogic func(scheme, host, uri string) (ok bool, url string)

const www = "www."

// DefaultRedirectConfig is the default Redirect middleware config.
var DefaultRedirectConfig = RedirectConfig{
	Skipper: DefaultSkipper,
	Code:    http.StatusMovedPermanently,
}

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
	return redirect(config, func(scheme, host, uri string) (bool, string) {
		if scheme != "https" {
			return true, "https://" + host + uri
		}
		return false, ""
	})
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
	return redirect(config, func(scheme, host, uri string) (bool, string) {
		if scheme != "https" && !hasWWWPrefix(host) {
			return true, "https://www." + host + uri
		}
		return false, ""
	})
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
	return redirect(config, func(scheme, host, uri string) (ok bool, url string) {
		if scheme != "https" {
			host = stripWWWPrefix(host)
			return true, "https://" + host + uri
		}
		return false, ""
	})
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
	return redirect(config, func(scheme, host, uri string) (bool, string) {
		if !hasWWWPrefix(host) {
			return true, scheme + "://www." + host + uri
		}
		return false, ""
	})
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
	return redirect(config, func(scheme, host, uri string) (bool, string) {
		if hasWWWPrefix(host) {
			return true, scheme + "://" + host[len(www):] + uri
		}
		return false, ""
	})
}

func redirect(config RedirectConfig, cb redirectLogic) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRedirectConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req, scheme := c.Request(), c.Scheme()
			if req.URL.Path == "*" {
				return next(c)
			}
			switch scheme {
			case "ws":
				scheme = "http"
			case "wss":
				scheme = "https"
			}
			host := req.Host
			ok, target := cb(scheme, host, "")
			if !ok {
				return next(c)
			}
			if !isValidRedirectHost(host) {
				return echo.ErrBadRequest
			}
			uri := req.RequestURI
			if uri == "" || uri[0] != '/' {
				uri = req.URL.RequestURI()
			}
			// The host already anchors this absolute URL, so preserve leading path
			// slashes. Only control characters need escaping here.
			return c.Redirect(config.Code, target+escapeControlChars(uri))
		}
	}
}

func hasWWWPrefix(host string) bool {
	return len(host) >= len(www) && strings.EqualFold(host[:len(www)], www)
}

func stripWWWPrefix(host string) string {
	if hasWWWPrefix(host) {
		return host[len(www):]
	}
	return host
}

// isValidRedirectHost checks the authority before it is used in a Location.
// Check the stripped host too, as removing a bare www. would leave no hostname.
func isValidRedirectHost(host string) bool {
	host = stripWWWPrefix(host)
	if host == "" || host[0] == ':' || host[0] == '.' {
		return false
	}
	// net/http validates Host syntax. Reject percent escapes too, since a
	// browser could interpret them as authority delimiters in Location.
	return !strings.Contains(host, "%")
}

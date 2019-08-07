package middleware

import (
	"path"

	"github.com/labstack/echo/v4"
)

type (
	// TrailingSlashConfig defines the config for TrailingSlash middleware.
	TrailingSlashConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Status code to be used when redirecting the request.
		// Optional, but when provided the request is redirected using this code.
		RedirectCode int `yaml:"redirect_code"`
	}
)

var (
	// DefaultTrailingSlashConfig is the default TrailingSlash middleware config.
	DefaultTrailingSlashConfig = TrailingSlashConfig{
		Skipper: DefaultSkipper,
	}
)

// AddTrailingSlash returns a root level (before router) middleware which adds a
// trailing slash to the request `URL#Path`.
//
// Usage `Echo#Pre(AddTrailingSlash())`
func AddTrailingSlash() echo.MiddlewareFunc {
	return AddTrailingSlashWithConfig(DefaultTrailingSlashConfig)
}

// AddTrailingSlashWithConfig returns a AddTrailingSlash middleware with config.
// See `AddTrailingSlash()`.
func AddTrailingSlashWithConfig(config TrailingSlashConfig) echo.MiddlewareFunc {
	return slashNormalizerWithConfig(config, true)
}

// RemoveTrailingSlash returns a root level (before router) middleware which removes
// a trailing slash from the request URI.
//
// Usage `Echo#Pre(RemoveTrailingSlash())`
func RemoveTrailingSlash() echo.MiddlewareFunc {
	return RemoveTrailingSlashWithConfig(TrailingSlashConfig{})
}

// RemoveTrailingSlashWithConfig returns a RemoveTrailingSlash middleware with config.
// See `RemoveTrailingSlash()`.
func RemoveTrailingSlashWithConfig(config TrailingSlashConfig) echo.MiddlewareFunc {
	return slashNormalizerWithConfig(config, false)
}

// A "slash normalizing" middleware that will either return path/URIs with or
// without trailing slashes depending on the addSlash parameter.
func slashNormalizerWithConfig(config TrailingSlashConfig, addSlash bool) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			url := req.URL

			cleanpath, cleanURI := normalizePath(url.Path, c.QueryString(), addSlash)

			// Redirect
			if config.RedirectCode != 0 {
				return c.Redirect(config.RedirectCode, cleanURI)
			}

			// Forward
			req.RequestURI = cleanURI
			url.Path = cleanpath

			return next(c)
		}
	}
}

// Given a potentially dirty path and a querystring (from a request),
// return a "clean" version of the path, and a "clean" version of the whole URI.
// If the addSlash parameter is set to true, a single trailing slash will be appended.
func normalizePath(dirtyPath string, querystring string, addSlash bool) (string, string) {
	cleanPath := ""
	if len(dirtyPath) != 0 {
		cleanPath = path.Clean(dirtyPath)
	}

	if addSlash {
		cleanPath += "/"
	}

	cleanURI := cleanPath

	if querystring != "" {
		cleanURI += "?" + querystring
	}

	return cleanPath, cleanURI
}

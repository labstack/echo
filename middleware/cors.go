package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `json:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `json:"allow_methods"`

		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `json:"allow_headers"`

		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `json:"allow_credentials"`

		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `json:"expose_headers"`

		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `json:"max_age"`
	}
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		Skipper:      defaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See: https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func CORS() echo.MiddlewareFunc {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig returns a CORS middleware from config.
// See: `CORS()`.
func CORSWithConfig(config CORSConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCORSConfig.Skipper
	}
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}
	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			origin := req.Header().Get(echo.HeaderOrigin)
			originSet := req.Header().Contains(echo.HeaderOrigin) // Issue #517

			// Check allowed origins
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = o
					break
				}
			}

			// Simple request
			if req.Method() != echo.OPTIONS {
				res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
				if !originSet || allowedOrigin == "" {
					return next(c)
				}
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowedOrigin)
				if config.AllowCredentials {
					res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					res.Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(c)
			}

			// Preflight request
			res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
			if !originSet || allowedOrigin == "" {
				return next(c)
			}
			res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowedOrigin)
			res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := req.Header().Get(echo.HeaderAccessControlRequestHeaders)
				if h != "" {
					res.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
				}
			}
			if config.MaxAge > 0 {
				res.Header().Set(echo.HeaderAccessControlMaxAge, maxAge)
			}
			return c.NoContent(http.StatusNoContent)
		}
	}
}

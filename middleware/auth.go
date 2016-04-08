package middleware

import (
	"encoding/base64"

	"github.com/labstack/echo"
)

type (
	// BasicAuthConfig defines the config for HTTP basic auth middleware.
	BasicAuthConfig struct {
		// AuthFunc is the function to validate basic auth credentials.
		AuthFunc BasicAuthFunc
	}

	// BasicAuthFunc defines a function to validate basic auth credentials.
	BasicAuthFunc func(string, string) bool
)

const (
	basic = "Basic"
)

var (
	// DefaultBasicAuthConfig is the default basic auth middleware config.
	DefaultBasicAuthConfig = BasicAuthConfig{}
)

// BasicAuth returns an HTTP basic auth middleware.
//
// For valid credentials it calls the next handler.
// For invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(f BasicAuthFunc) echo.MiddlewareFunc {
	c := DefaultBasicAuthConfig
	c.AuthFunc = f
	return BasicAuthWithConfig(c)
}

// BasicAuthWithConfig returns an HTTP basic auth middleware from config.
// See `BasicAuth()`.
func BasicAuthWithConfig(config BasicAuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header().Get(echo.HeaderAuthorization)
			l := len(basic)

			if len(auth) > l+1 && auth[:l] == basic {
				b, err := base64.StdEncoding.DecodeString(auth[l+1:])
				if err == nil {
					cred := string(b)
					for i := 0; i < len(cred); i++ {
						if cred[i] == ':' {
							// Verify credentials
							if config.AuthFunc(cred[:i], cred[i+1:]) {
								return next(c)
							}
						}
					}
				}
			}
			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm=Restricted")
			return echo.ErrUnauthorized
		}
	}
}

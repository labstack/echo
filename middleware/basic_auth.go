package middleware

import (
	"encoding/base64"

	"github.com/labstack/echo"
)

type (
	// BasicAuthConfig defines the config for BasicAuth middleware.
	BasicAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Validator is a function to validate BasicAuth credentials.
		// Required.
		Validator BasicAuthValidator
	}

	// BasicAuthValidator defines a function to validate BasicAuth credentials.
	BasicAuthValidator func(string, string) bool
)

const (
	basic = "Basic"
)

var (
	// DefaultBasicAuthConfig is the default BasicAuth middleware config.
	DefaultBasicAuthConfig = BasicAuthConfig{
		Skipper: defaultSkipper,
	}
)

// BasicAuth returns an BasicAuth middleware.
//
// For valid credentials it calls the next handler.
// For invalid credentials, it sends "401 - Unauthorized" response.
// For empty or invalid `Authorization` header, it sends "400 - Bad Request" response.
func BasicAuth(fn BasicAuthValidator) echo.MiddlewareFunc {
	c := DefaultBasicAuthConfig
	c.Validator = fn
	return BasicAuthWithConfig(c)
}

// BasicAuthWithConfig returns an BasicAuth middleware with config.
// See `BasicAuth()`.
func BasicAuthWithConfig(config BasicAuthConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Validator == nil {
		panic("basic-auth middleware requires validator function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultBasicAuthConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			auth := c.Request().Header().Get(echo.HeaderAuthorization)
			l := len(basic)

			if len(auth) > l+1 && auth[:l] == basic {
				b, err := base64.StdEncoding.DecodeString(auth[l+1:])
				if err != nil {
					return err
				}
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						if config.Validator(cred[:i], cred[i+1:]) {
							return next(c)
						}
					}
				}
			}
			// Need to return `401` for browsers to pop-up login box.
			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm=Restricted")
			return echo.ErrUnauthorized
		}
	}
}

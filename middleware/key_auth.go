package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

type (
	// KeyAuthConfig defines the config for KeyAuth middleware.
	KeyAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// KeyLookup is a string in the form of "<source>:<name>" that is used
		// to extract key from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		KeyLookup string `json:"key_lookup"`

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		// Validator is a function to validate key.
		// Required.
		Validator KeyAuthValidator
	}

	// KeyAuthValidator defines a function to validate KeyAuth credentials.
	KeyAuthValidator func(string, echo.Context) (bool, error)

	keyExtractor func(echo.Context) (string, error)
)

var (
	// DefaultKeyAuthConfig is the default KeyAuth middleware config.
	DefaultKeyAuthConfig = KeyAuthConfig{
		Skipper:    DefaultSkipper,
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
	}
)

// KeyAuth returns an KeyAuth middleware.
//
// For valid key it calls the next handler.
// For invalid key, it sends "401 - Unauthorized" response.
// For missing key, it sends "400 - Bad Request" response.
func KeyAuth(fn KeyAuthValidator) echo.MiddlewareFunc {
	c := DefaultKeyAuthConfig
	c.Validator = fn
	return KeyAuthWithConfig(c)
}

// KeyAuthWithConfig returns an KeyAuth middleware with config.
// See `KeyAuth()`.
func KeyAuthWithConfig(config KeyAuthConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultKeyAuthConfig.Skipper
	}
	// Defaults
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultKeyAuthConfig.AuthScheme
	}
	if config.KeyLookup == "" {
		config.KeyLookup = DefaultKeyAuthConfig.KeyLookup
	}
	if config.Validator == nil {
		panic("key-auth middleware requires a validator function")
	}

	// Initialize
	parts := strings.Split(config.KeyLookup, ":")
	extractor := keyFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = keyFromQuery(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Extract and verify key
			key, err := extractor(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			valid, err := config.Validator(key, c)
			if err != nil {
				return err
			} else if valid {
				return next(c)
			}

			return echo.ErrUnauthorized
		}
	}
}

// keyFromHeader returns a `keyExtractor` that extracts key from the request header.
func keyFromHeader(header string, authScheme string) keyExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		if auth == "" {
			return "", errors.New("Missing key in request header")
		}
		if header == echo.HeaderAuthorization {
			l := len(authScheme)
			if len(auth) > l+1 && auth[:l] == authScheme {
				return auth[l+1:], nil
			}
			return "", errors.New("Invalid key in the request header")
		}
		return auth, nil
	}
}

// keyFromQuery returns a `keyExtractor` that extracts key from the query string.
func keyFromQuery(param string) keyExtractor {
	return func(c echo.Context) (string, error) {
		key := c.QueryParam(param)
		if key == "" {
			return "", errors.New("Missing key in the query string")
		}
		return key, nil
	}
}

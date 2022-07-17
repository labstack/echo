package middleware

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

// BasicAuthConfig defines the config for BasicAuthWithConfig middleware.
type BasicAuthConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Validator is a function to validate BasicAuthWithConfig credentials. Note: if request contains multiple basic auth headers
	// this function would be called once for each header until first valid result is returned
	// Required.
	Validator BasicAuthValidator

	// Realm is a string to define realm attribute of BasicAuthWithConfig.
	// Default value "Restricted".
	Realm string
}

// BasicAuthValidator defines a function to validate BasicAuthWithConfig credentials.
type BasicAuthValidator func(c echo.Context, user string, password string) (bool, error)

const (
	basic        = "basic"
	defaultRealm = "Restricted"
)

// BasicAuth returns an BasicAuth middleware.
//
// For valid credentials it calls the next handler.
// For missing or invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn BasicAuthValidator) echo.MiddlewareFunc {
	return BasicAuthWithConfig(BasicAuthConfig{Validator: fn})
}

// BasicAuthWithConfig returns an BasicAuthWithConfig middleware with config.
func BasicAuthWithConfig(config BasicAuthConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts BasicAuthConfig to middleware or returns an error for invalid configuration
func (config BasicAuthConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Validator == nil {
		return nil, errors.New("echo basic-auth middleware requires a validator function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var lastError error
			l := len(basic)
			for _, auth := range c.Request().Header[echo.HeaderAuthorization] {
				if !(len(auth) > l+1 && strings.EqualFold(auth[:l], basic)) {
					continue
				}

				// Invalid base64 shouldn't be treated as error
				// instead should be treated as invalid client input
				b, errDecode := base64.StdEncoding.DecodeString(auth[l+1:])
				if errDecode != nil {
					lastError = echo.NewHTTPError(http.StatusBadRequest).WithInternal(errDecode)
					continue
				}
				idx := bytes.IndexByte(b, ':')
				if idx >= 0 {
					valid, errValidate := config.Validator(c, string(b[:idx]), string(b[idx+1:]))
					if errValidate != nil {
						lastError = errValidate
					} else if valid {
						return next(c)
					}
				}
			}

			if lastError != nil {
				return lastError
			}

			realm := defaultRealm
			if config.Realm != defaultRealm {
				realm = strconv.Quote(config.Realm)
			}

			// Need to return `401` for browsers to pop-up login box.
			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm="+realm)
			return echo.ErrUnauthorized
		}
	}, nil
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// BasicAuthConfig defines the config for BasicAuth middleware.
type BasicAuthConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Validator is a function to validate BasicAuth credentials.
	// Required.
	Validator BasicAuthValidator

	// HeaderValidationLimit limits the amount of authorization headers will be validated
	// for valid credentials. Set this value to be higher from in an environment where multiple
	// basic auth headers could be received.
	// Default value 1.
	HeaderValidationLimit int

	// Realm is a string to define realm attribute of BasicAuth.
	// Default value "Restricted".
	Realm string
}

// BasicAuthValidator defines a function to validate BasicAuth credentials.
// The function should return a boolean indicating whether the credentials are valid,
// and an error if any error occurs during the validation process.
type BasicAuthValidator func(user string, password string, c echo.Context) (bool, error)

const (
	basic        = "basic"
	defaultRealm = "Restricted"
)

// DefaultBasicAuthConfig is the default BasicAuth middleware config.
var DefaultBasicAuthConfig = BasicAuthConfig{
	Skipper:               DefaultSkipper,
	Realm:                 defaultRealm,
	HeaderValidationLimit: 1,
}

// BasicAuth returns an BasicAuth middleware.
//
// For valid credentials it calls the next handler.
// For missing or invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn BasicAuthValidator) echo.MiddlewareFunc {
	c := DefaultBasicAuthConfig
	c.Validator = fn
	return BasicAuthWithConfig(c)
}

// BasicAuthWithConfig returns an BasicAuthWithConfig middleware with config.
func BasicAuthWithConfig(config BasicAuthConfig) echo.MiddlewareFunc {
	mw, err := config.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// ToMiddleware converts BasicAuthConfig to middleware or returns an error for invalid configuration
func (config BasicAuthConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Validator == nil {
		return nil, errors.New("echo basic-auth middleware requires a validator function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	realm := defaultRealm
	if config.Realm != "" && config.Realm != realm {
		realm = strconv.Quote(config.Realm)
	}
	maxValidationAttemptCount := 1
	if config.HeaderValidationLimit > 1 {
		maxValidationAttemptCount = config.HeaderValidationLimit
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var lastError error
			l := len(basic)
			errCount := 0
			// multiple auth headers is something that can happen in environments like
			// corporate test environments that are secured by application proxy servers where
			// front facing proxy is also configured to require own basic auth value and does auth checks.
			// In that case middleware can receive multiple auth headers.
			for _, auth := range c.Request().Header[echo.HeaderAuthorization] {
				if !(len(auth) > l+1 && strings.EqualFold(auth[:l], basic)) {
					continue
				}
				if errCount >= maxValidationAttemptCount {
					break
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
					valid, errValidate := config.Validator(string(b[:idx]), string(b[idx+1:]), c)
					if errValidate != nil {
						lastError = errValidate
					} else if valid {
						return next(c)
					}
					errCount++
				}
			}

			if lastError != nil {
				return lastError
			}

			// Need to return `401` for browsers to pop-up login box.
			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm="+realm)
			return echo.ErrUnauthorized
		}
	}, nil
}

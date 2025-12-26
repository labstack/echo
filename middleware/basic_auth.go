// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"cmp"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

// BasicAuthConfig defines the config for BasicAuthWithConfig middleware.
//
// SECURITY: The Validator function is responsible for securely comparing credentials.
// See BasicAuthValidator documentation for guidance on preventing timing attacks.
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

	// AllowedCheckLimit set how many headers are allowed to be checked. This is useful
	// environments like corporate test environments with application proxies restricting
	// access to environment with their own auth scheme.
	// Defaults to 1.
	AllowedCheckLimit uint
}

// BasicAuthValidator defines a function to validate BasicAuthWithConfig credentials.
//
// SECURITY WARNING: To prevent timing attacks that could allow attackers to enumerate
// valid usernames or passwords, validator implementations MUST use constant-time
// comparison for credential checking. Use crypto/subtle.ConstantTimeCompare instead
// of standard string equality (==) or switch statements.
//
// Example of SECURE implementation:
//
//	import "crypto/subtle"
//
//	validator := func(c *echo.Context, username, password string) (bool, error) {
//	    // Fetch expected credentials from database/config
//	    expectedUser := "admin"
//	    expectedPass := "secretpassword"
//
//	    // Use constant-time comparison to prevent timing attacks
//	    userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1
//	    passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1
//
//	    if userMatch && passMatch {
//	        return true, nil
//	    }
//	    return false, nil
//	}
//
// Example of INSECURE implementation (DO NOT USE):
//
//	// VULNERABLE TO TIMING ATTACKS - DO NOT USE
//	validator := func(c *echo.Context, username, password string) (bool, error) {
//	    if username == "admin" && password == "secret" {  // Timing leak!
//	        return true, nil
//	    }
//	    return false, nil
//	}
type BasicAuthValidator func(c *echo.Context, user string, password string) (bool, error)

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
	realm := defaultRealm
	if config.Realm != "" {
		realm = config.Realm
	}
	realm = strconv.Quote(realm)
	limit := cmp.Or(config.AllowedCheckLimit, 1)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var lastError error
			l := len(basic)
			i := uint(0)
			for _, auth := range c.Request().Header[echo.HeaderAuthorization] {
				if i >= limit {
					break
				}
				if !(len(auth) > l+1 && strings.EqualFold(auth[:l], basic)) {
					continue
				}
				i++

				// Invalid base64 shouldn't be treated as error
				// instead should be treated as invalid client input
				b, errDecode := base64.StdEncoding.DecodeString(auth[l+1:])
				if errDecode != nil {
					lastError = echo.ErrBadRequest.Wrap(errDecode)
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

			// Need to return `401` for browsers to pop-up login box.
			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm="+realm)
			return echo.ErrUnauthorized
		}
	}, nil
}

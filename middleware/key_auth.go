// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

// KeyAuthConfig defines the config for KeyAuth middleware.
//
// SECURITY: The Validator function is responsible for securely comparing API keys.
// See KeyAuthValidator documentation for guidance on preventing timing attacks.
type KeyAuthConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// KeyLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract key from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>" or "header:<name>:<cut-prefix>"
	// 			`<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
	//			value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
	//			want to cut is `<auth-scheme> ` note the space at the end.
	//			In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove is `Basic `.
	// - "query:<name>"
	// - "form:<name>"
	// - "cookie:<name>"
	// Multiple sources example:
	// - "header:Authorization,header:X-Api-Key"
	KeyLookup string

	// AllowedCheckLimit set how many KeyLookup values are allowed to be checked. This is
	// useful environments like corporate test environments with application proxies restricting
	// access to environment with their own auth scheme.
	AllowedCheckLimit uint

	// Validator is a function to validate key.
	// Required.
	Validator KeyAuthValidator

	// ErrorHandler defines a function which is executed when all lookups have been done and none of them passed Validator
	// function. ErrorHandler is executed with last missing (ErrExtractionValueMissing) or an invalid key.
	// It may be used to define a custom error.
	//
	// Note: when error handler swallows the error (returns nil) middleware continues handler chain execution towards handler.
	// This is useful in cases when portion of your site/api is publicly accessible and has extra features for authorized users
	// In that case you can use ErrorHandler to set default public auth value to request and continue with handler chain.
	ErrorHandler KeyAuthErrorHandler

	// ContinueOnIgnoredError allows the next middleware/handler to be called when ErrorHandler decides to
	// ignore the error (by returning `nil`).
	// This is useful when parts of your site/api allow public access and some authorized routes provide extra functionality.
	// In that case you can use ErrorHandler to set a default public key auth value in the request context
	// and continue. Some logic down the remaining execution chain needs to check that (public) key auth value then.
	ContinueOnIgnoredError bool
}

// KeyAuthValidator defines a function to validate KeyAuth credentials.
//
// SECURITY WARNING: To prevent timing attacks that could allow attackers to enumerate
// valid API keys, validator implementations MUST use constant-time comparison.
// Use crypto/subtle.ConstantTimeCompare instead of standard string equality (==)
// or switch statements.
//
// Example of SECURE implementation:
//
//	import "crypto/subtle"
//
//	validator := func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
//	    // Fetch valid keys from database/config
//	    validKeys := []string{"key1", "key2", "key3"}
//
//	    for _, validKey := range validKeys {
//	        // Use constant-time comparison to prevent timing attacks
//	        if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) == 1 {
//	            return true, nil
//	        }
//	    }
//	    return false, nil
//	}
//
// Example of INSECURE implementation (DO NOT USE):
//
//	// VULNERABLE TO TIMING ATTACKS - DO NOT USE
//	validator := func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
//	    switch key {  // Timing leak!
//	    case "valid-key":
//	        return true, nil
//	    default:
//	        return false, nil
//	    }
//	}
type KeyAuthValidator func(c *echo.Context, key string, source ExtractorSource) (bool, error)

// KeyAuthErrorHandler defines a function which is executed for an invalid key.
type KeyAuthErrorHandler func(c *echo.Context, err error) error

// ErrKeyMissing denotes an error raised when key value could not be extracted from request
var ErrKeyMissing = echo.NewHTTPError(http.StatusUnauthorized, "missing key")

// ErrInvalidKey denotes an error raised when key value is invalid by validator
var ErrInvalidKey = echo.NewHTTPError(http.StatusUnauthorized, "invalid key")

// DefaultKeyAuthConfig is the default KeyAuth middleware config.
var DefaultKeyAuthConfig = KeyAuthConfig{
	Skipper:   DefaultSkipper,
	KeyLookup: "header:" + echo.HeaderAuthorization + ":Bearer ",
}

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

// KeyAuthWithConfig returns an KeyAuth middleware or panics if configuration is invalid.
//
// For first valid key it calls the next handler.
// For invalid key, it sends "401 - Unauthorized" response.
// For missing key, it sends "400 - Bad Request" response.
func KeyAuthWithConfig(config KeyAuthConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts KeyAuthConfig to middleware or returns an error for invalid configuration
func (config KeyAuthConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultKeyAuthConfig.Skipper
	}
	if config.KeyLookup == "" {
		config.KeyLookup = DefaultKeyAuthConfig.KeyLookup
	}
	if config.Validator == nil {
		return nil, errors.New("echo key-auth middleware requires a validator function")
	}

	limit := cmp.Or(config.AllowedCheckLimit, 1)

	extractors, cErr := createExtractors(config.KeyLookup, limit)
	if cErr != nil {
		return nil, fmt.Errorf("echo key-auth middleware could not create key extractor: %w", cErr)
	}
	if len(extractors) == 0 {
		return nil, errors.New("echo key-auth middleware could not create extractors from KeyLookup string")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var lastExtractorErr error
			var lastValidatorErr error
			for _, extractor := range extractors {
				keys, source, extrErr := extractor(c)
				if extrErr != nil {
					lastExtractorErr = extrErr
					continue
				}
				for _, key := range keys {
					valid, err := config.Validator(c, key, source)
					if err != nil {
						lastValidatorErr = err
						continue
					}
					if !valid {
						lastValidatorErr = ErrInvalidKey
						continue
					}
					return next(c)
				}
			}

			// prioritize validator errors over extracting errors
			err := lastValidatorErr
			if err == nil {
				err = lastExtractorErr
			}
			if config.ErrorHandler != nil {
				tmpErr := config.ErrorHandler(c, err)
				if config.ContinueOnIgnoredError && tmpErr == nil {
					return next(c)
				}
				return tmpErr
			}
			if lastValidatorErr == nil {
				return ErrKeyMissing.Wrap(err)
			}
			return echo.ErrUnauthorized.Wrap(err)
		}
	}, nil
}

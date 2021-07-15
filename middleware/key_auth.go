package middleware

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// KeyAuthConfig defines the config for KeyAuth middleware.
type KeyAuthConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// KeyLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract key(s) from the request.
	// Optional. Default value "header:Authorization:Bearer ".
	// Possible values:
	// - "header:<name>:<value prefix>"
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	// - "form:<name>"
	// Multiple sources example:
	// - "header:Authorization:Bearer ,cookie:myowncookie"
	KeyLookup string

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
}

// KeyAuthValidator defines a function to validate KeyAuth credentials.
type KeyAuthValidator func(c echo.Context, key string, keyType ExtractorType) (bool, error)

// KeyAuthErrorHandler defines a function which is executed for an invalid key.
type KeyAuthErrorHandler func(c echo.Context, err error) error

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
	extractors, err := createExtractors(config.KeyLookup)
	if err != nil {
		return nil, fmt.Errorf("echo key-auth middleware could not create key extractor: %w", err)
	}
	if len(extractors) == 0 {
		return nil, errors.New("echo key-auth middleware could not create extractors from KeyLookup string")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var lastExtractorErr error
			var lastValidatorErr error
			for _, extractor := range extractors {
				keys, keyType, extrErr := extractor(c)
				if extrErr != nil {
					lastExtractorErr = extrErr
					continue
				}
				for _, key := range keys {
					valid, err := config.Validator(c, key, keyType)
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
				// Allow error handler to swallow the error and continue handler chain execution
				// Useful in cases when portion of your site/api is publicly accessible and has extra features for authorized users
				// In that case you can use ErrorHandler to set default public auth value to request and continue with handler chain
				if handledErr := config.ErrorHandler(c, err); handledErr != nil {
					return handledErr
				}
				return next(c)
			}
			if err == ErrExtractionValueMissing {
				return ErrKeyMissing // do not wrap extractor errors
			}
			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  "Unauthorized",
				Internal: err,
			}
		}
	}, nil
}

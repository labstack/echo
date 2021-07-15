package middleware

import (
	"errors"
	"github.com/labstack/echo/v5"
	"net/http"
)

// JWTConfig defines the config for JWT middleware.
type JWTConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// BeforeFunc defines a function which is executed just before the middleware.
	BeforeFunc BeforeFunc

	// SuccessHandler defines a function which is executed for a valid token.
	SuccessHandler JWTSuccessHandler

	// ErrorHandler defines a function which is executed when all lookups have been done and none of them passed Validator
	// function. ErrorHandler is executed with last missing (ErrExtractionValueMissing) or an invalid key.
	// It may be used to define a custom JWT error.
	//
	// Note: when error handler swallows the error (returns nil) middleware continues handler chain execution towards handler.
	// This is useful in cases when portion of your site/api is publicly accessible and has extra features for authorized users
	// In that case you can use ErrorHandler to set default public JWT token value to request and continue with handler chain.
	ErrorHandler JWTErrorHandlerWithContext

	// ContinueOnIgnoredError allows the next middleware/handler to be called when ErrorHandlerWithContext decides to
	// ignore the error (by returning `nil`).
	// This is useful when parts of your site/api allow public access and some authorized routes provide extra functionality.
	// In that case you can use ErrorHandlerWithContext to set a default public JWT token value in the request context
	// and continue. Some logic down the remaining execution chain needs to check that (public) token value then.
	ContinueOnIgnoredError bool

	// Context key to store user information from the token into context.
	// Optional. Default value "user".
	ContextKey string

	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>" or "header:<name>:<cut-prefix>"
	// 			`<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
	//			value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
	//			want to cut is `<auth-scheme> ` note the space at the end.
	//			In case of JWT tokens `Authorization: Bearer <token>` prefix we cut is `Bearer `.
	// If prefix is left empty the whole value is returned.
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	// - "form:<name>"
	// Multiple sources example:
	// - "header:Authorization:Bearer ,cookie:myowncookie"
	TokenLookup string

	// TokenLookupFuncs defines a list of user-defined functions that extract JWT token from the given context.
	// This is one of the two options to provide a token extractor.
	// The order of precedence is user-defined TokenLookupFuncs, and TokenLookup.
	// You can also provide both if you want.
	TokenLookupFuncs []ValuesExtractor

	// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
	// parsing fails or parsed token is invalid.
	// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
	ParseTokenFunc func(c echo.Context, auth string) (interface{}, error)
}

// JWTSuccessHandler defines a function which is executed for a valid token.
type JWTSuccessHandler func(c echo.Context)

// JWTErrorHandler defines a function which is executed for an invalid token.
type JWTErrorHandler func(err error) error

// JWTErrorHandlerWithContext is almost identical to JWTErrorHandler, but it's passed the current context.
type JWTErrorHandlerWithContext func(c echo.Context, err error) error

const (
	// AlgorithmHS256 is token signing algorithm
	AlgorithmHS256 = "HS256"
)

// ErrJWTMissing denotes an error raised when JWT token value could not be extracted from request
var ErrJWTMissing = echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed jwt")

// ErrJWTInvalid denotes an error raised when JWT token value is invalid or expired
var ErrJWTInvalid = echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt")

// DefaultJWTConfig is the default JWT auth middleware config.
var DefaultJWTConfig = JWTConfig{
	Skipper:     DefaultSkipper,
	ContextKey:  "user",
	TokenLookup: "header:" + echo.HeaderAuthorization + ":Bearer ",
}

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For missing token, it returns "400 - Bad Request" error.
//
// See: https://jwt.io/introduction
func JWT(parseTokenFunc func(c echo.Context, auth string) (interface{}, error)) echo.MiddlewareFunc {
	c := DefaultJWTConfig
	c.ParseTokenFunc = parseTokenFunc
	return JWTWithConfig(c)
}

// JWTWithConfig returns a JSON Web Token (JWT) auth middleware or panics if configuration is invalid.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For missing token, it returns "400 - Bad Request" error.
//
// See: https://jwt.io/introduction
func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts JWTConfig to middleware or returns an error for invalid configuration
func (config JWTConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.ParseTokenFunc == nil {
		return nil, errors.New("echo jwt middleware requires parse token function")
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.TokenLookup == "" && len(config.TokenLookupFuncs) == 0 {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	extractors, err := createExtractors(config.TokenLookup)
	if err != nil {
		return nil, err
	}
	if len(config.TokenLookupFuncs) > 0 {
		extractors = append(config.TokenLookupFuncs, extractors...)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}
			var lastExtractorErr error
			var lastTokenErr error
			for _, extractor := range extractors {
				auths, extrErr := extractor(c)
				if extrErr != nil {
					lastExtractorErr = extrErr
					continue
				}
				for _, auth := range auths {
					token, err := config.ParseTokenFunc(c, auth)
					if err != nil {
						lastTokenErr = err
						continue
					}
					// Store user information from token into context.
					c.Set(config.ContextKey, token)
					if config.SuccessHandler != nil {
						config.SuccessHandler(c)
					}
					return next(c)
				}
			}

			// prioritize token errors over extracting errors
			err := lastTokenErr
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
			if lastTokenErr == nil {
				return ErrJWTMissing.WithInternal(err)
			}
			return ErrJWTInvalid.WithInternal(err)
		}
	}, nil
}

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type (
	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Signing key to validate token.
		// Required.
		SigningKey interface{} `json:"signing_key"`

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string `json:"signing_method"`

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string `json:"context_key"`

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		TokenLookup string `json:"token_lookup"`
	}

	jwtExtractor func(echo.Context) (string, error)
)

const (
	bearer = "Bearer"
)

// Algorithims
const (
	AlgorithmHS256 = "HS256"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		Skipper:       defaultSkipper,
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup:   "header:" + echo.HeaderAuthorization,
		Claims:        jwt.MapClaims{},
	}
)

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it sends "401 - Unauthorized" response.
// For empty or invalid `Authorization` header, it sends "400 - Bad Request".
//
// See: https://jwt.io/introduction
func JWT(key []byte) echo.MiddlewareFunc {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

// JWTWithConfig returns a JWT auth middleware from config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.SigningKey == nil {
		panic("jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1])
	switch parts[0] {
	case "query":
		extractor = jwtFromQuery(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			auth, err := extractor(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			token, err := jwt.ParseWithClaims(auth, config.Claims, func(t *jwt.Token) (interface{}, error) {
				// Check the signing method
				if t.Method.Alg() != config.SigningMethod {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return config.SigningKey, nil

			})
			if err == nil && token.Valid {
				// Store user information from token into context.
				c.Set(config.ContextKey, token)
				return next(c)
			}
			return echo.ErrUnauthorized
		}
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the provided
// request header.
func jwtFromHeader(header string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header().Get(header)
		l := len(bearer)
		if len(auth) > l+1 && auth[:l] == bearer {
			return auth[l+1:], nil
		}
		return "", errors.New("empty or invalid jwt in authorization header")
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the provided query
// parameter.
func jwtFromQuery(param string) jwtExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", errors.New("empty jwt in query param")
		}
		return token, nil
	}
}

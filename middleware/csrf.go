package middleware

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// CSRFConfig defines the config for CSRF middleware.
type CSRFConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// TokenLength is the length of the generated token.
	TokenLength uint8
	// Optional. Default value 32.

	// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:X-CSRF-Token".
	// Possible values:
	// - "header:<name>" or "header:<name>:<cut-prefix>"
	// - "query:<name>"
	// - "form:<name>"
	// Multiple sources example:
	// - "header:X-CSRF-Token,query:csrf"
	TokenLookup string `yaml:"token_lookup"`

	// Generator defines a function to generate token.
	// Optional. Defaults tp randomString(TokenLength).
	Generator func() string

	// Context key to store generated CSRF token into context.
	// Optional. Default value "csrf".
	ContextKey string

	// Name of the CSRF cookie. This cookie will store CSRF token.
	// Optional. Default value "csrf".
	CookieName string

	// Domain of the CSRF cookie.
	// Optional. Default value none.
	CookieDomain string

	// Path of the CSRF cookie.
	// Optional. Default value none.
	CookiePath string

	// Max age (in seconds) of the CSRF cookie.
	// Optional. Default value 86400 (24hr).
	CookieMaxAge int

	// Indicates if CSRF cookie is secure.
	// Optional. Default value false.
	CookieSecure bool

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieHTTPOnly bool

	// Indicates SameSite mode of the CSRF cookie.
	// Optional. Default value SameSiteDefaultMode.
	CookieSameSite http.SameSite

	// ErrorHandler defines a function which is executed for returning custom errors.
	ErrorHandler func(c echo.Context, err error) error
}

// ErrCSRFInvalid is returned when CSRF check fails
var ErrCSRFInvalid = echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")

// DefaultCSRFConfig is the default CSRF middleware config.
var DefaultCSRFConfig = CSRFConfig{
	Skipper:        DefaultSkipper,
	TokenLength:    32,
	TokenLookup:    "header:" + echo.HeaderXCSRFToken,
	ContextKey:     "csrf",
	CookieName:     "_csrf",
	CookieMaxAge:   86400,
	CookieSameSite: http.SameSiteDefaultMode,
}

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
// See: https://en.wikipedia.org/wiki/Cross-site_request_forgery
func CSRF() echo.MiddlewareFunc {
	return CSRFWithConfig(DefaultCSRFConfig)
}

// CSRFWithConfig returns a CSRF middleware with config or panics on invalid configuration.
func CSRFWithConfig(config CSRFConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts CSRFConfig to middleware or returns an error for invalid configuration
func (config CSRFConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCSRFConfig.Skipper
	}
	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFConfig.TokenLength
	}
	if config.Generator == nil {
		config.Generator = createRandomStringGenerator(config.TokenLength)
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}
	if config.CookieSameSite == http.SameSiteNoneMode {
		config.CookieSecure = true
	}

	extractors, cErr := createExtractors(config.TokenLookup)
	if cErr != nil {
		return nil, cErr
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			token := ""
			if k, err := c.Cookie(config.CookieName); err != nil {
				token = config.Generator() // Generate token
			} else {
				token = k.Value // Reuse token
			}

			switch c.Request().Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			default:
				// Validate token only for requests which are not defined as 'safe' by RFC7231
				var lastExtractorErr error
				var lastTokenErr error
			outer:
				for _, extractor := range extractors {
					clientTokens, _, err := extractor(c)
					if err != nil {
						lastExtractorErr = err
						continue
					}

					for _, clientToken := range clientTokens {
						if validateCSRFToken(token, clientToken) {
							lastTokenErr = nil
							lastExtractorErr = nil
							break outer
						}
						lastTokenErr = ErrCSRFInvalid
					}
				}
				var finalErr error
				if lastTokenErr != nil {
					finalErr = lastTokenErr
				} else if lastExtractorErr != nil {
					finalErr = echo.ErrBadRequest.WithInternal(lastExtractorErr)
				}
				if finalErr != nil {
					if config.ErrorHandler != nil {
						return config.ErrorHandler(c, finalErr)
					}
					return finalErr
				}
			}

			// Set CSRF cookie
			cookie := new(http.Cookie)
			cookie.Name = config.CookieName
			cookie.Value = token
			if config.CookiePath != "" {
				cookie.Path = config.CookiePath
			}
			if config.CookieDomain != "" {
				cookie.Domain = config.CookieDomain
			}
			if config.CookieSameSite != http.SameSiteDefaultMode {
				cookie.SameSite = config.CookieSameSite
			}
			cookie.Expires = time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second)
			cookie.Secure = config.CookieSecure
			cookie.HttpOnly = config.CookieHTTPOnly
			c.SetCookie(cookie)

			// Store token in the context
			c.Set(config.ContextKey, token)

			// Protect clients from caching the response
			c.Response().Header().Add(echo.HeaderVary, echo.HeaderCookie)

			return next(c)
		}
	}, nil
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}

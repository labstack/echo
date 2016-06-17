package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type (
	// CSRFConfig defines the config for CSRF middleware.
	CSRFConfig struct {
		// Key to create CSRF token.
		Secret []byte `json:"secret"`

		// TokenLookup is a string in the form of "<source>:<key>" that is used
		// to extract token from the request.
		// Optional. Default value "header:X-CSRF-Token".
		// Multiple extractors can be combined with `;` e.g. "form:csrf;header:X-CSRF-Token"
		// and they're processed in order they're declared
		// Possible values:
		// - "header:<name>"
		// - "form:<name>"
		// - "query:<name>"
		TokenLookup string `json:"token_lookup"`

		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string `json:"context_key"`

		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrf".
		CookieName string `json:"cookie_name"`

		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string `json:"cookie_domain"`

		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string `json:"cookie_path"`

		// Expiration time of the CSRF cookie.
		// Optional. Default value 24H.
		CookieExpires time.Time `json:"cookie_expires"`

		// Indicates if CSRF cookie is secure.
		CookieSecure bool `json:"cookie_secure"`
		// Optional. Default value false.

		// Indicates if CSRF cookie is HTTP only.
		// Optional. Default value false.
		CookieHTTPOnly bool `json:"cookie_http_only"`
	}

	// csrfTokenExtractor defines a function that takes `echo.Context` and returns
	// either a token or an error.
	csrfTokenExtractor func(echo.Context) string
)

var (
	// DefaultCSRFConfig is the default CSRF middleware config.
	DefaultCSRFConfig = CSRFConfig{
		TokenLookup:   "header:" + echo.HeaderXCSRFToken,
		ContextKey:    "csrf",
		CookieName:    "csrf",
		CookieExpires: time.Now().Add(24 * time.Hour),
	}

	CSRFTokenNotFoundErr = errors.New("csrf token not found")
	CSRFEmptyHTTPErr     = echo.NewHTTPError(http.StatusForbidden, "csrf token not found")
	CSRFInvalidHTTPErr   = echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")
)

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
// See: https://en.wikipedia.org/wiki/Cross-site_request_forgery
func CSRF(secret []byte) echo.MiddlewareFunc {
	c := DefaultCSRFConfig
	c.Secret = secret
	return CSRFWithConfig(c)
}

// CSRFWithConfig returns a CSRF middleware from config.
// See `CSRF()`.
func CSRFWithConfig(config CSRFConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Secret == nil {
		panic("csrf secret must be provided")
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
	if config.CookieExpires.IsZero() {
		config.CookieExpires = DefaultCSRFConfig.CookieExpires
	}

	// Initialize
	extractors := make([]csrfTokenExtractor, 0)

	for _, extractor := range strings.Split(config.TokenLookup, ";") {
		parts := strings.Split(extractor, ":")

		switch parts[0] {
		case "form":
			extractors = append(extractors, csrfTokenFromForm(parts[1]))
		case "query":
			extractors = append(extractors, csrfTokenFromQuery(parts[1]))
		case "header":
			extractors = append(extractors, csrfTokenFromHeader(parts[1]))
		default:
			panic("invalid extractor name '" + parts[0] + "'")
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Set CSRF token
			salt, err := generateSalt(8)
			if err != nil {
				return err
			}
			token := generateCSRFToken(config.Secret, salt)
			c.Set(config.ContextKey, token)
			cookie := new(echo.Cookie)
			cookie.SetName(config.CookieName)
			cookie.SetValue(token)
			if config.CookiePath != "" {
				cookie.SetPath(config.CookiePath)
			}
			if config.CookieDomain != "" {
				cookie.SetDomain(config.CookieDomain)
			}
			cookie.SetExpires(config.CookieExpires)
			cookie.SetSecure(config.CookieSecure)
			cookie.SetHTTPOnly(config.CookieHTTPOnly)
			c.SetCookie(cookie)

			switch req.Method() {
			case echo.GET, echo.HEAD, echo.OPTIONS, echo.TRACE:
			default:
				token, err := extractCSRFToken(extractors, c)
				if err != nil {
					return CSRFEmptyHTTPErr
				}
				ok, err := validateCSRFToken(token, config.Secret)
				if err != nil {
					return err
				}
				if !ok {
					return CSRFInvalidHTTPErr
				}
			}
			return next(c)
		}
	}
}

// extractCSRFToken returns token if any of the provided extractors succeeds or
// the `CSRFTokenNotFoundErr` error
func extractCSRFToken(extractors []csrfTokenExtractor, c echo.Context) (string, error) {
	var token string

	for _, extractor := range extractors {
		token = extractor(c)

		if token != "" {
			return token, nil
		}
	}

	return "", CSRFTokenNotFoundErr
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(c echo.Context) string {
		return c.Request().Header().Get(header)
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(c echo.Context) string {
		return c.FormValue(param)
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(c echo.Context) string {
		return c.QueryParam(param)
	}
}

func generateCSRFToken(secret, salt []byte) string {
	h := hmac.New(sha1.New, secret)
	h.Write(salt)
	return fmt.Sprintf("%s:%s", hex.EncodeToString(h.Sum(nil)), hex.EncodeToString(salt))
}

func validateCSRFToken(token string, secret []byte) (bool, error) {
	sep := strings.Index(token, ":")
	if sep < 0 {
		return false, nil
	}
	salt, err := hex.DecodeString(token[sep+1:])
	if err != nil {
		return false, err
	}
	return token == generateCSRFToken(secret, salt), nil
}

func generateSalt(len uint8) (salt []byte, err error) {
	salt = make([]byte, len)
	_, err = rand.Read(salt)
	return
}

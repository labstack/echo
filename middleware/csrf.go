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
		Secret []byte

		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string

		// Extractor is a function that extracts token from the request.
		// Optional. Default value CSRFTokenFromHeader(echo.HeaderXCSRFToken).
		Extractor CSRFTokenExtractor

		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrf".
		CookieName string

		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string

		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string

		// Expiration time of the CSRF cookie.
		// Optional. Default value 24H.
		CookieExpires time.Time

		// Indicates if CSRF cookie is secure.
		CookieSecure bool
		// Optional. Default value false.

		// Indicates if CSRF cookie is HTTP only.
		// Optional. Default value false.
		CookieHTTPOnly bool
	}

	// CSRFTokenExtractor defines a function that takes `echo.Context` and returns
	// either a token or an error.
	CSRFTokenExtractor func(echo.Context) (string, error)
)

var (
	// DefaultCSRFConfig is the default CSRF middleware config.
	DefaultCSRFConfig = CSRFConfig{
		ContextKey:    "csrf",
		Extractor:     CSRFTokenFromHeader(echo.HeaderXCSRFToken),
		CookieName:    "csrf",
		CookieExpires: time.Now().Add(24 * time.Hour),
	}
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
	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if config.Extractor == nil {
		config.Extractor = DefaultCSRFConfig.Extractor
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}
	if config.CookieExpires.IsZero() {
		config.CookieExpires = DefaultCSRFConfig.CookieExpires
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
				token, err := config.Extractor(c)
				if err != nil {
					return err
				}
				ok, err := validateCSRFToken(token, config.Secret)
				if err != nil {
					return err
				}
				if !ok {
					return echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")
				}
			}
			return next(c)
		}
	}
}

// CSRFTokenFromHeader returns a `CSRFTokenExtractor` that extracts token from the
// provided request header.
func CSRFTokenFromHeader(header string) CSRFTokenExtractor {
	return func(c echo.Context) (string, error) {
		return c.Request().Header().Get(header), nil
	}
}

// CSRFTokenFromForm returns a `CSRFTokenExtractor` that extracts token from the
// provided form parameter.
func CSRFTokenFromForm(param string) CSRFTokenExtractor {
	return func(c echo.Context) (string, error) {
		token := c.FormValue(param)
		if token == "" {
			return "", errors.New("empty csrf token in form param")
		}
		return token, nil
	}
}

// CSRFTokenFromQuery returns a `CSRFTokenExtractor` that extracts token from the
// provided query parameter.
func CSRFTokenFromQuery(param string) CSRFTokenExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", errors.New("empty csrf token in query param")
		}
		return token, nil
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

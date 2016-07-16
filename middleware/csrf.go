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
		// Possible values:
		// - "header:<name>"
		// - "form:<name>"
		// - "query:<name>"
		TokenLookup string `json:"token_lookup"`
		
		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string `json:"context_key"`
		
		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrfToken".
		CookieName string `json:"cookie_name"`
		
		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string `json:"cookie_domain"`
		
		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string `json:"cookie_path"`
		
		// Max age (in seconds) of the CSRF cookie.
		// Optional. Default value 86400 (24hr).
		CookieMaxAge int `json:"cookie_max_age"`
		
		// Indicates if CSRF cookie is secure.
		// Optional. Default value false.
		CookieSecure bool `json:"cookie_secure"`
		
		// Indicates if CSRF cookie is HttpOnly.
		// An HttpOnly cookie cannot be accessed by JavaScript or other client-side APIs.
		CookieHttpOnly bool `json:"cookie_httponly"`
		
		// The message which is sent to the client when the token is not valid.
		// Only set the message for debug purposes.
		// Optional. Default value <empty string>.
		ForbiddenMessage string `json:"forbidden_message"`
	}
	
	// csrfTokenExtractor defines a function that takes `echo.Context` and returns
	// either a token or an error.
	csrfTokenExtractor func(echo.Context) (string, error)
)

var (
	// DefaultCSRFConfig is the default CSRF middleware config.
	DefaultCSRFConfig = CSRFConfig{
		TokenLookup:  "header:" + echo.HeaderXCSRFToken,
		ContextKey:   "csrf",
		CookieName:   "csrfToken",
		CookieMaxAge: 86400,
		CookieHttpOnly: true,
	}
	
	// DebugModeCSRFConfig is a CSRF middleware config for debugging.
	DebugModeCSRFConfig = CSRFConfig{
		TokenLookup:  "header:" + echo.HeaderXCSRFToken,
		ContextKey:   "csrf",
		CookieName:   "csrfToken",
		CookieMaxAge: 86400,
		CookieHttpOnly: true,
		ForbiddenMessage: "invalid csrf token", // It is a security issue to send explicit error messages to the client.
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
	// Set defaults
	sanitizeCsrfConfig(&config)
	
	// Initialize
	extractor := getExtractor(&config)
	
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			cookie, err := c.Cookie(config.CookieName)
			var token string
			if err == nil {
				// Get the token to use it again
				token = cookie.Value()
			} else {
				// Generate a new token
				token, err = newToken(&config)
				if err != nil {
					return err
				}
			}
			
			// Validate token only for requests which are not defined as 'safe' by RFC7231
			if req.Method() != echo.GET && req.Method() != echo.HEAD && req.Method() != echo.OPTIONS && req.Method() != echo.TRACE {
				clientToken, err := extractor(c)
				if err != nil {
					return err
				}
				ok, err := validateCSRFToken(token, clientToken, config.Secret)
				if err != nil {
					return err
				}
				if !ok {
					return echo.NewHTTPError(http.StatusForbidden, config.ForbiddenMessage)
				}
				// Regenerate the token after using it
				token, err = newToken(&config)
				if err != nil {
					return err
				}
			}
			
			// Set the CSRF cookie even if it's already set to renew the expiration time.
			c.SetCookie(setNewCookie(&config, token))
			
			c.Set(config.ContextKey, token)
			
			// Protects clients from caching the response
			c.Response().Header().Add("Vary", "Cookie")
			
			return next(c)
		}
	}
}

// Sets the default values for missed configurations.
func sanitizeCsrfConfig(config *CSRFConfig) {
	if config.Secret == nil {
		panic("csrf secret must be provided")
	}
	if len(config.TokenLookup) == 0 {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}
	if len(config.ContextKey) == 0 {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if len(config.CookieName) == 0 {
		config.CookieName = DefaultCSRFConfig.CookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}
}

// Returns csrfTokenExtractor of config
func getExtractor(config *CSRFConfig) csrfTokenExtractor {
	parts := strings.Split(config.TokenLookup, ":")
	switch parts[0] {
	case "form":
		return csrfTokenFromForm(parts[1])
	case "query":
		return csrfTokenFromQuery(parts[1])
	}
	return csrfTokenFromHeader(parts[1])
}

// Sets a new cookie for the token
func setNewCookie(config *CSRFConfig, token string) *echo.Cookie {
	cookie := new(echo.Cookie)
	cookie.SetName(config.CookieName)
	cookie.SetValue(token)
	if len(config.CookiePath) > 0 {
		cookie.SetPath(config.CookiePath)
	}
	if len(config.CookieDomain) > 0 {
		cookie.SetDomain(config.CookieDomain)
	}
	cookie.SetExpires(time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second))
	cookie.SetSecure(config.CookieSecure)
	cookie.SetHTTPOnly(config.CookieHttpOnly)
	return cookie
}

// Generates a new token
func newToken(config *CSRFConfig) (string, error) {
	salt, err := generateSalt(8)
	if err != nil {
		return "", err
	}
	return generateCSRFToken(config.Secret, salt), err
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(c echo.Context) (string, error) {
		return c.Request().Header().Get(header), nil
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(c echo.Context) (string, error) {
		token := c.FormValue(param)
		if token == "" {
			return "", errors.New("empty csrf token in form param")
		}
		return token, nil
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfTokenExtractor {
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

func validateCSRFToken(serverToken, clientToken string, secret []byte) (bool, error) {
	if serverToken != clientToken {
		return false, nil
	}
	sep := strings.Index(clientToken, ":")
	if sep < 0 {
		return false, nil
	}
	salt, err := hex.DecodeString(clientToken[sep+1:])
	if err != nil {
		return false, err
	}
	return clientToken == generateCSRFToken(secret, salt), nil
}

func generateSalt(len uint8) (salt []byte, err error) {
	salt = make([]byte, len)
	_, err = rand.Read(salt)
	return
}

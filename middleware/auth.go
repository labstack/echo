package middleware

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo"
)

type (
	// BasicValidateFunc is the expected format a BasicAuth fn argument is
	// expected to implement.
	BasicValidateFunc func(string, string) bool
)

const (
	// Basic is the authentication scheme implemented by the middleware.
	Basic = "Basic"
)

// BasicAuth returns a HTTP basic authentication middleware.
// For valid credentials, it calls the next handler.
// For invalid credentials, it returns a "401 Unauthorized" HTTP error.
func BasicAuth(fn BasicValidateFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip WebSocket
		if (c.Request().Header.Get(echo.Upgrade)) == echo.WebSocket {
			return nil
		}

		auth := c.Request().Header.Get(echo.Authorization)
		l := len(Basic)

		if len(auth) > l+1 && auth[:l] == Basic {
			b, err := base64.StdEncoding.DecodeString(auth[l+1:])
			if err == nil {
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						if fn(cred[:i], cred[i+1:]) {
							return nil
						}
					}
				}
			}
		}
		c.Response().Header().Set(echo.WWWAuthenticate, Basic+" realm=Restricted")
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
}

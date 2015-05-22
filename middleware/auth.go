package middleware

import (
	"encoding/base64"
	"github.com/labstack/echo"
	"net/http"
)

type (
	AuthFunc func(string, string) bool
)

const (
	Basic = "Basic"
)

// BasicAuth returns an HTTP basic authentication middleware. For valid credentials
// it calls the next handler in the chain.

// For invalid Authorization header it sends "404 - Bad Request" response.
// For invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn AuthFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		auth := c.Request().Header.Get(echo.Authorization)
		i := 0
		code := http.StatusBadRequest

		for ; i < len(auth); i++ {
			c := auth[i]
			// Ignore empty spaces
			if c == ' ' {
				continue
			}

			// Check scheme
			if i < len(Basic) {
				// Ignore case
				if i == 0 {
					if c != Basic[i] && c != 'b' {
						break
					}
				} else {
					if c != Basic[i] {
						break
					}
				}
			} else {
				// Extract credentials
				b, err := base64.StdEncoding.DecodeString(auth[i:])
				if err != nil {
					break
				}
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						if fn(cred[:i], cred[i+1:]) {
							return nil
						}
						code = http.StatusUnauthorized
						break
					}
				}
			}
		}
		return echo.NewHTTPError(code)
	}
}

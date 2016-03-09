package middleware

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo"
)

type (
	BasicAuthOptions struct {
	}

	BasicAuthFunc func(string, string) bool
)

const (
	basic = "Basic"
)

// BasicAuth returns an HTTP basic authentication middleware.
//
// For valid credentials it calls the next handler.
// For invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn BasicAuthFunc, options ...*BasicAuthOptions) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			auth := c.Request().Header().Get(echo.Authorization)
			l := len(basic)

			if len(auth) > l+1 && auth[:l] == basic {
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
			c.Response().Header().Set(echo.WWWAuthenticate, basic+" realm=Restricted")
			return echo.NewHTTPError(http.StatusUnauthorized)
		})
	}
}

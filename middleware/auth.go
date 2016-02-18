package middleware

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo"
)

type (
	// BasicAuth defines an HTTP basic authentication middleware.
	BasicAuth struct {
		function BasicAuthFunc
		priority int
	}

	BasicAuthFunc func(string, string) bool
)

const (
	basic = "Basic"
)

func NewBasicAuth(fn BasicAuthFunc) *BasicAuth {
	return &BasicAuth{function: fn}
}

func (ba *BasicAuth) SetPriority(p int) {
	ba.priority = p
}

func (ba *BasicAuth) Priority() int {
	return ba.priority
}

// Handle validates credentials using `AuthFunc`
//
// For valid credentials it calls the next handler.
// For invalid credentials, it sends "401 - Unauthorized" response.
func (ba *BasicAuth) Handle(h echo.Handler) echo.Handler {
	return echo.HandlerFunc(func(c echo.Context) error {
		// Skip WebSocket
		if (c.Request().Header().Get(echo.Upgrade)) == echo.WebSocket {
			return nil
		}

		auth := c.Request().Header().Get(echo.Authorization)
		l := len(basic)

		if len(auth) > l+1 && auth[:l] == basic {
			b, err := base64.StdEncoding.DecodeString(auth[l+1:])
			if err == nil {
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						if ba.function(cred[:i], cred[i+1:]) {
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

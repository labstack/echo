package middleware

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
)

type (
	BasicValidateFunc func(string, string) bool
	JWTValidateFunc   func(string, jwt.SigningMethod) ([]byte, error)
)

const (
	Basic  = "Basic"
	Bearer = "Bearer"
)

// BasicAuth returns an HTTP basic authentication middleware.
//
// For valid credentials it calls the next handler.
// For invalid Authorization header it sends "404 - Bad Request" response.
// For invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn BasicValidateFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip WebSocket
		if (c.Request().Header.Get(echo.Upgrade)) == echo.WebSocket {
			return nil
		}

		auth := c.Request().Header.Get(echo.Authorization)
		l := len(Basic)
		he := echo.NewHTTPError(http.StatusBadRequest)
		println(auth)

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
						he.SetCode(http.StatusUnauthorized)
					}
				}
			}
		}
		return he
	}
}

// JWTAuth returns a JWT authentication middleware.
//
// For valid token it sets JWT claims in the context with key `_claims` and calls
// the next handler.
// For invalid Authorization header it sends "404 - Bad Request" response.
// For invalid credentials, it sends "401 - Unauthorized" response.
func JWTAuth(fn JWTValidateFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Skip WebSocket
		if (c.Request().Header.Get(echo.Upgrade)) == echo.WebSocket {
			return nil
		}

		auth := c.Request().Header.Get("Authorization")
		l := len(Bearer)
		he := echo.NewHTTPError(http.StatusBadRequest)

		if len(auth) > l+1 && auth[:l] == Bearer {
			t, err := jwt.Parse(auth[l+1:], func(token *jwt.Token) (interface{}, error) {
				// Lookup key and verify method
				if kid := token.Header["kid"]; kid != nil {
					return fn(kid.(string), token.Method)
				}
				return fn("", token.Method)
			})
			if err == nil && t.Valid {
				c.Set("_claims", t.Claims)
				return nil
			} else {
				he.SetCode(http.StatusUnauthorized)
			}
		}
		return he
	}
}

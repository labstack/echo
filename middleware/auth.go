package middleware

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type (
	BasicAuthFunc       func(string, string) bool
	AuthorizedHandler   echo.HandlerFunc
	UnauthorizedHandler func(*echo.Context, error)
	JwtKeyFunc          func(string) ([]byte, error)
	Claims              map[string]interface{}
)

var (
	ErrBasicAuth = errors.New("echo: basic auth error")
	ErrJwtAuth   = errors.New("echo: jwt auth error")
)

func BasicAuth(ah AuthorizedHandler, uah UnauthorizedHandler, fn BasicAuthFunc) echo.HandlerFunc {
	return func(c *echo.Context) {
		auth := strings.Fields(c.Request.Header.Get("Authorization"))
		if len(auth) == 2 {
			scheme := auth[0]
			s, err := base64.StdEncoding.DecodeString(auth[1])
			if err != nil {
				uah(c, err)
				return
			}
			cred := strings.Split(string(s), ":")
			if scheme == "Basic" && len(cred) == 2 {
				if ok := fn(cred[0], cred[1]); ok {
					ah(c)
					return
				}
			}
		}
		uah(c, ErrBasicAuth)
	}
}

func JwtAuth(ah AuthorizedHandler, uah UnauthorizedHandler, fn JwtKeyFunc) echo.HandlerFunc {
	return func(c *echo.Context) {
		auth := strings.Fields(c.Request.Header.Get("Authorization"))
		if len(auth) == 2 {
			t, err := jwt.Parse(auth[1], func(token *jwt.Token) (interface{}, error) {
				if kid := token.Header["kid"]; kid != nil {
					return fn(kid.(string))
				}
				return fn("")
			})
			if t.Valid {
				c.Set("claims", Claims(t.Claims))
				ah(c)
				// c.Next()
			} else {
				// TODO: capture errors
				uah(c, err)
			}
			return
		}
		uah(c, ErrJwtAuth)
	}
}

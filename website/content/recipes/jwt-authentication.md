---
title: JWT Authentication
menu:
  main:
    parent: recipes
---

Most applications dealing with client authentication will require a more secure
mechanism than that provided by [basic authentication](https://github.com/labstack/echo/blob/master/middleware/auth.go). [JSON Web Tokens](http://jwt.io/)
are one such mechanism - JWTs are a compact means of transferring cryptographically
signed claims between the client and server.

This recipe demonstrates the use of a simple JWT authentication Echo middleware
using Dave Grijalva's [jwt-go](https://github.com/dgrijalva/jwt-go). This middleware
expects the token to be present in an Authorization HTTP header using the method
"Bearer", although JWTs are also frequently sent using cookies, the request URL,
or even the request body. We will use the HS236 signing method, note that several
other algorithms are available.

`server.go`

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

const (
	Bearer     = "Bearer"
	SigningKey = "somethingsupersecret"
)

// A JSON Web Token middleware
func JWTAuth(key string) echo.HandlerFunc {
	return func(c *echo.Context) error {

		// Skip WebSocket
		if (c.Request().Header.Get(echo.Upgrade)) == echo.WebSocket {
			return nil
		}

		auth := c.Request().Header.Get("Authorization")
		l := len(Bearer)
		he := echo.NewHTTPError(http.StatusUnauthorized)

		if len(auth) > l+1 && auth[:l] == Bearer {
			t, err := jwt.Parse(auth[l+1:], func(token *jwt.Token) (interface{}, error) {

				// Always check the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// Return the key for validation
				return []byte(key), nil
			})
			if err == nil && t.Valid {
				// Store token claims in echo.Context
				c.Set("claims", t.Claims)
				return nil
			}
		}
		return he
	}
}

func accessible(c *echo.Context) error {
	return c.String(http.StatusOK, "No auth required for this route.\n")
}

func restricted(c *echo.Context) error {
	return c.String(http.StatusOK, "Access granted with JWT.\n")
}

func main() {

	// Echo instance
	e := echo.New()

	// Logger
	e.Use(mw.Logger())

	// Unauthenticated route
	e.Get("/", accessible)

	// Restricted group
	r := e.Group("/restricted")
	r.Use(JWTAuth(SigningKey))
	r.Get("", restricted)

	// Start server
	e.Run(":1323")
}
```

Run `server.go` and making a request to the root path `/` returns a 200 OK response,
as this route does not use our JWT authentication middleware. Sending requests to
`/restricted` (our authenticated route) with either no Authorization header or invalid
Authorization headers / tokens will return 401 Unauthorized.

```sh
# Unauthenticated route
$ curl localhost:1323/  => No auth required for this route.

# No Authentication header
$ curl localhost:1323/restricted  => Unauthorized

# Invalid Authentication method
$  curl localhost:1323/restricted -H "Authorization: Invalid " => Unauthorized

# Invalid token
$  curl localhost:1323/restricted -H "Authorization: Bearer InvalidToken" => Unauthorized
```

Running `token.go` (source) will print JWT that is valid against this middleware
to stdout. You can use this token to test succesful authentication on the `/restricted` path.

```go
package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const SigningKey = "somethingsupersecret"

func main() {

	// New web token.
	token := jwt.New(jwt.SigningMethodHS256)

	// Set a header and a claim
	token.Header["typ"] = "JWT"
	token.Claims["exp"] = time.Now().Add(time.Hour * 96).Unix()

	// Generate encoded token
	t, _ := token.SignedString([]byte(SigningKey))
	fmt.Println(t)
}
```

```sh
# Valid token
$  curl localhost:1323/restricted -H "Authorization: Bearer <token>" => Access granted with JWT.
```

## [Source Code](https://github.com/labstack/echo/blob/master/recipes/jwt-authentication)

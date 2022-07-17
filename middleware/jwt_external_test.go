package middleware_test

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"net/http"
	"net/http/httptest"
)

// CreateJWTGoParseTokenFunc creates JWTGo implementation for ParseTokenFunc
//
// signingKey is signing key to validate token.
// This is one of the options to provide a token validation key.
// The order of precedence is a user-defined SigningKeys and SigningKey.
// Required if signingKeys is not provided.
//
// signingKeys is Map of signing keys to validate token with kid field usage.
// This is one of the options to provide a token validation key.
// The order of precedence is a user-defined SigningKeys and SigningKey.
// Required if signingKey is not provided
func CreateJWTGoParseTokenFunc(signingKey interface{}, signingKeys map[string]interface{}) func(c echo.Context, auth string, source middleware.ExtractorSource) (interface{}, error) {
	// keyFunc defines a user-defined function that supplies the public key for a token validation.
	// The function shall take care of verifying the signing algorithm and selecting the proper key.
	// A user-defined KeyFunc can be useful if tokens are issued by an external party.
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != middleware.AlgorithmHS256 {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		if len(signingKeys) == 0 {
			return signingKey, nil
		}

		if kid, ok := t.Header["kid"].(string); ok {
			if key, ok := signingKeys[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
	}

	return func(c echo.Context, auth string, source middleware.ExtractorSource) (interface{}, error) {
		token, err := jwt.ParseWithClaims(auth, jwt.MapClaims{}, keyFunc) // you could add your default claims here
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			return nil, errors.New("invalid token")
		}
		return token, nil
	}
}

func ExampleJWTConfig_withJWTGoAsTokenParser() {
	mw := middleware.JWTWithConfig(middleware.JWTConfig{
		ParseTokenFunc: CreateJWTGoParseTokenFunc([]byte("secret"), nil),
	})

	e := echo.New()
	e.Use(mw)

	e.GET("/", func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		return c.JSON(http.StatusTeapot, user.Claims)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	fmt.Printf("status: %v, body: %v", res.Code, res.Body.String())
	// Output: status: 418, body: {"admin":true,"name":"John Doe","sub":"1234567890"}
}

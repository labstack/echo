package middleware

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"time"
)

func TestBasicAuth(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), echo.New())
	fn := func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}
	ba := BasicAuth(fn)

	// Valid credentials
	auth := Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	assert.NoError(t, ba(c))

	//---------------------
	// Invalid credentials
	//---------------------

	// Incorrect password
	auth = Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:password"))
	req.Header.Set(echo.Authorization, auth)
	he := ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code())

	// Empty Authorization header
	req.Header.Set(echo.Authorization, "")
	he = ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	he = ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// Invalid scheme
	auth = "Base " + base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	he = ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// WebSocket
	c.Request().Header.Set(echo.Upgrade, echo.WebSocket)
	assert.NoError(t, ba(c))
}

func TestJWTAuth(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), echo.New())
	key := []byte("key")
	fn := func(kid string, method jwt.SigningMethod) ([]byte, error) {
		return key, nil
	}
	ja := JWTAuth(fn)
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["foo"] = "bar"
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	ts, err := token.SignedString(key)
	assert.NoError(t, err)

	// Valid credentials
	auth := Bearer + " " + ts
	req.Header.Set(echo.Authorization, auth)
	assert.NoError(t, ja(c))

	//---------------------
	// Invalid credentials
	//---------------------

	// Expired token
	token.Claims["exp"] = time.Now().Add(-time.Second).Unix()
	ts, err = token.SignedString(key)
	assert.NoError(t, err)
	auth = Bearer + " " + ts
	req.Header.Set(echo.Authorization, auth)
	he := ja(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code())

	// Empty Authorization header
	req.Header.Set(echo.Authorization, "")
	he = ja(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// Invalid Authorization header
	auth = "token"
	req.Header.Set(echo.Authorization, auth)
	he = ja(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// Invalid scheme
	auth = "Bear token"
	req.Header.Set(echo.Authorization, auth)
	he = ja(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code())

	// WebSocket
	c.Request().Header.Set(echo.Upgrade, echo.WebSocket)
	assert.NoError(t, ja(c))
}

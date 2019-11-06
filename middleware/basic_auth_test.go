package middleware

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	f := func(u, p string, c echo.Context) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		return false, nil
	}
	h := BasicAuth(f)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	assert := assert.New(t)

	// Valid credentials
	auth := basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(h(c))

	h = BasicAuthWithConfig(BasicAuthConfig{
		Skipper:   nil,
		Validator: f,
		Realm:     "someRealm",
	})(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Valid credentials
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(h(c))

	// Case-insensitive header scheme
	auth = strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(h(c))

	// Invalid credentials
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:invalid-password"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	he := h(c).(*echo.HTTPError)
	assert.Equal(http.StatusUnauthorized, he.Code)
	assert.Equal(basic+` realm="someRealm"`, res.Header().Get(echo.HeaderWWWAuthenticate))

	// Missing Authorization header
	req.Header.Del(echo.HeaderAuthorization)
	he = h(c).(*echo.HTTPError)
	assert.Equal(http.StatusUnauthorized, he.Code)

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	he = h(c).(*echo.HTTPError)
	assert.Equal(http.StatusUnauthorized, he.Code)

	// Invalid Authorization header when decode header error
	auth = strings.ToUpper(basic) + " invalid base64"
	req.Header.Set(echo.HeaderAuthorization, auth)
	err := h(c)
	assert.Error(err, "Expected an error, but doesn't happen")

	errString := "invalid auth"
	ef := func(u, p string, c echo.Context) (bool, error) {
		return false, errors.New(errString)
	}
	h = BasicAuthWithConfig(BasicAuthConfig{
		Skipper:   nil,
		Validator: ef,
		Realm:     "someRealm",
	})(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	// Invalid Authorization header when validator returns error
	auth = strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	err = h(c)
	assert.EqualError(err, errString)
}

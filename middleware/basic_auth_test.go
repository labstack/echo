package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
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

	// Valid credentials
	auth := basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

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
	assert.NoError(t, h(c))

	// Case-insensitive header scheme
	auth = strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

	// Invalid credentials
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:invalid-password"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+` realm="someRealm"`, res.Header().Get(echo.HeaderWWWAuthenticate))

	// Missing Authorization header
	req.Header.Del(echo.HeaderAuthorization)
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

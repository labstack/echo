package middleware

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	f := func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}
	h := BasicAuth(f)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Valid credentials
	auth := basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header().Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

	// Incorrect password
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:password"))
	req.Header().Set(echo.HeaderAuthorization, auth)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+" realm=Restricted", res.Header().Get(echo.HeaderWWWAuthenticate))

	// Empty Authorization header
	req.Header().Set(echo.HeaderAuthorization, "")
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	req.Header().Set(echo.HeaderAuthorization, auth)
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

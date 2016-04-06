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
	rq := test.NewRequest(echo.GET, "/", nil)
	rs := test.NewResponseRecorder()
	c := echo.NewContext(rq, rs, e)
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
	rq.Header().Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

	//---------------------
	// Invalid credentials
	//---------------------

	// Incorrect password
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:password"))
	rq.Header().Set(echo.HeaderAuthorization, auth)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+" realm=Restricted", rs.Header().Get(echo.HeaderWWWAuthenticate))

	// Empty Authorization header
	rq.Header().Set(echo.HeaderAuthorization, "")
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+" realm=Restricted", rs.Header().Get(echo.HeaderWWWAuthenticate))

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	rq.Header().Set(echo.HeaderAuthorization, auth)
	he = h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+" realm=Restricted", rs.Header().Get(echo.HeaderWWWAuthenticate))
}

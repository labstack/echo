package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
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
	auth := echo.Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	assert.NoError(t, ba(c))

	//---------------------
	// Invalid credentials
	//---------------------

	// Incorrect password
	auth = echo.Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:password"))
	req.Header.Set(echo.Authorization, auth)
	he := ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code())

	// Empty Authorization header
	req.Header.Set(echo.Authorization, "")
	he = ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code())

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	req.Header.Set(echo.Authorization, auth)
	he = ba(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code())

	// WebSocket
	c.Request().Header.Set(echo.Upgrade, echo.WebSocket)
	assert.NoError(t, ba(c))
}

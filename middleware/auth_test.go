package middleware

import (
	"encoding/base64"
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	req, _ := http.NewRequest(echo.POST, "/", nil)
	res := &echo.Response{Writer: httptest.NewRecorder()}
	c := echo.NewContext(req, res, echo.New())
	fn := func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}
	b := BasicAuth(fn)

	//-------------------
	// Valid credentials
	//-------------------

	auth := Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	if b(c) != nil {
		t.Error("basic auth should pass")
	}

	// Case insensitive
	auth = "basic " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	if b(c) != nil {
		t.Error("basic auth should ignore case and pass")
	}

	//---------------------
	// Invalid credentials
	//---------------------

	auth = Basic + "  " + base64.StdEncoding.EncodeToString([]byte(" joe: secret"))
	req.Header.Set(echo.Authorization, auth)
	b = BasicAuth(fn)
	if b(c) == nil {
		t.Error("basic auth should fail")
	}

	// Invalid scheme
	auth = "Base " + base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	b = BasicAuth(fn)
	if b(c) == nil {
		t.Error("basic auth should fail for invalid scheme")
	}

	// Empty auth header
	b = BasicAuth(fn)
	if b(c) == nil {
		t.Error("basic auth should fail for empty auth header")
	}
}

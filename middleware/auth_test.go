package middleware

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/labstack/echo"
)

func TestBasicAuth(t *testing.T) {
	req, _ := http.NewRequest(echo.POST, "/", nil)
	res := &echo.Response{}
	c := echo.NewContext(req, res, echo.New())
	fn := func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}
	ba := BasicAuth(fn)

	//-------------------
	// Valid credentials
	//-------------------

	auth := Basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	if ba(c) != nil {
		t.Error("expected `pass`")
	}

	// Case insensitive
	auth = "basic " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.Authorization, auth)
	if ba(c) != nil {
		t.Error("expected `pass`, with case insensitive header.")
	}

	//---------------------
	// Invalid credentials
	//---------------------

	// Incorrect password
	auth = Basic + "  " + base64.StdEncoding.EncodeToString([]byte("joe: password"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	if ba(c) == nil {
		t.Error("expected `fail`, with incorrect password.")
	}

	// Empty Authorization header
	req.Header.Set(echo.Authorization, "")
	ba = BasicAuth(fn)
	if ba(c) == nil {
		t.Error("expected `fail`, with empty Authorization header.")
	}

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	if ba(c) == nil {
		t.Error("expected `fail`, with invalid Authorization header.")
	}

	// Invalid scheme
	auth = "Base " + base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	if ba(c) == nil {
		t.Error("expected `fail`, with invalid scheme.")
	}

}

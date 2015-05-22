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
	auth = Basic + "  " + base64.StdEncoding.EncodeToString([]byte("joe:password"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	he := ba(c).(*echo.HTTPError)
	if ba(c) == nil {
		t.Error("expected `fail`, with incorrect password.")
	} else if he.Code() != http.StatusUnauthorized {
		t.Errorf("expected status `401`, got %d", he.Code())
	}

	// Empty Authorization header
	req.Header.Set(echo.Authorization, "")
	ba = BasicAuth(fn)
	he = ba(c).(*echo.HTTPError)
	if he == nil {
		t.Error("expected `fail`, with empty Authorization header.")
	} else if he.Code() != http.StatusBadRequest {
		t.Errorf("expected status `400`, got %d", he.Code())
	}

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	he = ba(c).(*echo.HTTPError)
	if he == nil {
		t.Error("expected `fail`, with invalid Authorization header.")
	} else if he.Code() != http.StatusBadRequest {
		t.Errorf("expected status `400`, got %d", he.Code())
	}

	// Invalid scheme
	auth = "Ace " + base64.StdEncoding.EncodeToString([]byte(" :secret"))
	req.Header.Set(echo.Authorization, auth)
	ba = BasicAuth(fn)
	he = ba(c).(*echo.HTTPError)
	if he == nil {
		t.Error("expected `fail`, with invalid scheme.")
	} else if he.Code() != http.StatusBadRequest {
		t.Errorf("expected status `400`, got %d", he.Code())
	}
}

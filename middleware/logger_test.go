package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestParseIP(t *testing.T) {
	var cases = map[string]string{
		"127.0.0.1":      "127.0.0.1",
		"127.0.0.1:8080": "127.0.0.1",
		"::1":            "::1",
		"2001:4860:0:2001::68":        "2001:4860:0:2001::68",
		"[2001:4860:0:2001::68]:8001": "2001:4860:0:2001::68",
		"invalid":                     "unknown",
	}

	for k, v := range cases {
		if parseIP(k) != v {
			t.Errorf("%s should equal %s, but doesn't", parseIP(k), v)
		}
	}
}
func TestRealIPHeader(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	req.Header.Add("X-Real-IP", "127.0.0.1")
	req.Header.Add("X-Forwarded-For", "127.0.0.1")
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), e)

	// Status 2xx
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)
}

func TestLogger(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), e)

	// Status 2xx
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)

	// Status 3xx
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec), e)
	h = func(c *echo.Context) error {
		return c.String(http.StatusTemporaryRedirect, "test")
	}
	Logger()(h)(c)

	// Status 4xx
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec), e)
	h = func(c *echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	}
	Logger()(h)(c)

	// Status 5xx with empty path
	req, _ = http.NewRequest(echo.GET, "", nil)
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec), e)
	h = func(c *echo.Context) error {
		return errors.New("error")
	}
	Logger()(h)(c)
}

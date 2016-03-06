package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec, e), e)

	// Status 2xx
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)

	// Status 3xx
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec, e), e)
	h = func(c *echo.Context) error {
		return c.String(http.StatusTemporaryRedirect, "test")
	}
	Logger()(h)(c)

	// Status 4xx
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec, e), e)
	h = func(c *echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	}
	Logger()(h)(c)

	// Status 5xx with empty path
	req, _ = http.NewRequest(echo.GET, "", nil)
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec, e), e)
	h = func(c *echo.Context) error {
		return errors.New("error")
	}
	Logger()(h)(c)
}

func TestLoggerIPAddress(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec, e), e)
	buf := new(bytes.Buffer)
	e.Logger().SetOutput(buf)
	ip := "127.0.0.1"
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	mw := Logger()

	// With X-Real-IP
	req.Header.Add(echo.XRealIP, ip)
	mw(h)(c)
	assert.Contains(t, buf.String(), ip)

	// With X-Forwarded-For
	buf.Reset()
	req.Header.Del(echo.XRealIP)
	req.Header.Add(echo.XForwardedFor, ip)
	mw(h)(c)
	assert.Contains(t, buf.String(), ip)

	// with req.RemoteAddr
	buf.Reset()
	mw(h)(c)
	assert.Contains(t, buf.String(), ip)
}

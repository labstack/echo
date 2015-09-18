package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), e)

	// With X-Real-IP
	req.Header.Add(echo.XRealIP, "127.0.0.1")
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)

	// With X-Forwarded-For
	req.Header.Del(echo.XRealIP)
	req.Header.Add(echo.XForwardedFor, "127.0.0.1")
	h = func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)

	// Status 2xx
	h = func(c *echo.Context) error {
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

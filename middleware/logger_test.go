package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

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

package middleware

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	w := httptest.NewRecorder()
	res := echo.NewResponse(w)
	c := echo.NewContext(req, res, e)

	// Status 2xx
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Logger()(h)(c)

	// Status 4xx
	c.Response = echo.NewResponse(w)
	h = func(c *echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	}
	Logger()(h)(c)

	// Status 5xx
	c.Response = echo.NewResponse(w)
	h = func(c *echo.Context) error {
		return c.String(http.StatusInternalServerError, "test")
	}
	Logger()(h)(c)
}

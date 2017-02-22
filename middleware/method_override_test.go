package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestMethodOverride(t *testing.T) {
	e := echo.New()
	m := MethodOverride()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with http header
	req := httptest.NewRequest(echo.POST, "/", nil)
	rec := httptest.NewRecorder()
	req.Header.Set(echo.HeaderXHTTPMethodOverride, echo.DELETE)
	c := e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method)

	// Override with form parameter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromForm("_method")})
	req, _ = http.NewRequest(echo.POST, "/", bytes.NewReader([]byte("_method="+echo.DELETE)))
	rec = httptest.NewRecorder()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c = e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method)

	// Override with query parameter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromQuery("_method")})
	req, _ = http.NewRequest(echo.POST, "/?_method="+echo.DELETE, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method)

	// Ignore `GET`
	req, _ = http.NewRequest(echo.GET, "/", nil)
	req.Header.Set(echo.HeaderXHTTPMethodOverride, echo.DELETE)
	assert.Equal(t, echo.GET, req.Method)
}

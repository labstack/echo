package middleware

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestMethodOverride(t *testing.T) {
	e := echo.New()
	m := MethodOverride()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with http header
	req := test.NewRequest(echo.POST, "/", nil)
	rec := test.NewResponseRecorder()
	req.Header().Set(echo.HeaderXHTTPMethodOverride, echo.DELETE)
	c := e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method())

	// Override with form parameter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromForm("_method")})
	req = test.NewRequest(echo.POST, "/", bytes.NewReader([]byte("_method="+echo.DELETE)))
	rec = test.NewResponseRecorder()
	req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c = e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method())

	// Override with query paramter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromQuery("_method")})
	req = test.NewRequest(echo.POST, "/?_method="+echo.DELETE, nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	m(h)(c)
	assert.Equal(t, echo.DELETE, req.Method())

	// Ignore `GET`
	req = test.NewRequest(echo.GET, "/", nil)
	req.Header().Set(echo.HeaderXHTTPMethodOverride, echo.DELETE)
	assert.Equal(t, echo.GET, req.Method())
}

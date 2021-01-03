// +build go1.13

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Test for SameSiteModeNone moved to separate file for Go 1.12 support
func TestCSRFWithSameSiteModeNone(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	csrf := CSRFWithConfig(CSRFConfig{
		CookieSameSite: SameSiteNoneMode,
	})

	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	r := h(c)
	assert.NoError(t, r)
	assert.Regexp(t, "SameSite=None", rec.Header()["Set-Cookie"])
	assert.Regexp(t, "Secure", rec.Header()["Set-Cookie"])
}

package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestCSRF(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	csrf := CSRF([]byte("secret"))
	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Generate CSRF token
	h(c)
	assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), "csrf")

	// Empty/invalid CSRF token
	req = test.NewRequest(echo.POST, "/", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	req.Header().Set(echo.HeaderXCSRFToken, "")
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusForbidden, he.Code)

	// Valid CSRF token
	salt, _ := generateSalt(8)
	token := generateCSRFToken([]byte("secret"), salt)
	req.Header().Set(echo.HeaderXCSRFToken, token)
	h(c)
	assert.Equal(t, http.StatusOK, rec.Status())
}

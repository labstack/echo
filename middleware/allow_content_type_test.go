package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAllowContentType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	h := AllowContentType("application/json")(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Test valid content type
	req.Header.Add("Content-Type", "application/json")
	assert.NoError(t, h(c))

	// Test invalid content type
	req.Header.Add("Content-Type", "text/plain")
	assert.NoError(t, h(c))
}

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

	h := AllowContentType("application/json")(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Test valid content type
	req.Header.Add("Content-Type", "application/json")
	c := e.NewContext(req, res)
	assert.NoError(t, h(c))

	// Test invalid content type
	req.Header.Set("Content-Type", "text/plain")
	c = e.NewContext(req, res)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusUnsupportedMediaType, he.Code)
}

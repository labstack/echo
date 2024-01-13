package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAllowContentType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("Hello World!")))
	res := httptest.NewRecorder()

	h := AllowContentType("application/json", "text/plain")(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Test valid content type
	req.Header.Add("Content-Type", "application/json")

	c := e.NewContext(req, res)
	assert.NoError(t, h(c))

	// Test invalid content type
	req.Header.Add("Content-Type", "text/html")
	c = e.NewContext(req, res)
	assert.NoError(t, h(c))
}

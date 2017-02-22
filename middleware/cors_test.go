package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	e := echo.New()

	// Wildcard origin
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := CORS()(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Allow origins
	req, _ = http.NewRequest(echo.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"localhost"},
	})(echo.NotFoundHandler)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Preflight request
	req, _ = http.NewRequest(echo.OPTIONS, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cors := CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(echo.HeaderAccessControlMaxAge))
}

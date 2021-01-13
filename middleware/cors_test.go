package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo"
	"github.com/trafficstars/echo/test"
)

func TestCORS(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	cors := CORSWithConfig(CORSConfig{
		AllowCredentials: true,
	})
	h := cors(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Wildcard origin
	req = test.NewRequest(echo.GET, "/", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	req.Header().Set(echo.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Simple request
	req = test.NewRequest(echo.GET, "/", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	req.Header().Set(echo.HeaderOrigin, "localhost")
	cors = CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Preflight request
	req = test.NewRequest(echo.OPTIONS, "/", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	req.Header().Set(echo.HeaderOrigin, "localhost")
	req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(echo.HeaderAccessControlMaxAge))
}

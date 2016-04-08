package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	e := echo.New()
	rq := test.NewRequest(echo.GET, "/", nil)
	rc := test.NewResponseRecorder()
	c := echo.NewContext(rq, rc, e)
	cors := CORSWithConfig(CORSConfig{
		AllowCredentials: true,
	})
	h := cors(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// No origin header
	h(c)
	assert.Equal(t, "", rc.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Wildcard origin
	rq = test.NewRequest(echo.GET, "/", nil)
	rc = test.NewResponseRecorder()
	c = echo.NewContext(rq, rc, e)
	rq.Header().Set(echo.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "*", rc.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Simple request
	rq = test.NewRequest(echo.GET, "/", nil)
	rc = test.NewResponseRecorder()
	c = echo.NewContext(rq, rc, e)
	rq.Header().Set(echo.HeaderOrigin, "localhost")
	cors = CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	h(c)
	assert.Equal(t, "localhost", rc.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Preflight request
	rq = test.NewRequest(echo.OPTIONS, "/", nil)
	rc = test.NewResponseRecorder()
	c = echo.NewContext(rq, rc, e)
	rq.Header().Set(echo.HeaderOrigin, "localhost")
	rq.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	h(c)
	assert.Equal(t, "localhost", rc.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rc.Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rc.Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rc.Header().Get(echo.HeaderAccessControlMaxAge))
}

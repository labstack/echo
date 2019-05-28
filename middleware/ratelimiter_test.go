package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

)

func TestRateLimiter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rateLimit := RateLimiter()

	h := rateLimit(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// g
	h(c)
	assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "99")

	//
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	assert.Error(t, h(c))


}



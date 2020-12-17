package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	var inMemoryStore = new(InMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	testCases := []struct {
		id   string
		code int
	}{
		{"127.0.0.1", 200},
		{"127.0.0.1", 200},
		{"127.0.0.1", 200},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		mw := RateLimiter(func(c echo.Context) string {
			return c.Request().Header.Get(echo.HeaderXRealIP)
		}, inMemoryStore)

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig(t *testing.T) {
	var inMemoryStore = new(InMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	testCases := []struct {
		id   string
		code int
	}{
		{"127.0.0.1", 200},
		{"127.0.0.1", 200},
		{"127.0.0.1", 200},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
		{"127.0.0.1", 429},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		mw := RateLimiterWithConfig(RateLimiterConfig{
			SourceFunc: func(c echo.Context) string {
				return c.Request().Header.Get(echo.HeaderXRealIP)
			},
			Store: inMemoryStore,
		})

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestInMemoryStore_ShouldAllow(t *testing.T) {
	var inMemoryStore = new(InMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	testCases := []struct {
		id      string
		allowed bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.1", true},
		{"127.0.0.1", true},
		{"127.0.0.1", false},
		{"127.0.0.1", false},
		{"127.0.0.1", false},
		{"127.0.0.1", false},
	}

	for _, tc := range testCases {
		allowed := inMemoryStore.ShouldAllow(tc.id)

		assert.Equal(t, tc.allowed, allowed)
	}
}

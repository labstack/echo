package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	{
		var inMemoryStore = new(RateLimiterMemoryStore)
		inMemoryStore.rate = 1
		inMemoryStore.burst = 3

		assert.Panics(t, func() {
			RateLimiter(nil)
		})

		assert.NotPanics(t, func() {
			RateLimiter(inMemoryStore)
		})
	}

	{
		var skipped bool
		var inMemoryStore = new(RateLimiterMemoryStore)
		inMemoryStore.rate = 1
		inMemoryStore.burst = 3

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		mw := RateLimiterWithConfig(RateLimiterConfig{
			Skipper: func(c echo.Context) bool {
				skipped = true
				return true
			},
			Store: inMemoryStore,
			SourceFunc: func(ctx echo.Context) string {
				return "127.0.0.1"
			},
		})

		_ = mw(handler)(c)

		assert.Equal(t, true, skipped)
	}

	{
		var beforeRan bool
		var inMemoryStore = new(RateLimiterMemoryStore)
		inMemoryStore.rate = 1
		inMemoryStore.burst = 3

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		mw := RateLimiterWithConfig(RateLimiterConfig{
			BeforeFunc: func(c echo.Context) {
				beforeRan = true
			},
			Store: inMemoryStore,
			SourceFunc: func(ctx echo.Context) string {
				return "127.0.0.1"
			},
		})

		_ = mw(handler)(c)

		assert.Equal(t, true, beforeRan)
	}

	var inMemoryStore = new(RateLimiterMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	testCases := []struct {
		id   string
		code int
	}{
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusOK},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
		{"127.0.0.1", http.StatusTooManyRequests},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		mw := RateLimiter(inMemoryStore)

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig(t *testing.T) {
	{
		var inMemoryStore = new(RateLimiterMemoryStore)
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
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add(echo.HeaderXRealIP, tc.id)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			mw := RateLimiterWithConfig(RateLimiterConfig{
				SourceFunc: func(c echo.Context) string {
					return c.RealIP()
				},
				ErrorHandler: func(ctx echo.Context) error {
					return ctx.JSON(http.StatusTooManyRequests, nil)
				},
				Store: inMemoryStore,
			})

			_ = mw(handler)(c)

			assert.Equal(t, tc.code, rec.Code)
		}
	}
	{
		var inMemoryStore = new(RateLimiterMemoryStore)
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
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusOK},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
			{"127.0.0.1", http.StatusTooManyRequests},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add(echo.HeaderXRealIP, tc.id)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			mw := RateLimiterWithConfig(RateLimiterConfig{
				Store: inMemoryStore,
			})

			_ = mw(handler)(c)

			assert.Equal(t, tc.code, rec.Code)
		}
	}
}

func TestRateLimiterMemoryStore_Allow(t *testing.T) {
	var inMemoryStore = new(RateLimiterMemoryStore)
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
		allowed := inMemoryStore.Allow(tc.id)

		assert.Equal(t, tc.allowed, allowed)
	}
}

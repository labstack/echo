package middleware

import (
	"errors"
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
		println(c.RealIP())
		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiter_panicBehaviour(t *testing.T) {
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

func TestRateLimiterWithConfig(t *testing.T) {
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
		{"127.0.0.1", http.StatusForbidden},
		{"", http.StatusBadRequest},
		{"127.0.0.1", http.StatusForbidden},
		{"127.0.0.1", http.StatusForbidden},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		mw := RateLimiterWithConfig(RateLimiterConfig{
			IdentifierExtractor: func(c echo.Context) (string, error) {
				id := c.Request().Header.Get(echo.HeaderXRealIP)
				if id == "" {
					return "", errors.New("invalid identifier")
				}
				return id, nil
			},
			DenyHandler: func(ctx echo.Context) error {
				return ctx.JSON(http.StatusBadRequest, nil)
			},
			ErrorHandler: func(ctx echo.Context) error {
				return ctx.JSON(http.StatusForbidden, nil)
			},
			Store: inMemoryStore,
		})

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig_defaultDenyHandler(t *testing.T) {
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
		{"127.0.0.1", http.StatusForbidden},
		{"", http.StatusForbidden},
		{"127.0.0.1", http.StatusForbidden},
		{"127.0.0.1", http.StatusForbidden},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add(echo.HeaderXRealIP, tc.id)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		mw := RateLimiterWithConfig(RateLimiterConfig{
			IdentifierExtractor: func(c echo.Context) (string, error) {
				id := c.Request().Header.Get(echo.HeaderXRealIP)
				if id == "" {
					return "", errors.New("invalid identifier")
				}
				return id, nil
			},
			ErrorHandler: func(ctx echo.Context) error {
				return ctx.JSON(http.StatusForbidden, nil)
			},
			Store: inMemoryStore,
		})

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig_defaultConfig(t *testing.T) {
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

func TestRateLimiterWithConfig_skipper(t *testing.T) {
	e := echo.New()

	var skipped bool
	handler := func(c echo.Context) error {
		skipped = true
		return c.String(http.StatusOK, "test")
	}
	var inMemoryStore = new(RateLimiterMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := RateLimiterWithConfig(RateLimiterConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		Store: inMemoryStore,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return "127.0.0.1", nil
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, true, skipped)
}

func TestRateLimiterWithConfig_beforeFunc(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

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
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return "127.0.0.1", nil
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, true, beforeRan)
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

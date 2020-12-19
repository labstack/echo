package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

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
		mw := RateLimiter(&inMemoryStore)

		_ = mw(handler)(c)
		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiter_panicBehaviour(t *testing.T) {
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

	assert.Panics(t, func() {
		RateLimiter(nil)
	})

	assert.NotPanics(t, func() {
		RateLimiter(&inMemoryStore)
	})
}

func TestRateLimiterWithConfig(t *testing.T) {
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

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
			Store: &inMemoryStore,
		})

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig_defaultDenyHandler(t *testing.T) {
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

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
			Store: &inMemoryStore,
		})

		_ = mw(handler)(c)

		assert.Equal(t, tc.code, rec.Code)
	}
}

func TestRateLimiterWithConfig_defaultConfig(t *testing.T) {
	{
		var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

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
				Store: &inMemoryStore,
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
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := RateLimiterWithConfig(RateLimiterConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		Store: &inMemoryStore,
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
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := RateLimiterWithConfig(RateLimiterConfig{
		BeforeFunc: func(c echo.Context) {
			beforeRan = true
		},
		Store: &inMemoryStore,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return "127.0.0.1", nil
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, true, beforeRan)
}

func rateProducer(interval time.Duration, count int, f func(i int)) {
	ticker := time.NewTicker(interval)
	i := 0
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				f(i)
				if i < count {
					i++
					continue
				}
				close(quit)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	<-quit
}

func TestRateLimiterMemoryStore_Allow(t *testing.T) {
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3, 2*time.Second)

	testCases := []struct {
		id      string
		allowed bool
	}{
		{"127.0.0.1", true},  // 0 ms
		{"127.0.0.1", true},  // 220 ms burst #2
		{"127.0.0.1", true},  // 440 ms burst #3
		{"127.0.0.1", false}, // 660 ms block
		{"127.0.0.1", false}, // 880 ms block
		{"127.0.0.1", true},  // 1100 ms next second #1
		{"127.0.0.2", true},  // 1320 ms allow other ip
		{"127.0.0.1", false}, // 1540 ms no burst
		{"127.0.0.1", false}, // 1760 ms no burst
		{"127.0.0.1", false}, // 1980 ms no burst
		{"127.0.0.1", true},  // 2200 ms no burst
		{"127.0.0.1", false}, // 2420 ms no burst
		{"127.0.0.1", false}, // 2640 ms no burst
		{"127.0.0.1", false}, // 2860 ms no burst
		{"127.0.0.1", true},  // 3080 ms no burst
		{"127.0.0.1", false}, // 3300 ms no burst
	}

	f := func(i int) {
		t.Logf("Running testcase #%d", i)
		tc := testCases[i]
		allowed := inMemoryStore.Allow(tc.id)
		assert.Equal(t, tc.allowed, allowed)
	}
	rateProducer(220*time.Millisecond, len(testCases)-1, f)
}

func TestRateLimiterMemoryStore_cleanupStaleVisitors(t *testing.T) {
	var inMemoryStore = NewRateLimiterMemoryStore(1, 3)
	inMemoryStore.visitors = map[string]Visitor{
		"A": {
			Limiter:  rate.NewLimiter(1, 3),
			lastSeen: time.Now(),
		},
		"B": {
			Limiter:  rate.NewLimiter(1, 3),
			lastSeen: time.Now().Add(-1 * time.Minute),
		},
		"C": {
			Limiter:  rate.NewLimiter(1, 3),
			lastSeen: time.Now().Add(-5 * time.Minute),
		},
		"D": {
			Limiter:  rate.NewLimiter(1, 3),
			lastSeen: time.Now().Add(-10 * time.Minute),
		},
	}

	inMemoryStore.cleanupStaleVisitors()

	var exists bool

	_, exists = inMemoryStore.visitors["A"]
	assert.Equal(t, true, exists)

	_, exists = inMemoryStore.visitors["B"]
	assert.Equal(t, true, exists)

	_, exists = inMemoryStore.visitors["C"]
	assert.Equal(t, false, exists)

	_, exists = inMemoryStore.visitors["D"]
	assert.Equal(t, false, exists)
}

func TestNewRateLimiterMemoryStore(t *testing.T) {
	testCases := []struct {
		rate              rate.Limit
		burst             int
		expiresIn         time.Duration
		expectedExpiresIn time.Duration
	}{
		{1, 3, 5 * time.Second, 5 * time.Second},
		{2, 4, 0, 3 * time.Minute},
		{1, 5, 10 * time.Minute, 10 * time.Minute},
		{3, 7, 0, 3 * time.Minute},
	}

	for _, tc := range testCases {
		store := NewRateLimiterMemoryStore(tc.rate, tc.burst, tc.expiresIn)
		assert.Equal(t, tc.rate, store.rate)
		assert.Equal(t, tc.burst, store.burst)
		assert.Equal(t, tc.expectedExpiresIn, store.expiresIn)
	}
}

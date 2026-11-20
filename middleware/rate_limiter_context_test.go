// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ctxAwareStore implements both Allow and the optional AllowContext. AllowContext
// gives the store the request context so it can set response headers (e.g.
// Retry-After / X-RateLimit-*) — see #2961.
type ctxAwareStore struct {
	allowCalled    bool
	ctxAllowCalled bool
	allow          bool
}

func (s *ctxAwareStore) Allow(identifier string) (bool, error) {
	s.allowCalled = true
	return s.allow, nil
}

func (s *ctxAwareStore) AllowContext(c *echo.Context, identifier string) (bool, error) {
	s.ctxAllowCalled = true
	c.Response().Header().Set("Retry-After", "42")
	return s.allow, nil
}

type detailsAwareStore struct {
	allowCalled        bool
	ctxAllowCalled     bool
	detailsAllowCalled bool
	allow              bool
	metadata           RateLimitMetadata
}

func (s *detailsAwareStore) Allow(identifier string) (bool, error) {
	s.allowCalled = true
	return s.allow, nil
}

func (s *detailsAwareStore) AllowContext(c *echo.Context, identifier string) (bool, error) {
	s.ctxAllowCalled = true
	return s.allow, nil
}

func (s *detailsAwareStore) AllowWithDetails(identifier string) (bool, RateLimitMetadata, error) {
	s.detailsAllowCalled = true
	return s.allow, s.metadata, nil
}

// When the store implements AllowContext, the middleware must call it instead of
// Allow, so the store can set rate-limit headers on the response.
func TestRateLimiter_storeAllowContextIsPreferred(t *testing.T) {
	e := echo.New()
	store := &ctxAwareStore{allow: true}
	mw := RateLimiterWithConfig(RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
	})
	handler := mw(func(c *echo.Context) error { return c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, handler(c))
	assert.True(t, store.ctxAllowCalled, "AllowContext should be called when implemented")
	assert.False(t, store.allowCalled, "Allow should not be called when AllowContext is implemented")
	assert.Equal(t, "42", rec.Header().Get("Retry-After"), "store should be able to set headers via the context")
}

// When the store implements RateLimiterStoreWithDetails, the middleware must call
// AllowWithDetails and automatically emit standard response headers.
func TestRateLimiter_storeWithDetailsIsPreferred(t *testing.T) {
	fixedTime := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.UTC)
	meta := RateLimitMetadata{
		Limit:      100,
		Remaining:  95,
		Reset:      30 * time.Second,
		ResetTime:  fixedTime.Add(30 * time.Second),
		RetryAfter: 0,
	}
	store := &detailsAwareStore{
		allow:    true,
		metadata: meta,
	}

	e := echo.New()
	mw := RateLimiterWithConfig(RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
	})
	handler := mw(func(c *echo.Context) error { return c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, handler(c))
	assert.True(t, store.detailsAllowCalled)
	assert.False(t, store.ctxAllowCalled)
	assert.False(t, store.allowCalled)

	assert.Equal(t, "100", rec.Header().Get(HeaderXRateLimitLimit))
	assert.Equal(t, "95", rec.Header().Get(HeaderXRateLimitRemaining))
	assert.Equal(t, strconv.FormatInt(fixedTime.Add(30*time.Second).Unix(), 10), rec.Header().Get(HeaderXRateLimitReset))
	assert.Empty(t, rec.Header().Get(echo.HeaderRetryAfter))
}

func TestRateLimiter_storeWithDetails_DenyEmitsRetryAfter(t *testing.T) {
	fixedTime := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.UTC)
	meta := RateLimitMetadata{
		Limit:      10,
		Remaining:  0,
		Reset:      5 * time.Second,
		ResetTime:  fixedTime.Add(5 * time.Second),
		RetryAfter: 3 * time.Second,
	}
	store := &detailsAwareStore{
		allow:    false,
		metadata: meta,
	}

	e := echo.New()
	mw := RateLimiterWithConfig(RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
	})
	handler := mw(func(c *echo.Context) error { return c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.Error(t, err)

	assert.Equal(t, "10", rec.Header().Get(HeaderXRateLimitLimit))
	assert.Equal(t, "0", rec.Header().Get(HeaderXRateLimitRemaining))
	assert.Equal(t, strconv.FormatInt(fixedTime.Add(5*time.Second).Unix(), 10), rec.Header().Get(HeaderXRateLimitReset))
	assert.Equal(t, "3", rec.Header().Get(echo.HeaderRetryAfter))
}

func TestRateLimiter_storeWithDetails_ContextStorage(t *testing.T) {
	meta := RateLimitMetadata{
		Limit:      50,
		Remaining:  42,
		Reset:      15 * time.Second,
		RetryAfter: 0,
	}
	store := &detailsAwareStore{
		allow:    true,
		metadata: meta,
	}

	e := echo.New()
	var extractedMeta RateLimitMetadata
	var foundInHandler bool

	mw := RateLimiterWithConfig(RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
		ContextKey:          "rate_limit",
	})
	handler := mw(func(c *echo.Context) error {
		val := c.Get("rate_limit")
		if v, ok := val.(RateLimitMetadata); ok {
			foundInHandler = true
			extractedMeta = v
		}
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, handler(c))
	assert.True(t, foundInHandler)
	assert.Equal(t, 50, extractedMeta.Limit)
	assert.Equal(t, 42, extractedMeta.Remaining)
}

func TestRateLimiter_storeWithDetails_ZeroResetTimeFallback(t *testing.T) {
	before := time.Now().Unix()
	meta := RateLimitMetadata{
		Limit:      20,
		Remaining:  5,
		Reset:      10 * time.Second,
		RetryAfter: 0,
	}
	store := &detailsAwareStore{
		allow:    true,
		metadata: meta,
	}

	e := echo.New()
	mw := RateLimiterWithConfig(RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
	})
	handler := mw(func(c *echo.Context) error { return c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, handler(c))
	after := time.Now().Unix()

	resetHeader := rec.Header().Get(HeaderXRateLimitReset)
	require.NotEmpty(t, resetHeader)
	resetVal, err := strconv.ParseInt(resetHeader, 10, 64)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resetVal, before+10)
	assert.LessOrEqual(t, resetVal, after+10)
}

// The built-in memory store implements AllowContext and AllowWithDetails, so it sets
// X-RateLimit-Limit / X-RateLimit-Remaining / X-RateLimit-Reset on every request and
// Retry-After when the limit is hit (#2961).
func TestRateLimiterMemoryStore_AllowContextSetsHeaders(t *testing.T) {
	store := NewRateLimiterMemoryStoreWithConfig(RateLimiterMemoryStoreConfig{Rate: 1, Burst: 3})
	e := echo.New()
	e.GET("/", func(c *echo.Context) error { return c.String(http.StatusOK, "ok") },
		RateLimiterWithConfig(RateLimiterConfig{
			Store:               store,
			IdentifierExtractor: func(c *echo.Context) (string, error) { return "id", nil },
		}))

	do := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	// Burst of 3: each allowed request advertises the limit, decreasing remaining, and reset timestamp.
	for i := 0; i < 3; i++ {
		rec := do()
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "3", rec.Header().Get(HeaderXRateLimitLimit))
		assert.Equal(t, strconv.Itoa(2-i), rec.Header().Get(HeaderXRateLimitRemaining))
		assert.NotEmpty(t, rec.Header().Get(HeaderXRateLimitReset))
		assert.Empty(t, rec.Header().Get(echo.HeaderRetryAfter))
	}

	// 4th request is denied: 429, remaining 0, reset timestamp, and a Retry-After hint.
	rec := do()
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "0", rec.Header().Get(HeaderXRateLimitRemaining))
	assert.NotEmpty(t, rec.Header().Get(HeaderXRateLimitReset))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderRetryAfter))
}

func TestRateLimiterMemoryStore_AllowWithDetails(t *testing.T) {
	baseTime := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.UTC)
	store := NewRateLimiterMemoryStoreWithConfig(RateLimiterMemoryStoreConfig{
		Rate:      2,
		Burst:     4,
		ExpiresIn: 1 * time.Minute,
	})
	store.timeNow = func() time.Time {
		return baseTime
	}

	// Request 1: 1 token consumed, 3 remaining
	allowed, meta, err := store.AllowWithDetails("user1")
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 4, meta.Limit)
	assert.Equal(t, 3, meta.Remaining)
	assert.Equal(t, 500*time.Millisecond, meta.Reset)
	assert.Equal(t, baseTime.Add(500*time.Millisecond), meta.ResetTime)
	assert.Equal(t, time.Duration(0), meta.RetryAfter)

	// Request 2: 2 remaining
	allowed, meta, err = store.AllowWithDetails("user1")
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 2, meta.Remaining)
	assert.Equal(t, 1000*time.Millisecond, meta.Reset)

	// Request 3: 1 remaining
	allowed, meta, err = store.AllowWithDetails("user1")
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 1, meta.Remaining)

	// Request 4: 0 remaining
	allowed, meta, err = store.AllowWithDetails("user1")
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 0, meta.Remaining)
	assert.Equal(t, 2000*time.Millisecond, meta.Reset)

	// Request 5: limit exceeded
	allowed, meta, err = store.AllowWithDetails("user1")
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, meta.Remaining)
	assert.Equal(t, 500*time.Millisecond, meta.RetryAfter)
	assert.Equal(t, 2000*time.Millisecond, meta.Reset)
}

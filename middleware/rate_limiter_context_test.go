// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
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

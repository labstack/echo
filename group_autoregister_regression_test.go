// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests guard the behavior re-introduced by restoring automatic group
// catch-all (404) route registration (PR #2996). v5 originally removed that
// registration because the implicit catch-all could:
//   1. mask 405 Method Not Allowed as 404, and
//   2. shadow sibling/root routes.
// Restoring auto-registration must NOT bring those regressions back.

// passthroughMW is a no-op middleware that forces a group to register its
// implicit catch-all routes (a group only auto-registers when it has middleware).
func passthroughMW(next HandlerFunc) HandlerFunc {
	return func(c *Context) error { return next(c) }
}

// 1. Wrong method on an existing group route must still return 405 (with Allow
// header), not be swallowed as 404 by the group's catch-all.
func TestGroup_autoCatchAll_wrongMethodStillReturns405(t *testing.T) {
	e := New()
	g := e.Group("/api", passthroughMW)
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
		"POST to a GET-only group route must be 405, not masked to 404 by the catch-all")
	assert.Contains(t, rec.Header().Get(HeaderAllow), http.MethodGet,
		"405 response must advertise allowed methods")
}

// 2. The group's catch-all must not shadow a concrete sibling route under the
// same prefix: a matched route returns its own handler, not the 404 catch-all.
func TestGroup_autoCatchAll_doesNotShadowConcreteSiblingRoute(t *testing.T) {
	e := New()
	g := e.Group("/api", passthroughMW)
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })
	g.GET("/health", func(c *Context) error { return c.String(http.StatusOK, "health") })

	for path, want := range map[string]string{"/api/users": "users", "/api/health": "health"} {
		status, body := request(http.MethodGet, path, e)
		assert.Equal(t, http.StatusOK, status, "concrete route %s must win over the group catch-all", path)
		assert.Equal(t, want, body, "concrete route %s must run its own handler", path)
	}

	// Only a genuinely unmatched path under the prefix hits the catch-all (404).
	status, _ := request(http.MethodGet, "/api/nope", e)
	assert.Equal(t, http.StatusNotFound, status, "unmatched path under the prefix should hit the catch-all 404")
}

// 3. The group's catch-all (prefixed) must not shadow routes outside the group,
// including root-level routes.
func TestGroup_autoCatchAll_doesNotShadowRootRoute(t *testing.T) {
	e := New()
	e.GET("/health", func(c *Context) error { return c.String(http.StatusOK, "root-health") })
	g := e.Group("/api", passthroughMW)
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	status, body := request(http.MethodGet, "/health", e)
	assert.Equal(t, http.StatusOK, status, "root route must be unaffected by a group's catch-all")
	assert.Equal(t, "root-health", body)
}

// 4. Two sibling groups must not shadow each other's routes via their catch-alls.
func TestGroup_autoCatchAll_siblingGroupsDoNotShadow(t *testing.T) {
	e := New()
	v1 := e.Group("/api/v1", passthroughMW)
	v1.GET("/ping", func(c *Context) error { return c.String(http.StatusOK, "v1") })
	v2 := e.Group("/api/v2", passthroughMW)
	v2.GET("/ping", func(c *Context) error { return c.String(http.StatusOK, "v2") })

	for path, want := range map[string]string{"/api/v1/ping": "v1", "/api/v2/ping": "v2"} {
		status, body := request(http.MethodGet, path, e)
		assert.Equal(t, http.StatusOK, status, "%s must resolve to its own group handler", path)
		assert.Equal(t, want, body)
	}
}

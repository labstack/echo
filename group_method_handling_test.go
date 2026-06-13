// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests lock in v5's method-handling semantics for routes registered through
// a Group. v5 resolves method mismatches (405) and OPTIONS at the router level and
// does NOT register any implicit per-group catch-all route.
//
// They double as a regression gate. Registering a group-level catch-all — whether
// manually via g.RouteNotFound("/*", ...) or automatically (as proposed in #2996 to
// fix CORS-on-group preflight) — makes that catch-all match every method, which masks
// both 405 and v5's automatic OPTIONS response as 404 — demonstrated directly by
// TestGroupRoute_catchAllMasksMethodHandling below. If that masking becomes the
// default (e.g. #2996 lands), the first two tests below fail.

// A method mismatch on an existing group route must return 405 with the allowed
// methods, not be masked to 404.
func TestGroupRoute_methodMismatchReturns405(t *testing.T) {
	e := New()
	g := e.Group("/api")
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
		"POST to a GET-only group route must be 405, not masked to 404")
	assert.Equal(t, "OPTIONS, GET", rec.Header().Get(HeaderAllow),
		"405 response must advertise the allowed methods")
}

// OPTIONS on an existing group route is answered automatically by Echo (204 +
// Allow). This is the behavior CORS preflight relies on, so it must not be masked.
func TestGroupRoute_automaticOPTIONS(t *testing.T) {
	e := New()
	g := e.Group("/api")
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	req := httptest.NewRequest(http.MethodOptions, "/api/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code,
		"OPTIONS on a registered group route must be auto-answered (204), not masked to 404")
	assert.Equal(t, "OPTIONS, GET", rec.Header().Get(HeaderAllow),
		"automatic OPTIONS response must advertise the allowed methods")
}

// A matched concrete route resolves to its own handler; only a genuinely unmatched
// path under the prefix is a 404.
func TestGroupRoute_concreteRoutesResolve(t *testing.T) {
	e := New()
	g := e.Group("/api")
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	status, body := request(http.MethodGet, "/api/users", e)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "users", body)

	status, _ = request(http.MethodGet, "/api/nope", e)
	assert.Equal(t, http.StatusNotFound, status)
}

// A group prefix must not affect routing of routes registered outside the group.
func TestGroup_doesNotAffectRootRoutes(t *testing.T) {
	e := New()
	e.GET("/health", func(c *Context) error { return c.String(http.StatusOK, "root") })
	g := e.Group("/api")
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })

	status, body := request(http.MethodGet, "/health", e)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "root", body)
}

// Characterization of the regression the 405/OPTIONS tests above guard against:
// registering a group-wide catch-all (the manual equivalent of #2996's auto-
// registration) makes it match every method, so method mismatches and the automatic
// OPTIONS response are masked as 404 even though the concrete route still resolves.
// If a future change teaches the catch-all to preserve method semantics, update this.
func TestGroupRoute_catchAllMasksMethodHandling(t *testing.T) {
	e := New()
	g := e.Group("/api")
	g.GET("/users", func(c *Context) error { return c.String(http.StatusOK, "users") })
	g.RouteNotFound("/*", func(c *Context) error { return c.NoContent(http.StatusNotFound) })

	// The concrete route still resolves.
	status, body := request(http.MethodGet, "/api/users", e)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "users", body)

	// But the catch-all masks the method mismatch (would be 405) ...
	post := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	postRec := httptest.NewRecorder()
	e.ServeHTTP(postRec, post)
	assert.Equal(t, http.StatusNotFound, postRec.Code,
		"a group-wide catch-all masks the 405 method-mismatch as 404")

	// ... and the automatic OPTIONS response (would be 204).
	opts := httptest.NewRequest(http.MethodOptions, "/api/users", nil)
	optsRec := httptest.NewRecorder()
	e.ServeHTTP(optsRec, opts)
	assert.Equal(t, http.StatusNotFound, optsRec.Code,
		"a group-wide catch-all masks the automatic OPTIONS (204) response as 404")
}

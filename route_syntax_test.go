// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertRouteResponse(t *testing.T, e *Echo, path string, want string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if !assert.NotPanics(t, func() { e.ServeHTTP(rec, req) }) {
		return
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, want, rec.Body.String())
}

func TestRouterInlineVerbRoutes(t *testing.T) {
	for _, order := range [][]string{{"cancel", "get"}, {"get", "cancel"}} {
		e := New()
		for _, verb := range order {
			verb := verb
			e.GET("/r/:name\\:"+verb, func(c *Context) error {
				return c.String(http.StatusOK, verb+":"+c.Param("name"))
			})
		}
		assertRouteResponse(t, e, "/r/foo:cancel", "cancel:foo")
		assertRouteResponse(t, e, "/r/foo:get", "get:foo")
		assertRouteResponse(t, e, "/r/foo:bar:cancel", "cancel:foo:bar")
	}
}

func TestRouterInlineVerbLongestSuffix(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:foo\:bar`, func(c *Context) error {
		return c.String(http.StatusOK, "long:"+c.Param("name"))
	})
	e.GET(`/r/:name\:bar`, func(c *Context) error {
		return c.String(http.StatusOK, "short:"+c.Param("name"))
	})
	assertRouteResponse(t, e, "/r/a:foo:bar", "long:a")
	assertRouteResponse(t, e, "/r/a:bar", "short:a")
}

func TestRouterInlineVerbWithFollowingParam(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel/:action`, func(c *Context) error {
		return c.String(http.StatusOK, c.Param("name")+":"+c.Param("action"))
	})
	assertRouteResponse(t, e, "/r/foo:cancel/bar", "foo:bar")
}

func TestRouterInlineVerbWithWildcard(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel/*`, func(c *Context) error {
		return c.String(http.StatusOK, c.Param("name")+":"+c.Param("*"))
	})
	assertRouteResponse(t, e, "/r/foo:cancel/bar", "foo:bar")
	assertRouteResponse(t, e, "/r/foo:cancel/", "foo:")
}

func TestRouterInlineVerbAndGenericParam(t *testing.T) {
	e := New()
	e.GET("/r/:name", func(c *Context) error {
		return c.String(http.StatusOK, "generic:"+c.Param("name"))
	})
	e.GET(`/r/:name\:cancel`, func(c *Context) error {
		return c.String(http.StatusOK, "cancel:"+c.Param("name"))
	})
	assertRouteResponse(t, e, "/r/foo:cancel", "cancel:foo")
	assertRouteResponse(t, e, "/r/foo:other", "generic:foo:other")
}

func TestRouterRemoveEscapedColonAndReadd(t *testing.T) {
	e := New()
	static := func(c *Context) error { return c.String(http.StatusOK, "static") }
	e.GET(`/a\:b`, static)
	e.GET("/a:id", func(c *Context) error {
		return c.String(http.StatusOK, "param:"+c.Param("id"))
	})
	assertRouteResponse(t, e, "/a:b", "static")
	assert.NoError(t, e.Router().Remove(http.MethodGet, `/a\:b`))
	assertRouteResponse(t, e, "/a:b", "param::b")
	e.GET(`/a\:b`, static)
	assertRouteResponse(t, e, "/a:b", "static")
}

func TestRouterRemoveInlineVerbAndReadd(t *testing.T) {
	e := New()
	cancel := func(c *Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) }
	e.GET(`/r/:name\:cancel`, cancel)
	e.GET(`/r/:name\:get`, func(c *Context) error {
		return c.String(http.StatusOK, "get:"+c.Param("name"))
	})
	assert.NoError(t, e.Router().Remove(http.MethodGet, `/r/:name\:cancel`))
	assertRouteResponse(t, e, "/r/foo:get", "get:foo")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/foo:cancel", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	e.GET(`/r/:name\:cancel`, cancel)
	assertRouteResponse(t, e, "/r/foo:cancel", "cancel:foo")
}

func TestRouteInfoReverseInlineVerb(t *testing.T) {
	ri := RouteInfo{Path: `/r/:name\:cancel`}
	assert.Equal(t, "/r/foo:cancel", ri.Reverse("foo"))
	assert.Equal(t, "/r/:name:cancel", ri.Reverse())
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRouterInlineVerbBacktracksToGenericRoute(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:ab/:p/z`, func(c *Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:id/info`, func(c *Context) error { return c.String(http.StatusOK, c.Param("id")) })
	assertRouteResponse(t, e, "/r/q:abc/info", "q:abc")
	assertRouteResponse(t, e, "/r/q:ab/info", "q:ab")

	e = New()
	e.GET(`/r/:name\:y/:p/z`, func(c *Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:id/info`, func(c *Context) error { return c.String(http.StatusOK, c.Param("id")) })
	assertRouteResponse(t, e, "/r/q:y/info", "q:y")
}

func TestRouterInlineVerbMethodFallback(t *testing.T) {
	e := New()
	e.GET(`/r/:id`, func(c *Context) error { return c.String(http.StatusOK, c.Param("id")) })
	e.POST(`/r/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, "verb") })
	assertRouteResponse(t, e, "/r/foo:cancel", "foo:cancel")
}

func TestRouterInlineVerbRequiresNonemptyParameter(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, c.Param("name")) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/:cancel", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouterInlineVerbAndStaticSibling(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:x/:id`, func(c *Context) error { return c.String(http.StatusOK, "verb:"+c.Param("name")+":"+c.Param("id")) })
	e.GET(`/r/:name/q`, func(c *Context) error { return c.String(http.StatusOK, "static:"+c.Param("name")) })
	assertRouteResponse(t, e, "/r/a:x/q", "verb:a:q")
	assertRouteResponse(t, e, "/r/a:y/q", "static:a:y")
}

func TestRouterInlineVerbMustEndPathSegment(t *testing.T) {
	// An escaped colon that is followed by a param or wildcard in the same segment keeps its older meaning: it is part
	// of the param name. Trying every colon in a request segment for such routes could not be bounded.
	e := New()
	ri, err := e.AddRoute(Route{Method: http.MethodGet, Path: `/r/:name\:x:id`, Handler: func(c *Context) error { return nil }})
	assert.NoError(t, err)
	assert.Equal(t, []string{`name\:x:id`}, ri.Parameters)
	ri, err = e.AddRoute(Route{Method: http.MethodGet, Path: `/s/:name\:x*`, Handler: func(c *Context) error { return nil }})
	assert.NoError(t, err)
	assert.Equal(t, []string{`name\:x*`}, ri.Parameters)
}

func TestRouterInlineVerbLeafParamAfterVerb(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:x/:rest`, func(c *Context) error { return c.String(http.StatusOK, c.Param("name")+"|"+c.Param("rest")) })
	assertRouteResponse(t, e, "/r/a:x/b/c", "a|b/c")
}

func TestRouterInlineVerbWithGroupMiddlewareAndCatchAll(t *testing.T) {
	e := New()
	g := e.Group("/r", func(next HandlerFunc) HandlerFunc { return next })
	g.GET("/:name", func(c *Context) error { return c.String(http.StatusOK, "generic:"+c.Param("name")) })
	g.GET(`/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	assertRouteResponse(t, e, "/r/foo:other", "generic:foo:other")
	assertRouteResponse(t, e, "/r/foo:cancel", "cancel:foo")

	e = New()
	e.GET("/r/:name", func(c *Context) error { return c.String(http.StatusOK, "generic:"+c.Param("name")) })
	e.GET(`/r/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	e.GET("/*", func(c *Context) error { return c.String(http.StatusOK, "any") })
	assertRouteResponse(t, e, "/r/foo:other", "generic:foo:other")
	assertRouteResponse(t, e, "/r/foo:cancel", "cancel:foo")
}

func TestRouterInlineVerbManyColons(t *testing.T) {
	// Every colon in the segment is a possible split. Each is tried at most once, so a long run of colons is routed
	// in linear time.
	e := New()
	e.GET(`/r/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	e.GET(`/r/:name\:c`, func(c *Context) error { return c.String(http.StatusOK, "c:"+c.Param("name")) })
	e.GET(`/r/:name\:x/:a\:y/z`, func(c *Context) error { return c.String(http.StatusOK, "nested") })
	colons := strings.Repeat(":", 1<<16)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a"+colons+"b", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a"+colons+"x/b"+colons+"y/nope", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// every ":c" enters the shared ":c" verb node before failing, so each split is retried
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a"+strings.Repeat(":c", 1<<15)+"b", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)

	assertRouteResponse(t, e, "/r/a"+colons+"cancel", "cancel:a"+colons[1:])
}

func TestRouterInlineVerbMethodNotAllowedWithoutFallback(t *testing.T) {
	e := New()
	e.POST(`/r/:name\:cancel`, func(c *Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/foo:cancel", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

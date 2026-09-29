// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	// the first escaped colon decides, so a later one that is followed only by static text does not start a verb
	ri, err = e.AddRoute(Route{Method: http.MethodGet, Path: `/t/:a\:x:y\:z`, Handler: func(c *Context) error { return nil }})
	assert.NoError(t, err)
	assert.Equal(t, []string{`a\:x:y\:z`}, ri.Parameters)
	assert.Equal(t, "/t/:a:x:y:z", ri.Reverse())
}

func TestRouterInlineVerbBeforeWholeSegment(t *testing.T) {
	// a matching inline verb split is tried before the whole segment, also when a wildcard follows the verb
	e := New()
	e.GET(`/r/:name\:x/*`, func(c *Context) error { return c.String(http.StatusOK, "verb:"+c.Param("name")+"|"+c.Param("*")) })
	e.GET(`/r/:id/info`, func(c *Context) error { return c.String(http.StatusOK, "info:"+c.Param("id")) })
	assertRouteResponse(t, e, "/r/a:x/info", "verb:a|info")
	assertRouteResponse(t, e, "/r/a:y/info", "info:a:y")
}

func TestRouterInlineVerbRetriedAfterWildcard(t *testing.T) {
	// a wildcard ends the search, but the split above it is still retried with the next split and the whole segment
	e := New()
	e.POST(`/r/:n\:v/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n/*`, func(c *Context) error { return c.String(http.StatusOK, "get:"+c.Param("n")+"|"+c.Param("*")) })
	assertRouteResponse(t, e, "/r/a:v/q", "get:a:v|q")

	e = New()
	e.POST(`/r/:n\:a\:b/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n\:b/x`, func(c *Context) error { return c.String(http.StatusOK, "get:"+c.Param("n")) })
	assertRouteResponse(t, e, "/r/q:a:b/x", "get:q:a")

	e = New()
	e.POST(`/r/:n\:v/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a:v/q", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRouterRemoveInlineVerbsSharingPrefix(t *testing.T) {
	e := New()
	h := func(c *Context) error { return c.String(http.StatusOK, c.Path()+"|"+c.Param("n")) }
	e.GET(`/r/:n\:cancel`, h)
	e.GET(`/r/:n\:close`, h)
	e.GET(`/r/:n/x`, h)
	assert.NoError(t, e.Router().Remove(http.MethodGet, `/r/:n\:cancel`))
	assertRouteResponse(t, e, "/r/a:close", `/r/:n\:close|a`)
	assertRouteResponse(t, e, "/r/a:b/x", "/r/:n/x|a:b")
	assert.NoError(t, e.Router().Remove(http.MethodGet, `/r/:n\:close`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a:close", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	e.GET(`/r/:n\:close`, h)
	assertRouteResponse(t, e, "/r/a:close", `/r/:n\:close|a`)
	assertRouteResponse(t, e, "/r/a:b/x", "/r/:n/x|a:b")
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
	start := time.Now()
	defer func() {
		// linear routing takes milliseconds here; trying splits quadratically would take minutes
		assert.Less(t, time.Since(start), 10*time.Second)
	}()

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

func TestRouterInlineVerbWildcardBacktracksBelowSplit(t *testing.T) {
	// after a wildcard below a split fails, the other routes below that split are tried before the next split
	e := New()
	e.POST(`/r/:n\:v/a/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n\:v/:p/b`, func(c *Context) error { return c.String(http.StatusOK, "verb:"+c.Param("n")+"|"+c.Param("p")) })
	e.GET(`/r/:n/a/b`, func(c *Context) error { return c.String(http.StatusOK, "generic:"+c.Param("n")) })
	assertRouteResponse(t, e, "/r/q:v/a/b", "verb:q|a")

	// nested splits: the nearest pending split is retried first, then the outer one
	e = New()
	e.POST(`/r/:a\:x/:b\:y/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:a/:b\:y/*`, func(c *Context) error {
		return c.String(http.StatusOK, c.Param("a")+"|"+c.Param("b")+"|"+c.Param("*"))
	})
	assertRouteResponse(t, e, "/r/p:x/q:y/z", "p:x|q|z")

	e = New()
	e.POST(`/r/:a\:x/:b\:y/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:a\:x/:b/*`, func(c *Context) error {
		return c.String(http.StatusOK, c.Param("a")+"|"+c.Param("b")+"|"+c.Param("*"))
	})
	assertRouteResponse(t, e, "/r/p:x/q:y/z", "p|q:y|z")

	// a RouteNotFound wildcard below a split handles the request like any other RouteNotFound route
	e = New()
	e.RouteNotFound(`/r/:a\:x/*`, func(c *Context) error { return c.String(http.StatusOK, "not found:"+c.Param("a")) })
	e.GET(`/r/:a/k`, func(c *Context) error { return c.String(http.StatusOK, "k") })
	assertRouteResponse(t, e, "/r/p:x/k", "not found:p")
}

func TestRouterRemoveRouteSharingNode(t *testing.T) {
	e := New()
	e.GET("/u/:id", func(c *Context) error { return c.String(http.StatusOK, "get:"+c.Param("id")) })
	e.POST("/u/:uid", func(c *Context) error { return c.String(http.StatusOK, "post:"+c.Param("uid")) })
	assert.Error(t, e.Router().Remove(http.MethodGet, "/u/:uid"))
	assertRouteResponse(t, e, "/u/1", "get:1")
	assert.NoError(t, e.Router().Remove(http.MethodGet, "/u/:id"))
	assert.Len(t, e.Router().Routes(), 1)

	e.GET("x", func(c *Context) error { return c.NoContent(http.StatusOK) })
	assert.NoError(t, e.Router().Remove(http.MethodGet, "x"))
	assert.Len(t, e.Router().Routes(), 1)
}

func TestRouterInlineVerbMisc(t *testing.T) {
	e := New()
	e.POST(`/r/:n\:v/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.RouteNotFound(`/r/:n/*`, func(c *Context) error { return c.String(http.StatusOK, "not found:"+c.Param("n")) })
	// the whole segment reaches the RouteNotFound route, as a static sibling would
	assertRouteResponse(t, e, "/r/a:v/q", "not found:a:v")

	// with routing on the escaped path, an encoded colon is not a verb delimiter
	e = New()
	e.GET(`/r/:name\:cancel`, func(c *Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:name`, func(c *Context) error { return c.String(http.StatusOK, "generic:"+c.Param("name")) })
	assertRouteResponse(t, e, "/r/foo%3Acancel", "generic:foo%3Acancel")

	ri := RouteInfo{Path: `/r/:n\:v/*`}
	assert.Equal(t, "/r/:n:v/*", ri.Reverse())
	assert.Equal(t, "/r/a:v/b/c", ri.Reverse("a", "b/c"))

	// a param name with ':' keeps an escaped colon as part of the name
	ri, err := e.AddRoute(Route{Method: http.MethodGet, Path: `/s/:a:b\:v`, Handler: func(c *Context) error { return nil }})
	assert.NoError(t, err)
	assert.Equal(t, []string{`a:b\:v`}, ri.Parameters)
}

func TestRouterInlineVerbKeepsLeafParam(t *testing.T) {
	// a param with only an inline verb child still takes the rest of the path when no split matches
	e := New()
	e.GET("/files/:path", func(c *Context) error { return c.String(http.StatusOK, "get:"+c.Param("path")) })
	e.POST(`/files/:name\:upload`, func(c *Context) error { return c.String(http.StatusOK, "upload:"+c.Param("name")) })
	assertRouteResponse(t, e, "/files/a/b", "get:a/b")
	assertRouteResponse(t, e, "/files/a:upload/b", "get:a:upload/b")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/files/a:upload", nil))
	assert.Equal(t, "upload:a", rec.Body.String())

	// with another child the param stops at the slash, as before
	e.GET("/files/:path/meta", func(c *Context) error { return c.String(http.StatusOK, "meta:"+c.Param("path")) })
	assertRouteResponse(t, e, "/files/a/meta", "meta:a")
}

func TestRouterInlineVerbPendingAboveParam(t *testing.T) {
	// the pending split is found above a param without a split
	e := New()
	e.POST(`/r/:a\:v/:b/*`, func(c *Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:a/:b/q`, func(c *Context) error { return c.String(http.StatusOK, "get:"+c.Param("a")+"|"+c.Param("b")) })
	assertRouteResponse(t, e, "/r/x:v/y/q", "get:x:v|y")
}

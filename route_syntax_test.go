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

func assertInlineVerbResponse(t *testing.T, e *Echo, path, want string) {
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
			e.GET("/r/:name\\:"+verb, func(c Context) error {
				return c.String(http.StatusOK, verb+":"+c.Param("name"))
			})
		}
		assertInlineVerbResponse(t, e, "/r/foo:cancel", "cancel:foo")
		assertInlineVerbResponse(t, e, "/r/foo:get", "get:foo")
		assertInlineVerbResponse(t, e, "/r/foo:bar:cancel", "cancel:foo:bar")
	}
}

func TestRouterInlineVerbLongestSuffix(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:foo\:bar`, func(c Context) error {
		return c.String(http.StatusOK, "long:"+c.Param("name"))
	})
	e.GET(`/r/:name\:bar`, func(c Context) error {
		return c.String(http.StatusOK, "short:"+c.Param("name"))
	})
	assertInlineVerbResponse(t, e, "/r/a:foo:bar", "long:a")
	assertInlineVerbResponse(t, e, "/r/a:bar", "short:a")
}

func TestRouterInlineVerbWithFollowingParam(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel/:action`, func(c Context) error {
		return c.String(http.StatusOK, c.Param("name")+":"+c.Param("action"))
	})
	assertInlineVerbResponse(t, e, "/r/foo:cancel/bar", "foo:bar")
}

func TestRouterInlineVerbWithWildcard(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel/*`, func(c Context) error {
		return c.String(http.StatusOK, c.Param("name")+":"+c.Param("*"))
	})
	assertInlineVerbResponse(t, e, "/r/foo:cancel/bar", "foo:bar")
	assertInlineVerbResponse(t, e, "/r/foo:cancel/", "foo:")
}

func TestRouterInlineVerbAndGenericParam(t *testing.T) {
	e := New()
	e.GET("/r/:name", func(c Context) error {
		return c.String(http.StatusOK, "generic:"+c.Param("name"))
	})
	e.GET(`/r/:name\:cancel`, func(c Context) error {
		return c.String(http.StatusOK, "cancel:"+c.Param("name"))
	})
	assertInlineVerbResponse(t, e, "/r/foo:cancel", "cancel:foo")
	assertInlineVerbResponse(t, e, "/r/foo:other", "generic:foo:other")
}

func TestRouterReverseInlineVerb(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error { return nil }).Name = "inline-verb"
	assert.Equal(t, "/r/foo:cancel", e.Reverse("inline-verb", "foo"))
	assert.Equal(t, "/r/:name:cancel", e.Reverse("inline-verb"))
}

func TestRouterInlineVerbBacktracksToGenericRoute(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:v/:id/end`, func(c Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:name/other`, func(c Context) error { return c.String(http.StatusOK, c.Param("name")) })
	assertInlineVerbResponse(t, e, "/r/a:vq/other", "a:vq")
	assertInlineVerbResponse(t, e, "/r/a:v/other", "a:v")
	assertInlineVerbResponse(t, e, "/r/a:v/q/end", "verb")
}

func TestRouterInlineVerbMethodFallback(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error { return c.String(http.StatusOK, "verb") })
	e.POST(`/r/:name`, func(c Context) error { return c.String(http.StatusOK, c.Param("name")) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/r/foo:cancel", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "foo:cancel", rec.Body.String())
}

func TestRouterInlineVerbRequiresNonemptyParameter(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error { return c.String(http.StatusOK, c.Param("name")) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/:cancel", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouterInlineVerbAndStaticSibling(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:x/:id`, func(c Context) error { return c.String(http.StatusOK, "verb:"+c.Param("name")+":"+c.Param("id")) })
	e.GET(`/r/:name/q`, func(c Context) error { return c.String(http.StatusOK, "static:"+c.Param("name")) })
	assertInlineVerbResponse(t, e, "/r/a:x/q", "verb:a:q")
	assertInlineVerbResponse(t, e, "/r/a:y/q", "static:a:y")
}

func TestRouterInlineVerbMustEndPathSegment(t *testing.T) {
	// An escaped colon that is followed by a param or wildcard in the same segment keeps its older meaning: it is part
	// of the param name. Trying every colon in a request segment for such routes could not be bounded.
	e := New()
	e.GET(`/r/:name\:x:id`, func(c Context) error { return c.String(http.StatusOK, strings.Join(c.ParamNames(), ",")) })
	e.GET(`/s/:name\:x*`, func(c Context) error { return c.String(http.StatusOK, strings.Join(c.ParamNames(), ",")) })
	assertInlineVerbResponse(t, e, "/r/foo", `name\:x:id`)
	assertInlineVerbResponse(t, e, "/s/foo", `name\:x*`)
	// the first escaped colon decides, so a later one that is followed only by static text does not start a verb
	e.GET(`/t/:a\:x:y\:z`, func(c Context) error { return c.String(http.StatusOK, strings.Join(c.ParamNames(), ",")) }).Name = "legacy"
	assertInlineVerbResponse(t, e, "/t/q:z", `a\:x:y\:z`)
	assert.Equal(t, "/t/:a:x:y:z", e.Reverse("legacy"))
}

func TestRouterInlineVerbBeforeWholeSegment(t *testing.T) {
	// a matching inline verb split is tried before the whole segment, also when a wildcard follows the verb
	e := New()
	e.GET(`/r/:name\:x/*`, func(c Context) error { return c.String(http.StatusOK, "verb:"+c.Param("name")+"|"+c.Param("*")) })
	e.GET(`/r/:id/info`, func(c Context) error { return c.String(http.StatusOK, "info:"+c.Param("id")) })
	assertInlineVerbResponse(t, e, "/r/a:x/info", "verb:a|info")
	assertInlineVerbResponse(t, e, "/r/a:y/info", "info:a:y")
}

func TestRouterInlineVerbLeafParamAfterVerb(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:x/:rest`, func(c Context) error { return c.String(http.StatusOK, c.Param("name")+"|"+c.Param("rest")) })
	assertInlineVerbResponse(t, e, "/r/a:x/b/c", "a|b/c")
}

func TestRouterInlineVerbWithGroupMiddlewareAndCatchAll(t *testing.T) {
	e := New()
	g := e.Group("/r", func(next HandlerFunc) HandlerFunc { return next })
	g.GET("/:name", func(c Context) error { return c.String(http.StatusOK, "generic:"+c.Param("name")) })
	g.GET(`/:name\:cancel`, func(c Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	assertInlineVerbResponse(t, e, "/r/foo:other", "generic:foo:other")
	assertInlineVerbResponse(t, e, "/r/foo:cancel", "cancel:foo")

	e = New()
	e.GET("/r/:name", func(c Context) error { return c.String(http.StatusOK, "generic:"+c.Param("name")) })
	e.GET(`/r/:name\:cancel`, func(c Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	e.GET("/*", func(c Context) error { return c.String(http.StatusOK, "any") })
	assertInlineVerbResponse(t, e, "/r/foo:other", "generic:foo:other")
	assertInlineVerbResponse(t, e, "/r/foo:cancel", "cancel:foo")
}

func TestRouterInlineVerbManyColons(t *testing.T) {
	// Every colon in the segment is a possible split. Each is tried at most once, so a long run of colons is routed
	// in linear time.
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error { return c.String(http.StatusOK, "cancel:"+c.Param("name")) })
	e.GET(`/r/:name\:c`, func(c Context) error { return c.String(http.StatusOK, "c:"+c.Param("name")) })
	e.GET(`/r/:name\:x/:a\:y/z`, func(c Context) error { return c.String(http.StatusOK, "nested") })
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

	assertInlineVerbResponse(t, e, "/r/a"+colons+"cancel", "cancel:a"+colons[1:])
}

func TestRouterStaticParamNamesRemainEmptySlice(t *testing.T) {
	e := New()
	e.GET("/static", func(c Context) error {
		assert.NotNil(t, c.ParamNames())
		assert.Empty(t, c.ParamNames())
		return c.NoContent(http.StatusOK)
	})
	assertInlineVerbResponse(t, e, "/static", "")
}

func TestRouterInlineVerbEncodedColonUsesGenericRoute(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error {
		return c.String(http.StatusOK, "verb")
	})
	e.GET(`/r/:name`, func(c Context) error {
		return c.String(http.StatusOK, "generic:"+c.Param("name"))
	})
	assertInlineVerbResponse(t, e, "/r/foo%3Acancel", "generic:foo%3Acancel")
}

func TestRouterInlineVerbMethodNotAllowedWithoutFallback(t *testing.T) {
	e := New()
	e.POST(`/r/:name\:cancel`, func(c Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/foo:cancel", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRouterInlineVerbRetriedAfterWildcard(t *testing.T) {
	// a wildcard ends the search, but the split above it is still retried with the next split and the whole segment
	e := New()
	e.POST(`/r/:n\:v/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n/*`, func(c Context) error { return c.String(http.StatusOK, "get:"+c.Param("n")+"|"+c.Param("*")) })
	assertInlineVerbResponse(t, e, "/r/a:v/q", "get:a:v|q")

	e = New()
	e.POST(`/r/:n\:a\:b/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n\:b/x`, func(c Context) error { return c.String(http.StatusOK, "get:"+c.Param("n")) })
	assertInlineVerbResponse(t, e, "/r/q:a:b/x", "get:q:a")

	e = New()
	e.POST(`/r/:n\:v/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/a:v/q", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRouterInlineVerbChangesEscapedColonAfterParam(t *testing.T) {
	// Before inline verbs, `/:name\:cancel` was a single param named `name\:cancel` that matched any segment.
	e := New()
	e.GET(`/r/:name\:cancel`, func(c Context) error {
		return c.String(http.StatusOK, strings.Join(c.ParamNames(), ",")+"="+c.Param("name"))
	})
	assertInlineVerbResponse(t, e, "/r/foo:cancel", "name=foo")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/r/foo", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouterReverseEscapedColonPlaceholder(t *testing.T) {
	e := New()
	e.GET(`/r/:n\:x:id`, func(c Context) error { return nil }).Name = "legacy"
	e.GET(`/r/:name\:cancel`, func(c Context) error { return nil }).Name = "verb"
	assert.Equal(t, "/r/:n:x:id", e.Reverse("legacy"))
	assert.Equal(t, "/r/foo", e.Reverse("legacy", "foo"))
	assert.Equal(t, "/r/:name:cancel", e.Reverse("verb"))
	assert.Equal(t, "/r/foo:cancel", e.Reverse("verb", "foo"))
}

func TestRouterInlineVerbWildcardBacktracksBelowSplit(t *testing.T) {
	// after a wildcard below a split fails, the other routes below that split are tried before the next split
	e := New()
	e.POST(`/r/:n\:v/a/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:n\:v/:p/b`, func(c Context) error { return c.String(http.StatusOK, "verb:"+c.Param("n")+"|"+c.Param("p")) })
	e.GET(`/r/:n/a/b`, func(c Context) error { return c.String(http.StatusOK, "generic:"+c.Param("n")) })
	assertInlineVerbResponse(t, e, "/r/q:v/a/b", "verb:q|a")

	// nested splits: the nearest pending split is retried first, then the outer one
	e = New()
	e.POST(`/r/:a\:x/:b\:y/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:a/:b\:y/*`, func(c Context) error {
		return c.String(http.StatusOK, c.Param("a")+"|"+c.Param("b")+"|"+c.Param("*"))
	})
	assertInlineVerbResponse(t, e, "/r/p:x/q:y/z", "p:x|q|z")

	e = New()
	e.POST(`/r/:a\:x/:b\:y/*`, func(c Context) error { return c.String(http.StatusOK, "post") })
	e.GET(`/r/:a\:x/:b/*`, func(c Context) error {
		return c.String(http.StatusOK, c.Param("a")+"|"+c.Param("b")+"|"+c.Param("*"))
	})
	assertInlineVerbResponse(t, e, "/r/p:x/q:y/z", "p|q:y|z")

	// a RouteNotFound wildcard below a split handles the request like any other RouteNotFound route
	e = New()
	e.RouteNotFound(`/r/:a\:x/*`, func(c Context) error { return c.String(http.StatusOK, "not found:"+c.Param("a")) })
	e.GET(`/r/:a/k`, func(c Context) error { return c.String(http.StatusOK, "k") })
	assertInlineVerbResponse(t, e, "/r/p:x/k", "not found:p")
}

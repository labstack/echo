// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	e.GET(`/r/:name\:v:id/end`, func(c Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:name/other`, func(c Context) error { return c.String(http.StatusOK, c.Param("name")) })
	assertInlineVerbResponse(t, e, "/r/a:vq/other", "a:vq")
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

func TestRouterInlineVerbKeepsStaticSiblingPriority(t *testing.T) {
	e := New()
	e.GET(`/r/:name\:x:id`, func(c Context) error { return c.String(http.StatusOK, "verb") })
	e.GET(`/r/:name/q`, func(c Context) error { return c.String(http.StatusOK, "static:"+c.Param("name")) })
	assertInlineVerbResponse(t, e, "/r/a:x/q", "static:a:x")
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

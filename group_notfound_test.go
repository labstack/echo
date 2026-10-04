// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func groupNotFoundTraceMiddleware(name string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			c.Response().Header().Add("X-Trace", name)
			return next(c)
		}
	}
}

func TestGroupUsePreservesNotFoundHandler(t *testing.T) {
	for _, path := range []string{"", "/*"} {
		for _, order := range []string{"before middleware", "between middleware", "after middleware"} {
			t.Run(path+"/"+order, func(t *testing.T) {
				e := New()
				g := e.Group("/api")
				register := func() {
					_, err := g.AddRoute(Route{
						Method: RouteNotFound,
						Path:   path,
						Name:   "custom-not-found",
						Handler: func(c *Context) error {
							c.Response().Header().Add("X-Trace", "handler")
							return c.String(http.StatusNotFound, "custom group 404")
						},
						Middlewares: []MiddlewareFunc{groupNotFoundTraceMiddleware("route")},
					})
					require.NoError(t, err)
				}
				if order == "before middleware" {
					register()
				}
				g.Use(groupNotFoundTraceMiddleware("first"))
				if order == "between middleware" {
					register()
				}
				g.Use(groupNotFoundTraceMiddleware("second"))
				if order == "after middleware" {
					register()
				}
				url := "/api"
				if path == "/*" {
					url += "/missing"
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
				assert.Equal(t, http.StatusNotFound, rec.Code)
				assert.Equal(t, "custom group 404", rec.Body.String())
				assert.Equal(t, []string{"first", "second", "route", "handler"}, rec.Header().Values("X-Trace"))
				ri, err := e.Router().Routes().FindByMethodPath(RouteNotFound, "/api"+path)
				require.NoError(t, err)
				assert.Equal(t, "custom-not-found", ri.Name)
			})
		}
	}
}

func TestGroupUsePreservesNotFoundWithoutRouterOverwrite(t *testing.T) {
	e := NewWithConfig(Config{Router: NewRouter(RouterConfig{AllowOverwritingRoute: false})})
	g := e.Group("/api")
	g.RouteNotFound("/*", func(c *Context) error {
		return c.String(http.StatusNotFound, "original")
	})
	_, err := g.AddRoute(Route{
		Method:  RouteNotFound,
		Path:    "/*",
		Handler: func(c *Context) error { return c.String(http.StatusNotFound, "rejected") },
	})
	require.Error(t, err)
	g.Use(groupNotFoundTraceMiddleware("first"))
	g.Use(groupNotFoundTraceMiddleware("second"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "original", rec.Body.String())
	assert.Equal(t, []string{"first", "second"}, rec.Header().Values("X-Trace"))
}

func TestGroupUsePreservesNotFoundMiddlewareSnapshot(t *testing.T) {
	e := New()
	g := e.Group("/api")
	middlewares := []MiddlewareFunc{groupNotFoundTraceMiddleware("original")}
	_, err := g.AddRoute(Route{
		Method:      RouteNotFound,
		Path:        "/*",
		Handler:     func(c *Context) error { return c.String(http.StatusNotFound, "custom") },
		Middlewares: middlewares,
	})
	require.NoError(t, err)
	middlewares[0] = groupNotFoundTraceMiddleware("mutated")
	g.Use(groupNotFoundTraceMiddleware("group"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	assert.Equal(t, "custom", rec.Body.String())
	assert.Equal(t, []string{"group", "original"}, rec.Header().Values("X-Trace"))
}

func TestGroupUseRejectsNotFoundRegistrationWithoutReplacingHandler(t *testing.T) {
	e := New()
	g := e.Group("/api")
	g.RouteNotFound("/*", func(c *Context) error { return c.String(http.StatusNotFound, "original") })
	rejected := errors.New("registration rejected")
	e.OnAddRoute = func(route Route) error {
		if route.Name == "rejected" {
			return rejected
		}
		return nil
	}
	_, err := g.AddRoute(Route{
		Method:  RouteNotFound,
		Path:    "/*",
		Name:    "rejected",
		Handler: func(c *Context) error { return c.String(http.StatusNotFound, "rejected") },
	})
	require.ErrorIs(t, err, rejected)
	g.Use(groupNotFoundTraceMiddleware("group"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	assert.Equal(t, "original", rec.Body.String())
	assert.Equal(t, []string{"group"}, rec.Header().Values("X-Trace"))
}

func TestGroupUseNotFoundWithAutoRegistrationDisabled(t *testing.T) {
	e := NewWithConfig(Config{NoGroupAutoRegister404Routes: true})
	g := e.Group("/api")
	g.RouteNotFound("/*", func(c *Context) error { return c.String(http.StatusNotFound, "custom") },
		groupNotFoundTraceMiddleware("route"))
	g.Use(groupNotFoundTraceMiddleware("group"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	assert.Equal(t, "custom", rec.Body.String())
	assert.Equal(t, []string{"route"}, rec.Header().Values("X-Trace"))
	_, err := e.Router().Routes().FindByMethodPath(RouteNotFound, "/api")
	assert.Error(t, err)
}

func TestGroupUseDoesNotRestoreRemovedNotFoundHandler(t *testing.T) {
	e := NewWithConfig(Config{Router: NewRouter(RouterConfig{
		NotFoundHandler: func(c *Context) error { return c.String(http.StatusNotFound, "default") },
	})})
	g := e.Group("/api")
	g.RouteNotFound("/*", func(c *Context) error { return c.String(http.StatusNotFound, "removed") })
	g.GET("/*", func(c *Context) error { return c.NoContent(http.StatusOK) })
	require.NoError(t, e.Router().Remove(RouteNotFound, "/api/*"))
	require.NoError(t, e.Router().Remove(http.MethodGet, "/api/*"))
	g.Use(groupNotFoundTraceMiddleware("first"))
	g.Use(groupNotFoundTraceMiddleware("second"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	assert.Equal(t, "default", rec.Body.String())
	assert.Equal(t, []string{"first", "second"}, rec.Header().Values("X-Trace"))
}

func TestGroupUseNotFoundOverHTTP(t *testing.T) {
	e := NewWithConfig(Config{Router: NewRouter(RouterConfig{
		NotFoundHandler: func(c *Context) error { return c.String(http.StatusNotFound, "default") },
	})})
	g := e.Group("/api")
	g.RouteNotFound("/*", func(c *Context) error { return c.String(http.StatusNotFound, "custom") })
	g.Use(groupNotFoundTraceMiddleware("first"))
	g.Use(groupNotFoundTraceMiddleware("second"))
	server := httptest.NewServer(e)
	t.Cleanup(server.Close)
	client := server.Client()
	t.Cleanup(client.CloseIdleConnections)
	for _, tc := range []struct {
		path string
		body string
	}{
		{path: "/api/missing", body: "custom"},
		{path: "/api", body: "default"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response, err := client.Get(server.URL + tc.path)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
			assert.Equal(t, tc.body, string(body))
			assert.Equal(t, []string{"first", "second"}, response.Header.Values("X-Trace"))
		})
	}
}

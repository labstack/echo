package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	// Setup
	t1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "target 1")
	}))
	defer t1.Close()
	url1, _ := url.Parse(t1.URL)
	t2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "target 2")
	}))
	defer t2.Close()
	url2, _ := url.Parse(t2.URL)

	targets := []*ProxyTarget{
		{
			Name: "target 1",
			URL:  url1,
		},
		{
			Name: "target 2",
			URL:  url2,
		},
	}
	rb := NewRandomBalancer(nil)
	// must add targets:
	for _, target := range targets {
		assert.True(t, rb.AddTarget(target))
	}

	// must ignore duplicates:
	for _, target := range targets {
		assert.False(t, rb.AddTarget(target))
	}

	// Random
	e := echo.New()
	e.Use(Proxy(rb))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	body := rec.Body.String()
	expected := map[string]bool{
		"target 1": true,
		"target 2": true,
	}
	assert.Condition(t, func() bool {
		return expected[body]
	})

	for _, target := range targets {
		assert.True(t, rb.RemoveTarget(target.Name))
	}

	assert.False(t, rb.RemoveTarget("unknown target"))

	// Round-robin
	rrb := NewRoundRobinBalancer(targets)
	e = echo.New()
	e.Use(Proxy(rrb))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, "target 1", body)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, "target 2", body)

	// Rewrite
	e = echo.New()
	e.Use(ProxyWithConfig(ProxyConfig{
		Balancer: rrb,
		Rewrite: map[string]string{
			"/old":              "/new",
			"/api/*":            "/$1",
			"/js/*":             "/public/javascripts/$1",
			"/users/*/orders/*": "/user/$1/order/$2",
		},
	}))
	req.URL.Path = "/api/users"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/users", req.URL.Path)
	req.URL.Path = "/js/main.js"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/public/javascripts/main.js", req.URL.Path)
	req.URL.Path = "/old"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.Path)
	req.URL.Path = "/users/jack/orders/1"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jack/order/1", req.URL.Path)
	assert.Equal(t, http.StatusOK, rec.Code)

	// ProxyTarget is set in context
	contextObserver := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			next(c)
			assert.Contains(t, targets, c.Get("target"), "target is not set in context")
			return nil
		}
	}
	rrb1 := NewRoundRobinBalancer(targets)
	e = echo.New()
	e.Use(contextObserver)
	e.Use(Proxy(rrb1))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
}

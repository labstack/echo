package middleware

import (
	"fmt"
	"net"
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
  req.URL.Path = "/users/jill/orders/T%2FcO4lW%2Ft%2FVp%2F"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F", req.URL.Path)
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

func TestProxyRealIPHeader(t *testing.T) {
	// Setup
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer upstream.Close()
	url, _ := url.Parse(upstream.URL)
	rrb := NewRoundRobinBalancer([]*ProxyTarget{{Name: "upstream", URL: url}})
	e := echo.New()
	e.Use(Proxy(rrb))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	remoteAddrIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	realIPHeaderIP := "203.0.113.1"
	extractedRealIP := "203.0.113.10"
	tests := []*struct {
		hasRealIPheader bool
		hasIPExtractor  bool
		extectedXRealIP string
	}{
		{false, false, remoteAddrIP},
		{false, true, extractedRealIP},
		{true, false, realIPHeaderIP},
		{true, true, extractedRealIP},
	}

	for _, tt := range tests {
		if tt.hasRealIPheader {
			req.Header.Set(echo.HeaderXRealIP, realIPHeaderIP)
		} else {
			req.Header.Del(echo.HeaderXRealIP)
		}
		if tt.hasIPExtractor {
			e.IPExtractor = func(*http.Request) string {
				return extractedRealIP
			}
		} else {
			e.IPExtractor = nil
		}
		e.ServeHTTP(rec, req)
		assert.Equal(t, tt.extectedXRealIP, req.Header.Get(echo.HeaderXRealIP), "hasRealIPheader: %t / hasIPExtractor: %t", tt.hasRealIPheader, tt.hasIPExtractor)
	}
}

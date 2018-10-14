// +build go1.11

package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestProxy_1_11(t *testing.T) {
	// Setup
	url1, _ := url.Parse("http://127.0.0.1:27121")
	url2, _ := url.Parse("http://127.0.0.1:27122")

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
	rec := newCloseNotifyRecorder()

	// Remote unreachable
	rec = newCloseNotifyRecorder()
	req.URL.Path = "/api/users"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/api/users", req.URL.Path)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

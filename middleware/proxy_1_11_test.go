// +build go1.11

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
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
	rec := httptest.NewRecorder()

	// Remote unreachable
	rec = httptest.NewRecorder()
	req.URL.Path = "/api/users"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/api/users", req.URL.Path)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestClientCancelConnectionResultsHTTPCode499(t *testing.T) {
	var timeoutStop sync.WaitGroup
	timeoutStop.Add(1)
	HTTPTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeoutStop.Wait() // wait until we have canceled the request
		w.WriteHeader(http.StatusOK)
	}))
	defer HTTPTarget.Close()
	targetURL, _ := url.Parse(HTTPTarget.URL)
	target := &ProxyTarget{
		Name: "target",
		URL:  targetURL,
	}
	rb := NewRandomBalancer(nil)
	assert.True(t, rb.AddTarget(target))
	e := echo.New()
	e.Use(Proxy(rb))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	e.ServeHTTP(rec, req)
	timeoutStop.Done()
	assert.Equal(t, 499, rec.Code)
}

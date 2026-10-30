// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type clientIPTransport func(*http.Request) (*http.Response, error)

func (f clientIPTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientIPSecurity(t *testing.T) {
	_, proxyRange, err := net.ParseCIDR("10.0.0.10/32")
	assert.NoError(t, err)
	trust := []echo.TrustOption{echo.TrustLoopback(false), echo.TrustLinkLocal(false), echo.TrustPrivateNet(false), echo.TrustIPRange(proxyRange)}
	for _, tc := range []struct {
		name      string
		extractor echo.IPExtractor
		peer      string
		xff       [2]string
		realIP    [2]string
		wantIP    string
	}{
		{"default", nil, "198.51.100.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"203.0.113.3", "203.0.113.4"}, "198.51.100.10"},
		{"direct", echo.ExtractIPDirect(), "198.51.100.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"203.0.113.3", "203.0.113.4"}, "198.51.100.10"},
		{"XFF untrusted peer", echo.ExtractIPFromXFFHeader(trust...), "198.51.100.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"203.0.113.3", "203.0.113.4"}, "198.51.100.10"},
		{"XFF trusted proxy", echo.ExtractIPFromXFFHeader(trust...), "10.0.0.10:1234", [2]string{"203.0.113.1, 198.51.100.10", "203.0.113.2, 198.51.100.10"}, [2]string{"203.0.113.3", "203.0.113.4"}, "198.51.100.10"},
		{"real IP untrusted peer", echo.ExtractIPFromRealIPHeader(trust...), "198.51.100.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"203.0.113.3", "203.0.113.4"}, "198.51.100.10"},
		{"real IP trusted proxy", echo.ExtractIPFromRealIPHeader(trust...), "10.0.0.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"198.51.100.10", "198.51.100.10"}, "198.51.100.10"},
		{"custom extractor", func(*http.Request) string { return "custom-client" }, "10.0.0.10:1234", [2]string{"203.0.113.1", "203.0.113.2"}, [2]string{"203.0.113.3", "203.0.113.4"}, "custom-client"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := func(i int) *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.RemoteAddr = tc.peer
				r.Header.Set(echo.HeaderXForwardedFor, tc.xff[i])
				r.Header.Set(echo.HeaderXRealIP, tc.realIP[i])
				r.Header.Add(echo.HeaderXRealIP, "203.0.113.99")
				if i == 1 {
					r.Header.Add(echo.HeaderConnection, "keep-alive, x-ReAl-Ip")
					r.Header.Add(echo.HeaderConnection, "X-Real-IP")
				}
				return r
			}
			t.Run("rate limit", func(t *testing.T) {
				e := echo.New()
				e.IPExtractor = tc.extractor
				store := NewRateLimiterMemoryStoreWithConfig(RateLimiterMemoryStoreConfig{Rate: 1e-9, Burst: 1})
				e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, RateLimiter(store))
				for i, want := range []int{http.StatusOK, http.StatusTooManyRequests} {
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, request(i))
					assert.Equal(t, want, rec.Code)
				}
			})
			t.Run("upstream header", func(t *testing.T) {
				e := echo.New()
				e.IPExtractor = tc.extractor
				target, err := url.Parse("http://upstream.invalid")
				assert.NoError(t, err)
				calls := 0
				e.Use(ProxyWithConfig(ProxyConfig{
					Balancer: NewRoundRobinBalancer([]*ProxyTarget{{URL: target}}),
					Transport: clientIPTransport(func(r *http.Request) (*http.Response, error) {
						calls++
						assert.Equal(t, []string{tc.wantIP}, r.Header.Values(echo.HeaderXRealIP))
						return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("ok")), ContentLength: 2}, nil
					}),
				}))
				for i := range 2 {
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, request(i))
					assert.Equal(t, http.StatusOK, rec.Code)
				}
				assert.Equal(t, 2, calls)
			})
		})
	}
}

func TestClientIPTrustedProxySeparatesClients(t *testing.T) {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	store := NewRateLimiterMemoryStoreWithConfig(RateLimiterMemoryStoreConfig{Rate: 1e-9, Burst: 1})
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, RateLimiter(store))
	for _, client := range []string{"198.51.100.10", "198.51.100.11"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.10:1234"
		req.Header.Set(echo.HeaderXForwardedFor, client)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestProxyClientIPSurvivesConnectionHeader(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, []string{"198.51.100.10"}, r.Header.Values(echo.HeaderXRealIP))
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	assert.NoError(t, err)
	e := echo.New()
	calls := 0
	e.IPExtractor = func(r *http.Request) string {
		calls++
		return echo.ExtractIPDirect()(r)
	}
	e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.10:1234"
	req.Header.Set(echo.HeaderXRealIP, "203.0.113.1")
	req.Header.Set(echo.HeaderXForwardedFor, "203.0.113.2")
	req.Header.Add(echo.HeaderConnection, "keep-alive, x-ReAl-Ip")
	req.Header.Add(echo.HeaderConnection, "X-Real-IP")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, calls)
}

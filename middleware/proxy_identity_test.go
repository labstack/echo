// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package middleware

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestProxyIdentitySurvivesConnectionHeader(t *testing.T) {
	for _, configuredTransport := range []bool{false, true} {
		for _, configuredExtractor := range []bool{false, true} {
			for _, connection := range []string{"", "keep-alive, x-ReAl-Ip"} {
				t.Run(fmt.Sprintf("transport=%t/extractor=%t/connection=%s", configuredTransport, configuredExtractor, connection), func(t *testing.T) {
					upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_ = json.NewEncoder(w).Encode(r.Header.Values(echo.HeaderXRealIP))
					}))
					defer upstream.Close()
					target, err := url.Parse(upstream.URL)
					require.NoError(t, err)
					e := echo.New()
					wantIP := "198.51.100.10"
					if configuredExtractor {
						wantIP = "203.0.113.10"
						e.IPExtractor = func(*http.Request) string { return wantIP }
					}
					config := ProxyConfig{Balancer: NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})}
					if configuredTransport {
						transport := &http.Transport{}
						defer transport.CloseIdleConnections()
						config.Transport = transport
					}
					e.Use(ProxyWithConfig(config))
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.RemoteAddr = "198.51.100.10:1234"
					req.Header.Set(echo.HeaderXRealIP, "203.0.113.1")
					req.Header.Add(echo.HeaderXRealIP, "203.0.113.2")
					req.Header.Set(echo.HeaderXForwardedFor, "203.0.113.3")
					if connection != "" {
						req.Header.Add(echo.HeaderConnection, connection)
						req.Header.Add(echo.HeaderConnection, "X-Real-IP")
					}
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, req)
					require.Equal(t, http.StatusOK, rec.Code)
					var identity []string
					require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &identity))
					require.Equal(t, []string{wantIP}, identity)
				})
			}
		}
	}
}

type proxyIdentityTransport func(*http.Request) (*http.Response, error)

func (f proxyIdentityTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type proxyIdentitySwappingBalancer struct{ ProxyBalancer }

func (b proxyIdentitySwappingBalancer) Next(c *echo.Context) (*ProxyTarget, error) {
	req := c.Request().Clone(c.Request().Context())
	req.Header.Set(echo.HeaderXRealIP, "spoofed-by-replacement")
	c.SetRequest(req)
	return b.ProxyBalancer.Next(c)
}

func TestProxyIdentityCapturedAcrossRetries(t *testing.T) {
	e := echo.New()
	extractions := 0
	e.IPExtractor = func(*http.Request) string {
		extractions++
		return fmt.Sprintf("derived-%d", extractions)
	}
	target, err := url.Parse("http://upstream.invalid")
	require.NoError(t, err)
	attempts := 0
	e.Use(ProxyWithConfig(ProxyConfig{
		Balancer:   proxyIdentitySwappingBalancer{NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})},
		RetryCount: 2,
		Transport: proxyIdentityTransport(func(r *http.Request) (*http.Response, error) {
			attempts++
			require.Equal(t, []string{"derived-1"}, r.Header.Values(echo.HeaderXRealIP))
			if attempts < 3 {
				return nil, errors.New("retry this attempt")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("ok")), ContentLength: 2}, nil
		}),
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderConnection, "X-Real-IP")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 3, attempts)
	require.Equal(t, 1, extractions)
}

func TestProxyIdentityWebSocketConnectionHeader(t *testing.T) {
	received := make(chan []string, 1)
	upstream := httptest.NewServer(websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		received <- conn.Request().Header.Values(echo.HeaderXRealIP)
		_ = websocket.Message.Send(conn, "ok")
	}))
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	target.Scheme = "ws"
	e := echo.New()
	e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
	server := httptest.NewServer(e)
	defer server.Close()
	conn, err := net.Dial("tcp", server.Listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()
	require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
	_, err = fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade, X-Real-IP\r\nUpgrade: websocket\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\nOrigin: http://localhost/\r\nX-Real-IP: 203.0.113.1\r\nX-Forwarded-For: 203.0.113.2\r\n\r\n")
	require.NoError(t, err)
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusSwitchingProtocols, response.StatusCode)
	select {
	case identity := <-received:
		require.Equal(t, []string{"127.0.0.1"}, identity)
	case <-time.After(3 * time.Second):
		t.Fatal("upstream did not receive the upgrade")
	}
}

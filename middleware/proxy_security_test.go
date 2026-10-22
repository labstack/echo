// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package middleware

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestProxyRejectedUpgradeKeepsRouting(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var adminCalls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/admin" {
					adminCalls.Add(1)
					_, _ = io.WriteString(w, "upstream secret")
					return
				}
				w.WriteHeader(status)
				_, _ = io.WriteString(w, "upgrade declined")
			}))
			t.Cleanup(func() { upstream.CloseClientConnections(); upstream.Close() })
			target, err := url.Parse(upstream.URL)
			require.NoError(t, err)
			transport := &http.Transport{}
			t.Cleanup(transport.CloseIdleConnections)
			e := echo.New()
			e.GET("/gateway", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) },
				ProxyWithConfig(ProxyConfig{Balancer: NewRoundRobinBalancer([]*ProxyTarget{{URL: target}}), Transport: transport}))
			e.GET("/admin", func(c *echo.Context) error { return c.NoContent(http.StatusForbidden) })
			server := httptest.NewServer(e)
			t.Cleanup(server.Close)
			conn, err := net.Dial("tcp", server.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
			_, err = fmt.Fprint(conn, "GET /gateway HTTP/1.1\r\nHost: example.com\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
			require.NoError(t, err)
			reader := bufio.NewReader(conn)
			response, err := http.ReadResponse(reader, nil)
			require.NoError(t, err)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, response.Body.Close())
			require.NoError(t, err)
			require.Equal(t, status, response.StatusCode)
			require.Equal(t, "upgrade declined", string(body))
			_, err = fmt.Fprint(conn, "GET /admin HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")
			require.NoError(t, err)
			response, err = http.ReadResponse(reader, nil)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, response.Body)
			require.NoError(t, response.Body.Close())
			require.NoError(t, err)
			require.Equal(t, http.StatusForbidden, response.StatusCode)
			require.Zero(t, adminCalls.Load(), "later requests must return through Echo's route guards")
		})
	}
}

func TestProxyUpgradeClosesBothConnections(t *testing.T) {
	for _, disconnect := range []string{"client", "upstream", "client_upstream_ignores_eof"} {
		t.Run(disconnect, func(t *testing.T) {
			backendConn := make(chan net.Conn, 1)
			backendReturned := make(chan struct{})
			backendReadEOF := make(chan struct{})
			backendRelease := make(chan struct{})
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(backendReturned)
				conn, rw, err := http.NewResponseController(w).Hijack()
				if err != nil {
					return
				}
				defer conn.Close()
				backendConn <- conn
				_, _ = fmt.Fprint(rw, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
				if rw.Flush() != nil {
					return
				}
				message := make([]byte, 4)
				if _, err := io.ReadFull(rw, message); err != nil {
					return
				}
				if string(message) != "ping" {
					return
				}
				_, _ = conn.Write([]byte("pong"))
				_, _ = io.Copy(io.Discard, rw)
				close(backendReadEOF)
				if disconnect == "client_upstream_ignores_eof" {
					// Some upstreams retain the connection after receiving EOF.
					// The proxy must finish without waiting for them to close it.
					<-backendRelease
				}
			}))
			t.Cleanup(upstream.Close)
			t.Cleanup(func() { close(backendRelease) })
			target, err := url.Parse(upstream.URL)
			require.NoError(t, err)
			target.Scheme = "ws"
			e := echo.New()
			proxyReturned := make(chan struct{})
			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c *echo.Context) error { defer close(proxyReturned); return next(c) }
			})
			e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
			server := httptest.NewServer(e)
			t.Cleanup(server.Close)
			conn, err := net.Dial("tcp", server.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
			_, err = fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
			require.NoError(t, err)
			var backend net.Conn
			select {
			case backend = <-backendConn:
			case <-time.After(3 * time.Second):
				t.Fatal("upstream was not reached")
			}
			t.Cleanup(func() { _ = backend.Close() })
			reader := bufio.NewReader(conn)
			response, err := http.ReadResponse(reader, nil)
			require.NoError(t, err)
			require.Equal(t, http.StatusSwitchingProtocols, response.StatusCode)
			_, err = io.WriteString(conn, "ping")
			require.NoError(t, err)
			message := make([]byte, 4)
			_, err = io.ReadFull(reader, message)
			require.NoError(t, err)
			require.Equal(t, "pong", string(message))
			if disconnect != "upstream" {
				require.NoError(t, conn.Close())
			} else {
				require.NoError(t, backend.Close())
			}
			select {
			case <-proxyReturned:
			case <-time.After(3 * time.Second):
				t.Fatal("proxy retained the other side after disconnect")
			}
			require.Equal(t, "ws", target.Scheme, "shared target must not be mutated")
			if disconnect == "client_upstream_ignores_eof" {
				select {
				case <-backendReadEOF:
				case <-time.After(3 * time.Second):
					t.Fatal("upstream did not receive EOF")
				}
				select {
				case <-backendReturned:
					t.Fatal("test upstream must remain open until cleanup")
				default:
				}
				return
			}
			select {
			case <-backendReturned:
			case <-time.After(3 * time.Second):
				t.Fatal("upstream connection was not closed")
			}
		})
	}
}

func TestProxyRemovesForwardedHeader(t *testing.T) {
	for _, upgrade := range []bool{false, true} {
		t.Run(fmt.Sprint(upgrade), func(t *testing.T) {
			received := make(chan http.Header, 1)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received <- r.Header.Clone() }))
			defer upstream.Close()
			target, err := url.Parse(upstream.URL)
			require.NoError(t, err)
			e := echo.New()
			e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Forwarded", "for=127.0.0.1;proto=https")
			if upgrade {
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "websocket")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Empty(t, (<-received).Values("Forwarded"))
		})
	}
}

func TestProxyRejectsMismatchedUpgrade(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Upgrade", "h2c")
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	e := echo.New()
	e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestProxyUpgradeUsesConfiguredTransportAndResponseHook(t *testing.T) {
	upstream := createSimpleWebSocketServer(false)
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	var dials atomic.Int32
	modified := make(chan int, 1)
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		dials.Add(1)
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}}
	defer transport.CloseIdleConnections()
	e := echo.New()
	e.Use(ProxyWithConfig(ProxyConfig{
		Balancer:       NewRoundRobinBalancer([]*ProxyTarget{{URL: target}}),
		Transport:      transport,
		ModifyResponse: func(response *http.Response) error { modified <- response.StatusCode; return nil },
	}))
	server := httptest.NewServer(e)
	defer server.Close()
	conn, err := websocket.Dial("ws"+server.URL[len("http"):], "", "http://localhost/")
	require.NoError(t, err)
	defer conn.Close()
	require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
	require.NoError(t, websocket.Message.Send(conn, "hello"))
	var message string
	require.NoError(t, websocket.Message.Receive(conn, &message))
	require.Equal(t, "hello", message)
	require.Positive(t, dials.Load())
	select {
	case status := <-modified:
		require.Equal(t, http.StatusSwitchingProtocols, status)
	case <-time.After(3 * time.Second):
		t.Fatal("ModifyResponse was not called for the upgrade")
	}
}

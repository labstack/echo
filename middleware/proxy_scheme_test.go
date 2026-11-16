package middleware

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestProxySchemeSurvivesConnectionHeader(t *testing.T) {
	for _, isTLS := range []bool{false, true} {
		for _, connection := range []string{"", "keep-alive, x-forwarded-proto", "X-FORWARDED-PROTO"} {
			t.Run(fmt.Sprintf("tls=%t/connection=%s", isTLS, connection), func(t *testing.T) {
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode(r.Header.Values(echo.HeaderXForwardedProto))
				}))
				defer upstream.Close()

				target, err := url.Parse(upstream.URL)
				require.NoError(t, err)

				e := echo.New()
				wantScheme := "http"
				if isTLS {
					wantScheme = "https"
				}

				config := ProxyConfig{Balancer: NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})}
				e.Use(ProxyWithConfig(config))

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = "198.51.100.10:1234"
				if isTLS {
					req.TLS = &tls.ConnectionState{}
				}
				if connection != "" {
					req.Header.Set(echo.HeaderConnection, connection)
				}

				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code)
				var identity []string
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &identity))
				require.Equal(t, []string{wantScheme}, identity)
			})
		}
	}
}

func TestProxyWebSocketSchemeSurvivesConnectionHeader(t *testing.T) {
	received := make(chan []string, 1)
	upstream := httptest.NewServer(websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		received <- conn.Request().Header.Values(echo.HeaderXForwardedProto)
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

	_, err = fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade, X-Forwarded-Proto\r\nUpgrade: websocket\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\nOrigin: http://localhost/\r\n\r\n")
	require.NoError(t, err)

	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusSwitchingProtocols, response.StatusCode)

	select {
	case proto := <-received:
		require.Equal(t, []string{"http"}, proto)
	case <-time.After(3 * time.Second):
		t.Fatal("upstream did not receive the upgrade")
	}
}

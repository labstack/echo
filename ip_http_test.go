// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealIPScopedIPv6ThroughHTTP(t *testing.T) {
	for _, tc := range []struct {
		name       string
		remoteAddr string
		forwarded  []string
		trusted    string
		want       string
	}{
		{
			name:       "explicit range trusts scoped peer with default trust disabled",
			remoteAddr: "[fe80::1%eth0]:8080",
			forwarded:  []string{"198.51.100.7"},
			trusted:    "fe80::1/128",
			want:       "198.51.100.7",
		},
		{
			name:       "repeated fields stop at nearest untrusted scoped hop",
			remoteAddr: "[fe80::1%eth0]:8080",
			forwarded:  []string{"198.51.100.7, [2001:db8::5%uplink]", "[fe80::2%eth1]"},
			trusted:    "fe80::/64",
			want:       "2001:db8::5",
		},
		{
			name:       "peer outside explicit range retains zone and ignores forged chain",
			remoteAddr: "[fe80::2%eth0]:8080",
			forwarded:  []string{"127.0.0.1"},
			trusted:    "fe80::1/128",
			want:       "fe80::2%eth0",
		},
		{
			name:       "malformed scoped hop rejects entire chain",
			remoteAddr: "[fe80::1%eth0]:8080",
			forwarded:  []string{"198.51.100.7", "[fe80:::2%eth1]"},
			trusted:    "fe80::/64",
			want:       "fe80::1%eth0",
		},
		{
			name:       "empty zone rejects entire chain",
			remoteAddr: "[fe80::1%eth0]:8080",
			forwarded:  []string{"198.51.100.7, [fe80::2%]"},
			trusted:    "fe80::/64",
			want:       "fe80::1%eth0",
		},
		{
			name:       "malformed scoped peer cannot authorize forwarding headers",
			remoteAddr: "[fe80:::1%eth0]:8080",
			forwarded:  []string{"127.0.0.1"},
			trusted:    "fe80::/64",
			want:       "fe80:::1%eth0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.IPExtractor = ExtractIPFromXFFHeader(
				TrustLoopback(false), TrustLinkLocal(false), TrustPrivateNet(false),
				TrustIPRange(mustParseCIDR(tc.trusted)),
			)
			e.GET("/client-ip", func(c *Context) error {
				return c.String(http.StatusOK, c.RealIP())
			})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Exercise a real TCP HTTP request while substituting the proxy peer.
				// This fixture does not require a local link-local IPv6 interface.
				r.RemoteAddr = tc.remoteAddr
				e.ServeHTTP(w, r)
			}))
			t.Cleanup(server.Close)

			req, err := http.NewRequest(http.MethodGet, server.URL+"/client-ip", nil)
			require.NoError(t, err)
			for _, value := range tc.forwarded {
				req.Header.Add(HeaderXForwardedFor, value)
			}
			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, tc.want, string(body))
		})
	}
}

func TestSchemeScopedIPv6ThroughHTTP(t *testing.T) {
	for _, tc := range []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{
			name:       "trusted scoped peer can set forwarded scheme",
			remoteAddr: "[fe80::1%eth0]:8080",
			want:       "https",
		},
		{
			name:       "IPv4 with zone-like suffix cannot set forwarded scheme",
			remoteAddr: "127.0.0.1%garbage:8080",
			want:       "http",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.SchemeExtractor = ExtractSchemeFromHeaders()
			e.GET("/scheme", func(c *Context) error {
				return c.String(http.StatusOK, c.Scheme())
			})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Match the proxy-peer fixture used by the RealIP regression above.
				r.RemoteAddr = tc.remoteAddr
				e.ServeHTTP(w, r)
			}))
			t.Cleanup(server.Close)

			req, err := http.NewRequest(http.MethodGet, server.URL+"/scheme", nil)
			require.NoError(t, err)
			req.Header.Set(HeaderXForwardedProto, "https")
			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, tc.want, string(body))
		})
	}
}

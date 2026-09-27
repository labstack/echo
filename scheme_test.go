// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSchemeDirect(t *testing.T) {
	var testCases = []struct {
		name            string
		givenIsTLS      bool
		givenRemoteAddr string
		givenHeaders    http.Header
		expect          string
	}{
		{
			name:            "returns https for TLS",
			givenIsTLS:      true,
			givenRemoteAddr: "203.0.113.10:1234",
			expect:          "https",
		},
		{
			name:            "returns http without TLS",
			givenRemoteAddr: "203.0.113.10:1234",
			expect:          "http",
		},
		{
			name:            "ignores forwarding headers even from trusted network",
			givenRemoteAddr: "10.0.0.1:1234",
			givenHeaders: http.Header{
				HeaderXForwardedProto:    []string{"https"},
				HeaderXForwardedProtocol: []string{"https"},
				HeaderXForwardedSsl:      []string{"on"},
				HeaderXUrlScheme:         []string{"https"},
			},
			expect: "http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.givenRemoteAddr
			if tc.givenHeaders != nil {
				req.Header = tc.givenHeaders
			}
			if tc.givenIsTLS {
				req.TLS = &tls.ConnectionState{}
			}

			assert.Equal(t, tc.expect, ExtractSchemeDirect()(req))
		})
	}
}

func TestExtractSchemeFromHeaders(t *testing.T) {
	var testCases = []struct {
		name            string
		givenOptions    []TrustOption
		givenIsTLS      bool
		givenRemoteAddr string
		expect          string
	}{
		{
			name:            "trusts private network by default",
			givenRemoteAddr: "10.0.0.1:1234",
			expect:          "https",
		},
		{
			name:            "trusts loopback by default",
			givenRemoteAddr: "127.0.0.1:1234",
			expect:          "https",
		},
		{
			name:            "trusts link-local by default",
			givenRemoteAddr: "169.254.0.1:1234",
			expect:          "https",
		},
		{
			name:            "trusts IPv6 link-local address with zone by default",
			givenRemoteAddr: "[fe80::1%eth0]:1234",
			expect:          "https",
		},
		{
			name:            "trusts unix socket peer by default (Linux)",
			givenRemoteAddr: "@",
			expect:          "https",
		},
		{
			name:            "trusts unix socket peer by default (Linux abstract socket)",
			givenRemoteAddr: "@client",
			expect:          "https",
		},
		{
			name:            "trusts unix socket peer by default (bound socket path)",
			givenRemoteAddr: "/tmp/client.sock",
			expect:          "https",
		},
		{
			name:            "trusts unix socket peer by default (unnamed)",
			givenRemoteAddr: "",
			expect:          "https",
		},
		{
			name:            "trusts private address without port",
			givenRemoteAddr: "10.0.0.1",
			expect:          "https",
		},
		{
			name:            "does not trust public address by default",
			givenRemoteAddr: "203.0.113.10:1234",
			expect:          "http",
		},
		{
			name:            "does not trust unparsable remote address",
			givenRemoteAddr: "not-an-ip:1234",
			expect:          "http",
		},
		{
			name:            "returns https for TLS from public address",
			givenIsTLS:      true,
			givenRemoteAddr: "203.0.113.10:1234",
			expect:          "https",
		},
		{
			name:            "trusts public address in configured range",
			givenOptions:    []TrustOption{TrustIPRange(mustParseCIDR("203.0.113.0/24"))},
			givenRemoteAddr: "203.0.113.10:1234",
			expect:          "https",
		},
		{
			name:            "does not trust private network when disabled",
			givenOptions:    []TrustOption{TrustPrivateNet(false)},
			givenRemoteAddr: "10.0.0.1:1234",
			expect:          "http",
		},
		{
			name:            "does not trust loopback when disabled",
			givenOptions:    []TrustOption{TrustLoopback(false)},
			givenRemoteAddr: "127.0.0.1:1234",
			expect:          "http",
		},
		{
			name:            "does not trust unix socket peer when loopback is disabled",
			givenOptions:    []TrustOption{TrustLoopback(false)},
			givenRemoteAddr: "@",
			expect:          "http",
		},
		{
			name:            "does not trust link-local when disabled",
			givenOptions:    []TrustOption{TrustLinkLocal(false)},
			givenRemoteAddr: "169.254.0.1:1234",
			expect:          "http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.givenRemoteAddr
			req.Header.Set(HeaderXForwardedProto, "https")
			if tc.givenIsTLS {
				req.TLS = &tls.ConnectionState{}
			}

			assert.Equal(t, tc.expect, ExtractSchemeFromHeaders(tc.givenOptions...)(req))
		})
	}
}

func TestLegacySchemeExtractor(t *testing.T) {
	var testCases = []struct {
		name         string
		givenIsTLS   bool
		givenHeaders http.Header
		expect       string
	}{
		{
			name:         "returns https for TLS",
			givenIsTLS:   true,
			givenHeaders: http.Header{HeaderXForwardedProto: []string{"http"}},
			expect:       "https",
		},
		{
			name:         "trusts X-Forwarded-Proto from public address",
			givenHeaders: http.Header{HeaderXForwardedProto: []string{"https"}},
			expect:       "https",
		},
		{
			name:         "still rejects invalid values",
			givenHeaders: http.Header{HeaderXForwardedProto: []string{"javascript"}},
			expect:       "http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "203.0.113.10:1234"
			req.Header = tc.givenHeaders
			if tc.givenIsTLS {
				req.TLS = &tls.ConnectionState{}
			}

			assert.Equal(t, tc.expect, LegacySchemeExtractor()(req))
		})
	}
}

func TestContext_Scheme_usesEchoSchemeExtractor(t *testing.T) {
	e := New()
	e.SchemeExtractor = LegacySchemeExtractor()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set(HeaderXForwardedProto, "https")
	c := e.NewContext(req, httptest.NewRecorder())

	assert.Equal(t, "https", c.Scheme())
}

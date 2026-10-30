// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextRealIPIgnoresUntrustedHeaders(t *testing.T) {
	for _, remote := range []struct {
		address string
		want    string
	}{
		{"198.51.100.10:1234", "198.51.100.10"},
		{"[2001:db8::10]:1234", "2001:db8::10"},
		{"198.51.100.10", "198.51.100.10"},
		{"2001:db8::10", "2001:db8::10"},
		{"invalid", ""},
	} {
		t.Run(remote.address, func(t *testing.T) {
			for _, headers := range []http.Header{
				{},
				{HeaderXForwardedFor: {"203.0.113.1"}},
				{HeaderXForwardedFor: {"203.0.113.1, 203.0.113.2", "203.0.113.3"}},
				{HeaderXRealIP: {"203.0.113.1"}},
				{HeaderXRealIP: {"[2001:db8::1]"}},
				{HeaderXForwardedFor: {"invalid"}, HeaderXRealIP: {"203.0.113.1"}},
			} {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = remote.address
				req.Header = headers
				assert.Equal(t, remote.want, New().NewContext(req, httptest.NewRecorder()).RealIP())
				assert.Equal(t, remote.want, (&context{request: req}).RealIP())
			}
		})
	}
}

func TestContextRealIPPreservesCustomExtractor(t *testing.T) {
	e := New()
	e.IPExtractor = func(r *http.Request) string {
		return r.Header.Get("Custom-Client-IP")
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Custom-Client-IP", "203.0.113.10")
	assert.Equal(t, "203.0.113.10", e.NewContext(req, httptest.NewRecorder()).RealIP())
}

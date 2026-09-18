// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net"
	"net/http"
	"strings"
	"testing"
)

// Keep the original implementation as a compatibility oracle for malformed
// headers, multiple header lines, and custom trust configurations.
func referenceExtractXFF(req *http.Request, checker *ipChecker) string {
	directIP := extractIP(req)
	xffs := req.Header[HeaderXForwardedFor]
	if len(xffs) == 0 {
		return directIP
	}
	ips := append(strings.Split(strings.Join(xffs, ","), ","), directIP)
	for i := len(ips) - 1; i >= 0; i-- {
		ips[i] = strings.TrimSpace(ips[i])
		ips[i] = strings.TrimPrefix(ips[i], "[")
		ips[i] = strings.TrimSuffix(ips[i], "]")
		ip := net.ParseIP(ips[i])
		if ip == nil {
			return directIP
		}
		if !checker.trust(ip) {
			return ip.String()
		}
	}
	return strings.TrimSpace(ips[0])
}

func FuzzExtractIPFromXFFHeaderCompatibility(f *testing.F) {
	for _, seed := range [][3]string{
		{"10.0.0.1:80", "10.0.0.2", "127.0.0.1"},
		{"10.0.0.1:80", "203.0.113.1, 10.0.0.2", "127.0.0.1"},
		{"[::1]:80", " [fc00:0::1] ", "[::1]"},
		{"10.0.0.1:80", "invalid", "203.0.113.1"},
		{"10.0.0.1:80", "10.0.0.2,", ""},
		{"invalid", "", ""},
		{"203.0.113.1:80", "invalid", ""},
		{"10.0.0.1:80", "[ 10.0.0.2 ]", "10.0.0.3"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	configurations := [][]TrustOption{
		nil,
		{TrustLoopback(false), TrustLinkLocal(false), TrustPrivateNet(false)},
		{TrustIPRange(mustParseCIDR("0.0.0.0/0")), TrustIPRange(mustParseCIDR("::/0"))},
	}
	f.Fuzz(func(t *testing.T, remote, first, second string) {
		for _, options := range configurations {
			extractor := ExtractIPFromXFFHeader(options...)
			checker := newIPChecker(options)
			for _, headers := range [][]string{nil, {}, {first}, {first, second}} {
				req := &http.Request{RemoteAddr: remote, Header: http.Header{HeaderXForwardedFor: headers}}
				if got, want := extractor(req), referenceExtractXFF(req, checker); got != want {
					t.Fatalf("remote=%q headers=%q: got %q, want %q", remote, headers, got, want)
				}
			}
		}
	})
}

var benchmarkExtractedIP string

func BenchmarkExtractIPFromXFFHeader(b *testing.B) {
	for _, tc := range []struct {
		name, remote string
		headers      []string
	}{
		{"TrustedChain", "10.0.0.1:80", []string{"10.0.0.2, 10.0.0.3", "127.0.0.1"}},
		{"PublicClient", "10.0.0.1:80", []string{"203.0.113.1, 10.0.0.2", "127.0.0.1"}},
		{"UntrustedPeer", "203.0.113.1:80", []string{"10.0.0.2, 10.0.0.3", "127.0.0.1"}},
		{"TrustedIPv6", "[::1]:80", []string{"[fc00::1], [fc00::2]", "[::1]"}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			req := &http.Request{RemoteAddr: tc.remote, Header: http.Header{HeaderXForwardedFor: tc.headers}}
			extractor := ExtractIPFromXFFHeader()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkExtractedIP = extractor(req)
			}
		})
	}
}

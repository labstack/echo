// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net"
	"net/http"
	"strings"
)

// SchemeExtractor is a function to determine the scheme (for example `http` or `https`) of http.Request.
// Set appropriate one to Echo#SchemeExtractor.
//
// When Echo#SchemeExtractor is not set, Context.Scheme uses ExtractSchemeFromHeaders with the default trust options.
type SchemeExtractor func(*http.Request) string

// ExtractSchemeDirect returns a SchemeExtractor that uses only the actual connection: `https` for TLS connections
// and `http` otherwise. Forwarding headers are ignored.
// Use this if your server faces the internet directly (i.e.: uses no proxy).
func ExtractSchemeDirect() SchemeExtractor {
	return extractSchemeDirect
}

func extractSchemeDirect(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

// ExtractSchemeFromHeaders returns a SchemeExtractor that uses the `X-Forwarded-Proto`, `X-Forwarded-Protocol`,
// `X-Forwarded-Ssl` and `X-Url-Scheme` headers only when the request comes directly from a trusted address
// (http.Request.RemoteAddr). For requests from other addresses the headers are ignored and the scheme of the actual
// connection is used.
//
// By default, loopback, link-local and private network addresses and unix socket peers are trusted. Use TrustOption
// to change this, for example TrustIPRange to trust a proxy that connects from a public address (such as a CDN or a
// cloud load balancer). TrustLoopback(false) also stops trusting unix socket peers.
//
// When `X-Forwarded-Proto` is present, the other headers are ignored, and the last value is used. The trusted proxy
// must therefore set (overwrite) `X-Forwarded-Proto` rather than pass through the value sent by the client.
//
// This is the default strategy when Echo#SchemeExtractor is not set.
func ExtractSchemeFromHeaders(options ...TrustOption) SchemeExtractor {
	checker := newIPChecker(options)
	return func(req *http.Request) string {
		return extractScheme(req, checker)
	}
}

// LegacySchemeExtractor returns a SchemeExtractor that uses the forwarding headers from any client. This was the
// behavior of Context.Scheme before the address of the client was checked.
//
// It is not safe against spoofing unless every request passes through a proxy that sets or removes the
// `X-Forwarded-Proto`, `X-Forwarded-Protocol`, `X-Forwarded-Ssl` and `X-Url-Scheme` headers.
// Use ExtractSchemeFromHeaders instead.
func LegacySchemeExtractor() SchemeExtractor {
	return legacySchemeExtractor
}

func legacySchemeExtractor(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	return schemeFromHeaders(req)
}

// defaultSchemeChecker is used by Context.Scheme when Echo#SchemeExtractor is not set.
var defaultSchemeChecker = newIPChecker(nil)

func extractScheme(req *http.Request, checker *ipChecker) string {
	if req.TLS != nil {
		return "https"
	}
	if !isTrustedPeer(req.RemoteAddr, checker) {
		return "http"
	}
	return schemeFromHeaders(req)
}

// isTrustedPeer reports whether the direct peer of the request is trusted to set forwarding headers.
func isTrustedPeer(remoteAddr string, checker *ipChecker) bool {
	// Unix socket peers are on the same host, so they are trusted like loopback addresses. net/http reports them
	// as an empty string, as "@" or "@name" (Linux, unnamed or abstract socket) or as the path of a bound socket.
	if remoteAddr == "" || remoteAddr[0] == '@' || remoteAddr[0] == '/' {
		return checker.trustLoopback
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	if i := strings.IndexByte(host, '%'); i != -1 { // IPv6 zone, e.g. "fe80::1%eth0"
		host = host[:i]
	}
	ip := net.ParseIP(host)
	return ip != nil && checker.trust(ip)
}

func schemeFromHeaders(req *http.Request) string {
	// When X-Forwarded-Proto is present it is the only header used: a trusted proxy sets it, and the other headers
	// might be client-supplied values that the proxy did not remove. With several values (repeated header lines or a
	// comma-separated list) the last one was added by the nearest proxy.
	if values := req.Header.Values(HeaderXForwardedProto); len(values) > 0 && values[len(values)-1] != "" {
		last := values[len(values)-1]
		if i := strings.LastIndexByte(last, ','); i != -1 {
			last = last[i+1:]
		}
		if scheme, ok := canonicalProto(strings.TrimSpace(last)); ok {
			return scheme
		}
		return "http"
	}
	if scheme, ok := canonicalProto(req.Header.Get(HeaderXForwardedProtocol)); ok {
		return scheme
	}
	if ssl := req.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme, ok := canonicalProto(req.Header.Get(HeaderXUrlScheme)); ok {
		return scheme
	}
	return "http"
}

// canonicalProto returns the lowercase form of proto if it is `http`, `https`, `ws` or `wss` (in any case).
func canonicalProto(proto string) (string, bool) {
	for _, p := range [...]string{"http", "https", "ws", "wss"} {
		if strings.EqualFold(proto, p) {
			return p, true
		}
	}
	return "", false
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

// Package pathutil holds internal helpers for safely handling request and file
// paths. It is internal so it can be shared between the echo package and the
// middleware package without becoming part of the public API.
package pathutil

import "strings"

// HasEncodedPathSeparator reports whether s contains a percent-encoded path
// separator, case-insensitively: %2F/%2f (forward slash) or %5C/%5c (backslash).
// Backslash is included as defense-in-depth against Windows-style separators even
// though fs.FS itself only uses forward slashes.
//
// Such sequences let an attacker smuggle a separator past the router, which by
// default matches on the raw encoded path, so they must be rejected before
// unescaping when resolving static files.
func HasEncodedPathSeparator(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		switch {
		case s[i+1] == '2' && (s[i+2] == 'f' || s[i+2] == 'F'): // %2F
			return true
		case s[i+1] == '5' && (s[i+2] == 'c' || s[i+2] == 'C'): // %5C
			return true
		}
	}
	return false
}

// HasDotDotSegment reports whether p, split on '/', contains a ".." path segment.
// Both ends of the string and every '/' act as boundaries, so "..", "../x", "x/..",
// and "a/../b" match while "..foo" and "foo.." do not.
//
// A ".." segment is parent-directory traversal. The router matches a specific path
// prefix, but a ".." that only becomes a real segment after decoding (encoded as
// "%2E%2E", or decoded by the router itself when UseEscapedPathForMatching is set)
// is never seen as traversal during routing. Cleaning it then resolves a file
// across the route or static-mount boundary the matched route authorized, bypassing
// route-level middleware (GHSA-3pmx-cf9f-34xr). This mirrors the "no .. element"
// invariant of fs.ValidPath; no real filename is "..", so reject it.
//
// Only '/' is a separator here: encoded backslashes are rejected earlier by
// HasEncodedPathSeparator, and on a forward-slash fs.FS a literal backslash never
// acts as a separator, so a "..\\" segment cannot traverse.
func HasDotDotSegment(p string) bool {
	for len(p) > 0 {
		seg := p
		if i := strings.IndexByte(p, '/'); i >= 0 {
			seg, p = p[:i], p[i+1:]
		} else {
			p = ""
		}
		if seg == ".." {
			return true
		}
	}
	return false
}

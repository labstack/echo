// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

// Package pathutil holds internal helpers for safely handling request and file
// paths. It is internal so it can be shared between the echo package and the
// middleware package without becoming part of the public API.
package pathutil

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

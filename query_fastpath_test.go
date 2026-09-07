// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/url"
	"testing"
)

const benchQuery = "id=42&name=Jon&lang=en&page=2"

// BenchmarkQueryParam_FastPath measures the single-key raw scan (current behavior).
func BenchmarkQueryParam_FastPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = getRawQueryParam(benchQuery, "name")
	}
}

// BenchmarkQueryParam_Map measures the previous behavior: build the full url.Values then Get.
func BenchmarkQueryParam_Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, _ := url.ParseQuery(benchQuery)
		_ = v.Get("name")
	}
}

// TestGetRawQueryParamMatchesStdlib asserts the single-key fast path returns exactly what
// url.Values.Get (over url.ParseQuery output) would return, across a battery of edge cases.
func TestGetRawQueryParamMatchesStdlib(t *testing.T) {
	queries := []string{
		"",
		"a=1",
		"a=1&b=2",
		"a=1&a=2",    // first match wins
		"a=&b=2",     // empty value
		"=v",         // empty key
		"a",          // key without '='
		"a=x+y",      // '+' -> space
		"a=%20space", // percent escape
		"a=%2",       // bad value escape -> pair skipped
		"a=%ZZ&a=2",  // bad escape on first match -> skip, fall through to second
		"%ZZ=1&a=2",  // bad key escape -> pair skipped
		"a%3Db=c",    // encoded key 'a=b'
		"a=1;b=2",    // semicolon segment skipped entirely
		"a=1&c=3;d=4&e=5",
		"name=Jon&name=Snow&age=24",
		"q=%E4%B8%AD%E6%96%87", // unicode value
		"empty=&x=1",
		"a=1&&b=2", // empty segment
	}
	names := []string{"a", "b", "c", "d", "e", "name", "age", "q", "empty", "x", "a=b", "missing", ""}

	for _, q := range queries {
		want, _ := url.ParseQuery(q) // url.Values; mirrors URL.Query() (error ignored)
		for _, name := range names {
			got := getRawQueryParam(q, name)
			exp := want.Get(name)
			if got != exp {
				t.Errorf("getRawQueryParam(%q, %q) = %q; url.Values.Get = %q", q, name, got, exp)
			}
		}
	}
}

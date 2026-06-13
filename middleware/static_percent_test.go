// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// Regression test for #2599: a file whose name contains a percent sign must be
// downloadable. http.Request.URL.Path is already decoded by net/http, so the
// static middleware must not unescape it a second time (which turned
// "/100%25.txt" into an "invalid URL escape" error or a missing file).
func TestStatic_servesFileWithPercentInName(t *testing.T) {
	e := echo.New()
	e.Use(StaticWithConfig(StaticConfig{
		Root: ".",
		Filesystem: fstest.MapFS{
			"100%.txt":      &fstest.MapFile{Data: []byte("hundred percent")},
			"foo%20bar.txt": &fstest.MapFile{Data: []byte("literal percent twenty")},
		},
	}))

	cases := map[string]string{
		"/100%25.txt":      "hundred percent",
		"/foo%2520bar.txt": "literal percent twenty",
	}
	for url, want := range cases {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "GET %s should serve the file", url)
		assert.Equal(t, want, rec.Body.String(), "GET %s should return the file contents", url)
	}
}

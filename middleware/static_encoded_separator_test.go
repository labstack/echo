// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Regression for GHSA-vfp3-v2gw-7wfq (v4 backport): the static middleware mounted on a
// group must not let an encoded separator in the wildcard bypass route-level middleware
// and disclose a file the matched route never authorized.
func TestStatic_EncodedSeparatorDoesNotBypassRoute(t *testing.T) {
	root := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "admin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "admin", "secret.txt"), []byte("TOP-SECRET"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("public"), 0o644))

	e := echo.New()
	g := e.Group("/files", StaticWithConfig(StaticConfig{Root: root}))
	g.GET("/*", func(c echo.Context) error { return echo.ErrNotFound })

	cases := []struct {
		target   string
		wantCode int
	}{
		{"/files/index.html", http.StatusOK},
		{"/files/admin%2Fsecret.txt", http.StatusNotFound},
		{"/files/admin%2fsecret.txt", http.StatusNotFound},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.target, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
		assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
	}
}

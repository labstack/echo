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
	var testCases = []struct {
		name     string
		config   StaticConfig
		target   string
		wantCode int
	}{
		{
			name:     "ok, legitimate file is served",
			target:   "/files/index.html",
			wantCode: http.StatusOK,
		},
		{
			// With EnablePathUnescaping=false (default/safe), the wildcard param "admin%2Fsecret.txt"
			// is NOT decoded, so the FS lookup is for literal "admin%2Fsecret.txt" which does
			// not exist → falls through → 404. ACL is not bypassed.
			name:     "ok, encoded slash returns 404 with default safe config (EnablePathUnescaping=false)",
			target:   "/files/admin%2Fsecret.txt",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "ok, lower-case encoded slash also returns 404",
			target:   "/files/admin%2fsecret.txt",
			wantCode: http.StatusNotFound,
		},
		{
			// With EnablePathUnescaping=true, the wildcard param "admin%2Fsecret.txt" IS decoded
			// to "admin/secret.txt" before the FS lookup. The router already routed to /* so the
			// ACL guard on /admin/* never ran. The file is served — ACL bypass.
			// Only use EnablePathUnescaping: true when not relying on route-based ACL guards.
			name:     "nok, encoded slash bypasses ACL when EnablePathUnescaping=true",
			config:   StaticConfig{EnablePathUnescaping: true},
			target:   "/files/admin%2Fsecret.txt",
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			assert.NoError(t, os.MkdirAll(filepath.Join(root, "admin"), 0o755))
			assert.NoError(t, os.WriteFile(filepath.Join(root, "admin", "secret.txt"), []byte("TOP-SECRET"), 0o644))
			assert.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte("public"), 0o644))

			cfg := tc.config
			cfg.Root = root

			e := echo.New()
			g := e.Group("/files", StaticWithConfig(cfg))
			g.GET("/*", func(c echo.Context) error { return echo.ErrNotFound })

			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
			if tc.wantCode != http.StatusOK {
				assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
			}
		})
	}
}

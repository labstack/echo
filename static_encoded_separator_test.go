package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

// Regression for GHSA-vfp3-v2gw-7wfq: an encoded slash (%2F) must not let a static
// file request resolve across a path separator and bypass route-level middleware.
func TestStaticDirectoryHandler_EncodedSeparatorDoesNotBypassRoute(t *testing.T) {
	var testCases = []struct {
		name     string
		target   string
		wantCode int
		wantBody string
	}{
		{
			name:     "protected route fires",
			target:   "/admin/secret.txt",
			wantCode: http.StatusForbidden,
			wantBody: "denied",
		},
		{
			name:     "encoded slash rejected, no disclosure",
			target:   "/admin%2Fsecret.txt",
			wantCode: http.StatusNotFound,
			wantBody: "",
		},
		{
			name:     "lower-case hex variant",
			target:   "/admin%2fsecret.txt",
			wantCode: http.StatusNotFound,
			wantBody: "",
		},
		{
			name:     "encoded backslash variant - Windows specific related",
			target:   "/admin%5Csecret.txt",
			wantCode: http.StatusNotFound,
			wantBody: "",
		},
		{
			name:     "double-encoded: single unescape -> literal filename, not a separator",
			target:   "/admin%252Fsecret.txt",
			wantCode: http.StatusNotFound,
			wantBody: "",
		},
		{
			name:     "legitimate static file still served",
			target:   "/index.html",
			wantCode: http.StatusOK,
			wantBody: "public",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"admin/secret.txt": {Data: []byte("TOP-SECRET")},
				"index.html":       {Data: []byte("public")},
			}
			e := New()
			g := e.Group("/admin", func(next HandlerFunc) HandlerFunc {
				return func(c *Context) error { return c.String(http.StatusForbidden, "denied") }
			})
			g.GET("/*", func(c *Context) error { return c.String(http.StatusOK, "reached-protected-handler") })
			e.StaticFS("/", fsys)

			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, rec.Body.String(), "GET %s", tc.target)
			}
			assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
		})
	}
}

// A Group-mounted StaticFS shares StaticDirectoryHandler, so it must reject the
// same encoded separators when served under a non-root prefix.
func TestGroupStaticFS_EncodedSeparatorDoesNotBypassRoute(t *testing.T) {
	var testCases = []struct {
		name     string
		target   string
		wantCode int
	}{
		{
			name:     "ok",
			target:   "/files/index.html",
			wantCode: http.StatusOK,
		},
		{
			name:     "nok, encoded slash",
			target:   "/files/admin%2Fsecret.txt",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "nok encoded backslash",
			target:   "/files/admin%5Csecret.txt",
			wantCode: http.StatusNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"admin/secret.txt": {Data: []byte("TOP-SECRET")},
				"index.html":       {Data: []byte("public")},
			}
			e := New()
			g := e.Group("/files")
			g.StaticFS("/", fsys)

			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
			assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
		})
	}
}

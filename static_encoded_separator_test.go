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

	cases := []struct {
		target   string
		wantCode int
		wantBody string
	}{
		{"/admin/secret.txt", http.StatusForbidden, "denied"}, // protected route fires
		{"/admin%2Fsecret.txt", http.StatusNotFound, ""},      // encoded slash rejected, no disclosure
		{"/admin%2fsecret.txt", http.StatusNotFound, ""},      // lower-case hex variant
		{"/admin%5Csecret.txt", http.StatusNotFound, ""},      // encoded backslash variant
		{"/admin%252Fsecret.txt", http.StatusNotFound, ""},    // double-encoded: single unescape -> literal filename, not a separator
		{"/index.html", http.StatusOK, "public"},              // legitimate static file still served
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.target, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
		if tc.wantBody != "" {
			assert.Equal(t, tc.wantBody, rec.Body.String(), "GET %s", tc.target)
		}
		assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
	}
}

// A Group-mounted StaticFS shares StaticDirectoryHandler, so it must reject the
// same encoded separators when served under a non-root prefix.
func TestGroupStaticFS_EncodedSeparatorDoesNotBypassRoute(t *testing.T) {
	fsys := fstest.MapFS{
		"admin/secret.txt": {Data: []byte("TOP-SECRET")},
		"index.html":       {Data: []byte("public")},
	}
	e := New()
	g := e.Group("/files")
	g.StaticFS("/", fsys)

	cases := []struct {
		target   string
		wantCode int
	}{
		{"/files/index.html", http.StatusOK},
		{"/files/admin%2Fsecret.txt", http.StatusNotFound},
		{"/files/admin%5Csecret.txt", http.StatusNotFound},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.target, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
		assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
	}
}

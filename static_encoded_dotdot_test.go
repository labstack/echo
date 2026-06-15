package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func adminSecretFS() fstest.MapFS {
	return fstest.MapFS{
		"admin/secret.txt": {Data: []byte("TOP-SECRET")},
		"public/ok.txt":    {Data: []byte("public")},
		"index.html":       {Data: []byte("index")},
	}
}

// Regression for GHSA-3pmx-cf9f-34xr: encoded/decoded ".." in a static wildcard must
// not resolve a file across a route-level middleware guard. With unescaping disabled
// by default the encoded form never decodes, and the dot-dot guard rejects any ".."
// that the router (e.g. UseEscapedPathForMatching) decoded itself.
func TestStaticFS_EncodedDotDotDoesNotBypassRoute(t *testing.T) {
	run := func(t *testing.T, e *Echo) {
		g := e.Group("/admin", func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error { return c.String(http.StatusForbidden, "denied") }
		})
		g.GET("/*", func(c *Context) error { return c.String(http.StatusOK, "reached-protected-handler") })
		e.StaticFS("/", adminSecretFS())

		cases := []struct {
			target   string
			wantCode int
		}{
			{"/admin/secret.txt", http.StatusForbidden},                // protected route fires
			{"/public/%2E%2E/admin/secret.txt", http.StatusNotFound},   // high-sev: encoded dot-dot
			{"/public/%2e%2e/admin/secret.txt", http.StatusNotFound},   // lower-case hex
			{"/public%2F..%2Fadmin%2Fsecret.txt", http.StatusNotFound}, // encoded-slash variant
			{"/public/../admin/secret.txt", http.StatusNotFound},       // literal dot-dot
			{"/public/ok.txt", http.StatusOK},                          // legitimate static file
			{"/index.html", http.StatusOK},                             // legitimate static file
		}
		for _, tc := range cases {
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
			assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
		}
	}

	t.Run("default router", func(t *testing.T) { run(t, New()) })
	t.Run("UseEscapedPathForMatching", func(t *testing.T) {
		run(t, NewWithConfig(Config{Router: NewRouter(RouterConfig{UseEscapedPathForMatching: true})}))
	})
}

// With EnablePathUnescapingStaticFiles the wildcard is unescaped (so encoded file
// names work), but encoded separators and ".." segments are still rejected.
func TestStaticFS_EnablePathUnescapingStaticFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"hello world.txt":  {Data: []byte("spaced")},
		"admin/secret.txt": {Data: []byte("TOP-SECRET")},
	}
	e := NewWithConfig(Config{EnablePathUnescapingStaticFiles: true})
	g := e.Group("/admin", func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error { return c.String(http.StatusForbidden, "denied") }
	})
	g.GET("/*", func(c *Context) error { return c.String(http.StatusOK, "reached") })
	e.StaticFS("/", fsys)

	cases := []struct {
		target   string
		wantCode int
		wantBody string
	}{
		{"/hello%20world.txt", http.StatusOK, "spaced"},              // encoded space now decoded and served
		{"/admin%2Fsecret.txt", http.StatusNotFound, ""},             // encoded slash still rejected
		{"/public/%2E%2E/admin/secret.txt", http.StatusNotFound, ""}, // encoded dot-dot still rejected
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

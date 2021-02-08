package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

//Assert expected with url.EscapedPath method to obtain the path.
func TestRewrite(t *testing.T) {
	e := echo.New()
	e.Use(RewriteWithConfig(RewriteConfig{
		Rules: map[string]string{
			"/old":              "/new",
			"/api/*":            "/$1",
			"/js/*":             "/public/javascripts/$1",
			"/users/*/orders/*": "/user/$1/order/$2",
		},
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	req.URL, _ = url.Parse("/api/users")
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/users", req.URL.EscapedPath())
	req.URL, _ = url.Parse("/js/main.js")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/public/javascripts/main.js", req.URL.EscapedPath())
	req.URL, _ = url.Parse("/old")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.EscapedPath())
	req.URL, _ = url.Parse("/users/jack/orders/1")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jack/order/1", req.URL.EscapedPath())
	req.URL, _ = url.Parse("/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F", req.URL.EscapedPath())
	req.URL, _ = url.Parse("/api/new users")
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new%20users", req.URL.EscapedPath())
}

// Issue #1086
func TestEchoRewritePreMiddleware(t *testing.T) {
	e := echo.New()
	r := e.Router()

	// Rewrite old url to new one
	e.Pre(Rewrite(map[string]string{
		"/old": "/new",
	},
	))

	// Route
	r.Add(http.MethodGet, "/new", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.EscapedPath())
	assert.Equal(t, http.StatusOK, rec.Code)
}

// Issue #1143
func TestRewriteWithConfigPreMiddleware_Issue1143(t *testing.T) {
	e := echo.New()
	r := e.Router()

	e.Pre(RewriteWithConfig(RewriteConfig{
		Rules: map[string]string{
			"/api/*/mgmt/proj/*/agt": "/api/$1/hosts/$2",
			"/api/*/mgmt/proj":       "/api/$1/eng",
		},
	}))

	r.Add(http.MethodGet, "/api/:version/hosts/:name", func(c echo.Context) error {
		return c.String(http.StatusOK, "hosts")
	})
	r.Add(http.MethodGet, "/api/:version/eng", func(c echo.Context) error {
		return c.String(http.StatusOK, "eng")
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/mgmt/proj/test/agt", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, "/api/v1/hosts/test", req.URL.EscapedPath())
		assert.Equal(t, http.StatusOK, rec.Code)

		defer rec.Result().Body.Close()
		bodyBytes, _ := ioutil.ReadAll(rec.Result().Body)
		assert.Equal(t, "hosts", string(bodyBytes))
	}
}

// Issue #1573
func TestEchoRewriteWithCaret(t *testing.T) {
	e := echo.New()

	e.Pre(RewriteWithConfig(RewriteConfig{
		Rules: map[string]string{
			"^/abc/*": "/v1/abc/$1",
		},
	}))

	rec := httptest.NewRecorder()

	var req *http.Request

	req = httptest.NewRequest(http.MethodGet, "/abc/test", nil)
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/v1/abc/test", req.URL.Path)

	req = httptest.NewRequest(http.MethodGet, "/v1/abc/test", nil)
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/v1/abc/test", req.URL.Path)

	req = httptest.NewRequest(http.MethodGet, "/v2/abc/test", nil)
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/v2/abc/test", req.URL.Path)
}

// Verify regex used with rewrite
func TestEchoRewriteWithRegexRules(t *testing.T) {
	e := echo.New()

	e.Pre(RewriteWithConfig(RewriteConfig{
		Rules: map[string]string{
			"^/a/*":     "/v1/$1",
			"^/b/*/c/*": "/v2/$2/$1",
			"^/c/*/*":   "/v3/$2",
		},
		RegexRules: map[*regexp.Regexp]string{
			regexp.MustCompile("^/x/.+?/(.*)"):   "/v4/$1",
			regexp.MustCompile("^/y/(.+?)/(.*)"): "/v5/$2/$1",
		},
	}))

	var rec *httptest.ResponseRecorder
	var req *http.Request

	testCases := []struct {
		requestPath string
		expectPath  string
	}{
		{"/unmatched", "/unmatched"},
		{"/a/test", "/v1/test"},
		{"/b/foo/c/bar/baz", "/v2/bar/baz/foo"},
		{"/c/ignore/test", "/v3/test"},
		{"/c/ignore1/test/this", "/v3/test/this"},
		{"/x/ignore/test", "/v4/test"},
		{"/y/foo/bar", "/v5/bar/foo"},
	}

		for _, tc := range testCases {
			t.Run(tc.requestPath, func(t *testing.T) {
				req = httptest.NewRequest(http.MethodGet, tc.requestPath, nil)
				rec = httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				assert.Equal(t, tc.expectPath, req.URL.EscapedPath())
			})
		}
}

package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	req.URL.Path = "/api/users"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/users", req.URL.EscapedPath())
	req.URL.Path = "/js/main.js"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/public/javascripts/main.js", req.URL.EscapedPath())
	req.URL.Path = "/old"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.EscapedPath())
	req.URL.Path = "/users/jack/orders/1"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jack/order/1", req.URL.EscapedPath())
	req.URL.Path = "/api/new users"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new%20users", req.URL.EscapedPath())
	req.URL.Path = "/users/jill/orders/T%2FcO4lW%2Ft%2FVp%2F"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F", req.URL.EscapedPath())
	req.URL.Path = "/users/jill/orders/%%%%"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Issue #1086
func TestEchoRewritePreMiddleware(t *testing.T) {
	e := echo.New()
	r := e.Router()

	// Rewrite old url to new one
	e.Pre(RewriteWithConfig(RewriteConfig{
		Rules: map[string]string{
			"/old": "/new",
		},
	}))

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

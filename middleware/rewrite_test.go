package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, "/users", req.URL.Path)
	req.URL.Path = "/js/main.js"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/public/javascripts/main.js", req.URL.Path)
	req.URL.Path = "/old"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.Path)
	req.URL.Path = "/users/jack/orders/1"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/user/jack/order/1", req.URL.Path)
	req.URL.Path = "/api/new users"
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new users", req.URL.Path)
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
		return c.NoContent(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.Path)
	assert.Equal(t, 200, rec.Code)
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
		return c.String(200, "hosts")
	})
	r.Add(http.MethodGet, "/api/:version/eng", func(c echo.Context) error {
		return c.String(200, "eng")
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/mgmt/proj/test/agt", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, "/api/v1/hosts/test", req.URL.Path)
		assert.Equal(t, 200, rec.Code)

		defer rec.Result().Body.Close()
		bodyBytes, _ := ioutil.ReadAll(rec.Result().Body)
		assert.Equal(t, "hosts", string(bodyBytes))
	}
}

package middleware

import (
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
	req := httptest.NewRequest(echo.GET, "/", nil)
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
	r.Add(echo.GET, "/new", func(c echo.Context) error {
		return c.NoContent(200)
		return nil
	})

	req := httptest.NewRequest(echo.GET, "/old", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "/new", req.URL.Path)
	assert.Equal(t, 200, rec.Code)
}

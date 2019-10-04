package echo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Fix me
func TestGroup(t *testing.T) {
	g := New().Group("/group")
	h := func(Context) error { return nil }
	g.CONNECT("/", h)
	g.DELETE("/", h)
	g.GET("/", h)
	g.HEAD("/", h)
	g.OPTIONS("/", h)
	g.PATCH("/", h)
	g.POST("/", h)
	g.PUT("/", h)
	g.TRACE("/", h)
	g.Any("/", h)
	g.Match([]string{http.MethodGet, http.MethodPost}, "/", h)
	g.Static("/static", "/tmp")
	g.File("/walle", "_fixture/images//walle.png")
}

func TestGroupRouteMiddleware(t *testing.T) {
	// Ensure middleware slices are not re-used
	e := New()
	g := e.Group("/group")
	h := func(Context) error { return nil }
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m4 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(404)
		}
	}
	m5 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(405)
		}
	}
	g.Use(m1, m2, m3)
	g.GET("/404", h, m4)
	g.GET("/405", h, m5)

	c, _ := request(http.MethodGet, "/group/404", e)
	assert.Equal(t, 404, c)
	c, _ = request(http.MethodGet, "/group/405", e)
	assert.Equal(t, 405, c)
}

func TestGroupRouteMiddlewareWithMatchAny(t *testing.T) {
	// Ensure middleware and match any routes do not conflict
	e := New()
	g := e.Group("/group")
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, c.Path())
		}
	}
	h := func(c Context) error {
		return c.String(http.StatusOK, c.Path())
	}
	g.Use(m1)
	g.GET("/help", h, m2)
	g.GET("/*", h, m2)
	g.GET("", h, m2)
	e.GET("unrelated", h, m2)
	e.GET("*", h, m2)

	_, m := request(http.MethodGet, "/group/help", e)
	assert.Equal(t, "/group/help", m)
	_, m = request(http.MethodGet, "/group/help/other", e)
	assert.Equal(t, "/group/*", m)
	_, m = request(http.MethodGet, "/group/404", e)
	assert.Equal(t, "/group/*", m)
	_, m = request(http.MethodGet, "/group", e)
	assert.Equal(t, "/group", m)
	_, m = request(http.MethodGet, "/other", e)
	assert.Equal(t, "/*", m)
	_, m = request(http.MethodGet, "/", e)
	assert.Equal(t, "/*", m)

}

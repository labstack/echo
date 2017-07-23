package echo

import (
	"testing"

	"net/http"

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
	g.Match([]string{GET, POST}, "/", h)
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

	c, _ := request(GET, "/group/404", e)
	assert.Equal(t, 404, c)
	c, _ = request(GET, "/group/405", e)
	assert.Equal(t, 405, c)
}

func TestGroupMiddlewareNotCalledIfRouteNotInGroup(t *testing.T) {
	// Ensure that group middlewares are not called for routes that are not part of the group

	m := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "BODY")
		}
	}

	e := New()
	g := e.Group("", m)
	g.GET("/route", func(Context) error { return nil })

	_, b := request(GET, "/route", e)
	assert.Equal(t, "BODY", b)

	_, b = request(GET, "/other", e)
	assert.Equal(t, "{\"message\":\"Not Found\"}", b)
}

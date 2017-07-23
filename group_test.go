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

func TestWrapMiddlewareNotCalledForRoutesNotInGroup(t *testing.T) {
	// Ensure that wrap middlewares are not called for routes that are not wrapped

	m := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "Middleware")
		}
	}

	e := New()
	w := e.Wrap(m)
	w.GET("/exists", func(Context) error { return nil })

	_, b := request(GET, "/exists", e)
	assert.Equal(t, "Middleware", b)

	_, b = request(GET, "/does_not_exists", e)
	assert.Equal(t, "{\"message\":\"Not Found\"}", b)
}

func TestGroupMiddlewareCatchAllRoutes(t *testing.T) {
	// Ensure group middlewares are called for all possible sub routes.

	m := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "Middleware")
		}
	}

	e := New()
	g := e.Group("/proxy", m)
	g.GET("/proxy/exists", func(Context) error { return nil })

	_, b := request(GET, "/proxy/exists", e)
	assert.Equal(t, "Middleware", b)

	_, b = request(GET, "/proxy/does_not_exists", e)
	assert.Equal(t, "Middleware", b)
}

func TestWrapWorksInsideAGroup(t *testing.T) {
	// Ensure wrap works properly inside a group

	m := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "Middleware")
		}
	}

	e := New()
	g := e.Group("/prefix")
	g.GET("/route", func(Context) error { return nil })

	w := g.Wrap(m)
	w.GET("/wrapped", func(Context) error { return nil })

	_, b := request(GET, "/prefix/wrapped", e)
	assert.Equal(t, "Middleware", b)

	_, b = request(GET, "/prefix/route", e)
	assert.Equal(t, "", b)

	_, b = request(GET, "/profix/other", e)
	assert.Equal(t, "{\"message\":\"Not Found\"}", b)
}

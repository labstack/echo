package echo

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"reflect"
	"strings"

	"errors"

	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

type (
	user struct {
		ID   string `json:"id" xml:"id"`
		Name string `json:"name" xml:"name"`
	}
)

func TestEcho(t *testing.T) {
	e := New()
	req := test.NewRequest(GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := NewContext(req, rec, e)

	// Router
	assert.NotNil(t, e.Router())

	// Debug
	e.SetDebug(true)
	assert.True(t, e.debug)

	// DefaultHTTPErrorHandler
	e.DefaultHTTPErrorHandler(errors.New("error"), c)
	assert.Equal(t, http.StatusInternalServerError, rec.Status())
}

func TestEchoStatic(t *testing.T) {
	e := New()

	// OK
	e.Static("/images", "_fixture/images")
	c, b := request(GET, "/images/walle.png", e)
	assert.Equal(t, http.StatusOK, c)
	assert.NotEmpty(t, b)

	// No file
	e.Static("/images", "_fixture/scripts")
	c, _ = request(GET, "/images/bolt.png", e)
	assert.Equal(t, http.StatusNotFound, c)

	// Directory
	e.Static("/images", "_fixture/images")
	c, _ = request(GET, "/images", e)
	assert.Equal(t, http.StatusNotFound, c)

	// Directory with index.html
	e.Static("/", "_fixture")
	c, r := request(GET, "/", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, true, strings.HasPrefix(r, "<!doctype html>"))

	// Sub-directory with index.html
	c, r = request(GET, "/folder", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, true, strings.HasPrefix(r, "<!doctype html>"))
}

func TestEchoFile(t *testing.T) {
	e := New()
	e.File("/walle", "_fixture/images/walle.png")
	c, b := request(GET, "/walle", e)
	assert.Equal(t, http.StatusOK, c)
	assert.NotEmpty(t, b)
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Use(MiddlewareFunc(func(h Handler) Handler {
		return HandlerFunc(func(c Context) error {
			buf.WriteString("a")
			return h.Handle(c)
		})
	}))

	e.Use(MiddlewareFunc(func(h Handler) Handler {
		return HandlerFunc(func(c Context) error {
			buf.WriteString("b")
			return h.Handle(c)
		})
	}))

	e.Use(MiddlewareFunc(func(h Handler) Handler {
		return HandlerFunc(func(c Context) error {
			buf.WriteString("c")
			return h.Handle(c)
		})
	}))

	// Route
	e.Get("/", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	}))

	c, b := request(GET, "/", e)
	assert.Equal(t, "abc", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)

	// Error
	e.Use(MiddlewareFunc(func(Handler) Handler {
		return HandlerFunc(func(c Context) error {
			return errors.New("error")
		})
	}))
	c, b = request(GET, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.Get("/ok", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	}))

	c, b := request(GET, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoConnect(t *testing.T) {
	e := New()
	testMethod(t, CONNECT, "/", e)
}

func TestEchoDelete(t *testing.T) {
	e := New()
	testMethod(t, DELETE, "/", e)
}

func TestEchoGet(t *testing.T) {
	e := New()
	testMethod(t, GET, "/", e)
}

func TestEchoHead(t *testing.T) {
	e := New()
	testMethod(t, HEAD, "/", e)
}

func TestEchoOptions(t *testing.T) {
	e := New()
	testMethod(t, OPTIONS, "/", e)
}

func TestEchoPatch(t *testing.T) {
	e := New()
	testMethod(t, PATCH, "/", e)
}

func TestEchoPost(t *testing.T) {
	e := New()
	testMethod(t, POST, "/", e)
}

func TestEchoPut(t *testing.T) {
	e := New()
	testMethod(t, PUT, "/", e)
}

func TestEchoTrace(t *testing.T) {
	e := New()
	testMethod(t, TRACE, "/", e)
}

func TestEchoAny(t *testing.T) { // JFC
	e := New()
	e.Any("/", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "Any")
	}))
}

func TestEchoMatch(t *testing.T) { // JFC
	e := New()
	e.Match([]string{GET, POST}, "/", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "Match")
	}))
}

func TestEchoURL(t *testing.T) {
	e := New()
	static := HandlerFunc(func(Context) error { return nil })
	getUser := HandlerFunc(func(Context) error { return nil })
	getFile := HandlerFunc(func(Context) error { return nil })

	e.Get("/static/file", static)
	e.Get("/users/:id", getUser)
	g := e.Group("/group")
	g.Get("/users/:uid/files/:fid", getFile)

	assert.Equal(t, "/static/file", e.URL(static))
	assert.Equal(t, "/users/:id", e.URL(getUser))
	assert.Equal(t, "/users/1", e.URL(getUser, "1"))
	assert.Equal(t, "/group/users/1/files/:fid", e.URL(getFile, "1"))
	assert.Equal(t, "/group/users/1/files/1", e.URL(getFile, "1", "1"))
}

func TestEchoRoutes(t *testing.T) {
	e := New()
	routes := []Route{
		{GET, "/users/:user/events", ""},
		{GET, "/users/:user/events/public", ""},
		{POST, "/repos/:owner/:repo/git/refs", ""},
		{POST, "/repos/:owner/:repo/git/tags", ""},
	}
	for _, r := range routes {
		e.add(r.Method, r.Path, HandlerFunc(func(c Context) error {
			return c.String(http.StatusOK, "OK")
		}))
	}

	for i, r := range e.Routes() {
		assert.Equal(t, routes[i].Method, r.Method)
		assert.Equal(t, routes[i].Path, r.Path)
	}
}

func TestEchoGroup(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	e.Use(MiddlewareFunc(func(h Handler) Handler {
		return HandlerFunc(func(c Context) error {
			buf.WriteString("0")
			return h.Handle(c)
		})
	}))
	h := HandlerFunc(func(c Context) error {
		return c.NoContent(http.StatusOK)
	})

	//--------
	// Routes
	//--------

	e.Get("/users", h)

	// Group
	g1 := e.Group("/group1")
	g1.Use(MiddlewareFunc(func(h Handler) Handler {
		return HandlerFunc(func(c Context) error {
			buf.WriteString("1")
			return h.Handle(c)
		})
	}))
	g1.Get("", h)

	// Nested groups
	g2 := e.Group("/group2")
	g3 := g2.Group("/group3")
	g3.Get("", h)

	request(GET, "/users", e)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request(GET, "/group1", e)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	c, _ := request(GET, "/group2/group3", e)
	assert.Equal(t, http.StatusOK, c)
}

func TestEchoNotFound(t *testing.T) {
	e := New()
	req := test.NewRequest(GET, "/files", nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(req, rec)
	assert.Equal(t, http.StatusNotFound, rec.Status())
}

func TestEchoMethodNotAllowed(t *testing.T) {
	e := New()
	e.Get("/", HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, "Echo!")
	}))
	req := test.NewRequest(POST, "/", nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(req, rec)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Status())
}

func TestEchoHTTPError(t *testing.T) {
	m := http.StatusText(http.StatusBadRequest)
	he := NewHTTPError(http.StatusBadRequest, m)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, m, he.Error())
}

func testMethod(t *testing.T, method, path string, e *Echo) {
	m := fmt.Sprintf("%c%s", method[0], strings.ToLower(method[1:]))
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(HandlerFunc(func(c Context) error {
		return c.String(http.StatusOK, method)
	}))
	i := interface{}(e)
	reflect.ValueOf(i).MethodByName(m).Call([]reflect.Value{p, h})
	_, body := request(method, path, e)
	if body != method {
		t.Errorf("expected body `%s`, got %s.", method, body)
	}
}

func request(method, path string, e *Echo) (int, string) {
	req := test.NewRequest(method, path, nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(req, rec)
	return rec.Status(), rec.Body.String()
}

package echo

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"reflect"
	"strings"

	"errors"

	"github.com/stretchr/testify/assert"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id"`
		Name string `json:"name" xml:"name" form:"name"`
	}
)

const (
	userJSON       = `{"id":1,"name":"Jon Snow"}`
	userXML        = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm       = `id=1&name=Jon Snow`
	invalidContent = "invalid content"
)

func TestEcho(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Router
	assert.NotNil(t, e.Router())

	// DefaultHTTPErrorHandler
	e.DefaultHTTPErrorHandler(errors.New("error"), c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
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

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	// Route
	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(GET, "/", e)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return errors.New("error")
		}
	})
	e.GET("/", NotFoundHandler)
	c, _ := request(GET, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(GET, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoWrapHandler(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestEchoWrapMiddleware(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.Write([]byte("mw"))
			h.ServeHTTP(w, r)
		})
	})
	h := mw(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
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
	e.Any("/", func(c Context) error {
		return c.String(http.StatusOK, "Any")
	})
}

func TestEchoMatch(t *testing.T) { // JFC
	e := New()
	e.Match([]string{GET, POST}, "/", func(c Context) error {
		return c.String(http.StatusOK, "Match")
	})
}

func TestEchoURL(t *testing.T) {
	e := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	e.GET("/static/file", static)
	e.GET("/users/:id", getUser)
	g := e.Group("/group")
	g.GET("/users/:uid/files/:fid", getFile)

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
		e.add(r.Method, r.Path, func(c Context) error {
			return c.String(http.StatusOK, "OK")
		})
	}

	for _, r := range e.Routes() {
		found := false
		for _, rr := range routes {
			if r.Method == rr.Method && r.Path == rr.Path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Route %s : %s not found", r.Method, r.Path)
		}
	}
}

func TestEchoGroup(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	e.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))
	h := func(c Context) error {
		return c.NoContent(http.StatusOK)
	}

	//--------
	// Routes
	//--------

	e.GET("/users", h)

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.GET("", h)

	// Nested groups with middleware
	g2 := e.Group("/group2")
	g2.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})
	g3 := g2.Group("/group3")
	g3.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})
	g3.GET("", h)

	request(GET, "/users", e)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request(GET, "/group1", e)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request(GET, "/group2/group3", e)
	assert.Equal(t, "023", buf.String())
}

func TestEchoNotFound(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/files", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEchoMethodNotAllowed(t *testing.T) {
	e := New()
	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "Echo!")
	})
	req, _ := http.NewRequest(POST, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestEchoHTTPError(t *testing.T) {
	m := http.StatusText(http.StatusBadRequest)
	he := NewHTTPError(http.StatusBadRequest, m)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, m, he.Error())
}

func TestEchoContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(context), c)
	e.ReleaseContext(c)
}

func TestEchoStart(t *testing.T) {
	e := New()
	go func() {
		assert.NoError(t, e.Start(":0"))
	}()
	time.Sleep(200 * time.Millisecond)
	e.Shutdown(1 * time.Second)
}

func TestEchoStartTLS(t *testing.T) {
	e := New()
	go func() {
		assert.NoError(t, e.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem"))
	}()
	time.Sleep(200 * time.Millisecond)
	e.ShutdownTLS(1 * time.Second)
}

func testMethod(t *testing.T, method, path string, e *Echo) {
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(func(c Context) error {
		return c.String(http.StatusOK, method)
	})
	i := interface{}(e)
	reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{p, h})
	_, body := request(method, path, e)
	if body != method {
		t.Errorf("expected body `%s`, got %s", method, body)
	}
}

func request(method, path string, e *Echo) (int, string) {
	req, _ := http.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

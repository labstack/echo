package echo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"reflect"
	"strings"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

type (
	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

// TODO: Improve me!
func TestEchoMaxParam(t *testing.T) {
	e := New()
	e.SetMaxParam(8)
	if e.maxParam != 8 {
		t.Errorf("max param should be 8, found %d", e.maxParam)
	}
}

func TestEchoIndex(t *testing.T) {
	e := New()
	e.Index("examples/website/public/index.html")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/", nil)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoFavicon(t *testing.T) {
	e := New()
	e.Favicon("examples/website/public/favicon.ico")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/favicon.ico", nil)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoStatic(t *testing.T) {
	e := New()
	e.Static("/scripts", "examples/website/public/scripts")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/scripts/main.js", nil)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	b := new(bytes.Buffer)

	// echo.MiddlewareFunc
	e.Use(MiddlewareFunc(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			b.WriteString("a")
			return h(c)
		}
	}))

	// func(echo.HandlerFunc) echo.HandlerFunc
	e.Use(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			b.WriteString("b")
			return h(c)
		}
	})

	// echo.HandlerFunc
	e.Use(HandlerFunc(func(c *Context) error {
		b.WriteString("c")
		return nil
	}))

	// func(*echo.Context) error
	e.Use(func(c *Context) error {
		b.WriteString("d")
		return nil
	})

	// func(http.Handler) http.Handler
	e.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b.WriteString("e")
			h.ServeHTTP(w, r)
		})
	})

	// http.Handler
	e.Use(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("f")
	})))

	// http.HandlerFunc
	e.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("g")
	}))

	// func(http.ResponseWriter, *http.Request)
	e.Use(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("h")
	})

	// Unknown
	assert.Panics(t, func() {
		e.Use(nil)
	})

	// Route
	e.Get("/hello", func(c *Context) error {
		return c.String(http.StatusOK, "world")
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/hello", nil)
	e.ServeHTTP(w, r)
	if b.String() != "abcdefgh" {
		t.Errorf("buffer should be abcdefgh, found %s", b.String())
	}
	if w.Body.String() != "world" {
		t.Error("body should be world")
	}
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.Get("/1", HandlerFunc(func(c *Context) error {
		return c.String(http.StatusOK, "1")
	}))
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/1", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "1" {
		t.Error("body should be 1")
	}

	// func(*echo.Context) error
	e.Get("/2", func(c *Context) error {
		return c.String(http.StatusOK, "2")
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/2", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "2" {
		t.Error("body should be 2")
	}

	// http.Handler/http.HandlerFunc
	e.Get("/3", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("3"))
	}))
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/3", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "3" {
		t.Error("body should be 3")
	}

	// func(http.ResponseWriter, *http.Request)
	e.Get("/4", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("4"))
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/4", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "4" {
		t.Error("body should be 4")
	}

	// Unknown
	assert.Panics(t, func() {
		e.Get("/5", nil)
	})
}

func TestEchoGroup(t *testing.T) {
	b := new(bytes.Buffer)
	e := New()
	e.Use(func(*Context) error {
		b.WriteString("1")
		return nil
	})
	e.Get("/users", func(*Context) error { return nil })
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/users", nil)
	e.ServeHTTP(w, r)
	if b.String() != "1" {
		t.Errorf("should only execute middleware 1, executed %s", b.String())
	}

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(*Context) error {
		b.WriteString("2")
		return nil
	})
	g1.Get("/home", func(*Context) error { return nil })
	b.Reset()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/group1/home", nil)
	e.ServeHTTP(w, r)
	if b.String() != "12" {
		t.Errorf("should execute middleware 1 & 2, executed %s", b.String())
	}

	// Group with no parent middleware
	g2 := e.Group("/group2", func(*Context) error {
		b.WriteString("3")
		return nil
	})
	g2.Get("/home", func(*Context) error { return nil })
	b.Reset()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/group2/home", nil)
	e.ServeHTTP(w, r)
	if b.String() != "3" {
		t.Errorf("should execute middleware 3, executed %s", b.String())
	}

	// Nested group
	g3 := e.Group("/group3")
	g4 := g3.Group("/group4")
	g4.Get("/home", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/group3/group4/home", nil)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoConnect(t *testing.T) {
	e := New()
	testMethod(t, e, nil, CONNECT, "/")
}

func TestEchoDelete(t *testing.T) {
	e := New()
	testMethod(t, e, nil, DELETE, "/")
}

func TestEchoGet(t *testing.T) {
	e := New()
	testMethod(t, e, nil, GET, "/")
}

func TestEchoHead(t *testing.T) {
	e := New()
	testMethod(t, e, nil, HEAD, "/")
}

func TestEchoOptions(t *testing.T) {
	e := New()
	testMethod(t, e, nil, OPTIONS, "/")
}

func TestEchoPatch(t *testing.T) {
	e := New()
	testMethod(t, e, nil, PATCH, "/")
}

func TestEchoPost(t *testing.T) {
	e := New()
	testMethod(t, e, nil, POST, "/")
}

func TestEchoPut(t *testing.T) {
	e := New()
	testMethod(t, e, nil, PUT, "/")
}

func TestEchoTrace(t *testing.T) {
	e := New()
	testMethod(t, e, nil, TRACE, "/")
}

func testMethod(t *testing.T, e *Echo, g *Group, method, path string) {
	m := fmt.Sprintf("%c%s", method[0], strings.ToLower(method[1:]))
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(func(c *Context) error {
		c.String(http.StatusOK, method)
		return nil
	})
	i := interface{}(e)
	if g != nil {
		path = e.prefix + path
		i = g
	}
	reflect.ValueOf(i).MethodByName(m).Call([]reflect.Value{p, h})
	_, body := request(method, path, e)
	if body != method {
		t.Errorf("expected body `%s`, got %s.", method, body)
	}
}

func request(method, path string, e *Echo) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

func TestWebSocket(t *testing.T) {
	e := New()
	e.WebSocket("/ws", func(c *Context) error {
		c.socket.Write([]byte("test"))
		return nil
	})
	srv := httptest.NewServer(e)
	defer srv.Close()
	addr := srv.Listener.Addr().String()
	origin := "http://localhost"
	url := fmt.Sprintf("ws://%s/ws", addr)
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	ws.Write([]byte("test"))
	defer ws.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(ws)
	s := buf.String()
	if s != "test" {
		t.Errorf("expected `test`, got %s.", s)
	}
}

func TestEchoURL(t *testing.T) {
	e := New()
	static := func(*Context) error { return nil }
	getUser := func(*Context) error { return nil }
	getFile := func(*Context) error { return nil }
	e.Get("/static/file", static)
	e.Get("/users/:id", getUser)
	e.Get("/users/:uid/files/:fid", getFile)

	if e.URL(static) != "/static/file" {
		t.Error("uri should be /static/file")
	}
	if e.URI(static) != "/static/file" {
		t.Error("uri should be /static/file")
	}
	if e.URI(getUser) != "/users/:id" {
		t.Error("uri should be /users/:id")
	}
	if e.URI(getUser, "1") != "/users/1" {
		t.Error("uri should be /users/1")
	}
	if e.URI(getFile, "1") != "/users/1/files/:fid" {
		t.Error("uri should be /users/1/files/:fid")
	}
	if e.URI(getFile, "1", "1") != "/users/1/files/1" {
		t.Error("uri should be /users/1/files/1")
	}
}

func TestEchoNotFound(t *testing.T) {
	e := New()

	// Default NotFound handler
	r, _ := http.NewRequest(GET, "/files", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("status code should be 404, found %d", w.Code)
	}
}

//func verifyUser(u2 *user, t *testing.T) {
//	if u2.ID != u1.ID {
//		t.Errorf("user id should be %s, found %s", u1.ID, u2.ID)
//	}
//	if u2.Name != u1.Name {
//		t.Errorf("user name should be %s, found %s", u1.Name, u2.Name)
//	}
//}

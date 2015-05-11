package echo

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

var u1 = user{
	ID:   "1",
	Name: "Joe",
}

// TODO: Improve me!
func TestEchoMaxParam(t *testing.T) {
	e := New()
	e.MaxParam(8)
	if e.maxParam != 8 {
		t.Errorf("max param should be 8, found %d", e.maxParam)
	}
}

func TestEchoIndex(t *testing.T) {
	e := New()
	e.Index("examples/web/public/index.html")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoStatic(t *testing.T) {
	e := New()
	e.Static("/scripts", "examples/web/public/scripts")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/scripts/main.js", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	b := new(bytes.Buffer)

	// MiddlewareFunc
	e.Use(MiddlewareFunc(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) *HTTPError {
			b.WriteString("a")
			return h(c)
		}
	}))

	// func(echo.HandlerFunc) (echo.HandlerFunc, error)
	e.Use(func(h HandlerFunc) HandlerFunc {
		return func(c *Context) *HTTPError {
			b.WriteString("b")
			return h(c)
		}
	})

	// func(*echo.Context) *HTTPError
	e.Use(func(c *Context) *HTTPError {
		b.WriteString("c")
		return nil
	})

	// func(*echo.Context)
	e.Use(func(c *Context) {
		b.WriteString("d")
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

	// func(http.ResponseWriter, *http.Request) *HTTPError
	e.Use(func(w http.ResponseWriter, r *http.Request) *HTTPError {
		b.WriteString("i")
		return nil
	})

	// Route
	e.Get("/hello", func(c *Context) {
		c.String(http.StatusOK, "world")
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/hello", nil)
	e.ServeHTTP(w, r)
	if b.String() != "abcdefghi" {
		t.Errorf("buffer should be abcdefghi, found %s", b.String())
	}
	if w.Body.String() != "world" {
		t.Error("body should be world")
	}
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.Get("/1", HandlerFunc(func(c *Context) *HTTPError {
		return c.String(http.StatusOK, "1")
	}))
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/1", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "1" {
		t.Error("body should be 1")
	}

	// func(*echo.Context) *HTTPError
	e.Get("/2", func(c *Context) *HTTPError {
		return c.String(http.StatusOK, "2")
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/2", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "2" {
		t.Error("body should be 2")
	}

	// func(*echo.Context)
	e.Get("/3", func(c *Context) {
		c.String(http.StatusOK, "3")
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/3", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "3" {
		t.Error("body should be 3")
	}

	// http.Handler/http.HandlerFunc
	e.Get("/4", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("4"))
	}))
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/4", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "4" {
		t.Error("body should be 4")
	}

	// func(http.ResponseWriter, *http.Request)
	e.Get("/5", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("5"))
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/5", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "5" {
		t.Error("body should be 5")
	}

	// func(http.ResponseWriter, *http.Request) *HTTPError
	e.Get("/6", func(w http.ResponseWriter, r *http.Request) *HTTPError {
		w.Write([]byte("6"))
		return nil
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/6", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "6" {
		t.Error("body should be 6")
	}
}

func TestEchoGroup(t *testing.T) {
	b := new(bytes.Buffer)
	e := New()
	e.Use(func(*Context) {
		b.WriteString("1")
	})
	e.Get("/users", func(*Context) {})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(GET, "/users", nil)
	e.ServeHTTP(w, r)
	if b.String() != "1" {
		t.Errorf("should only execute middleware 1, executed %s", b.String())
	}

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(*Context) {
		b.WriteString("2")
	})
	g1.Get("/home", func(*Context) {})
	b.Reset()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/group1/home", nil)
	e.ServeHTTP(w, r)
	if b.String() != "12" {
		t.Errorf("should execute middleware 1 & 2, executed %s", b.String())
	}

	// Group with no parent middleware
	g2 := e.Group("/group2", func(*Context) {
		b.WriteString("3")
	})
	g2.Get("/home", func(*Context) {})
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
	g4.Get("/home", func(c *Context) {
		c.NoContent(http.StatusOK)
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(GET, "/group3/group4/home", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMethod(t *testing.T) {
	e := New()
	h := func(*Context) {}
	e.Connect("/", h)
	e.Delete("/", h)
	e.Get("/", h)
	e.Head("/", h)
	e.Options("/", h)
	e.Patch("/", h)
	e.Post("/", h)
	e.Put("/", h)
	e.Trace("/", h)
}

func TestEchoURL(t *testing.T) {
	e := New()
	static := func(*Context) {}
	getUser := func(*Context) {}
	getFile := func(*Context) {}
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

	// Customized NotFound handler
	e.NotFoundHandler(func(c *Context) {
		c.String(http.StatusNotFound, "not found")
	})
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Body.String() != "not found" {
		t.Errorf("body should be `not found`")
	}
}

func verifyUser(u2 *user, t *testing.T) {
	if u2.ID != u1.ID {
		t.Errorf("user id should be %s, found %s", u1.ID, u2.ID)
	}
	if u2.Name != u1.Name {
		t.Errorf("user name should be %s, found %s", u1.Name, u2.Name)
	}
}

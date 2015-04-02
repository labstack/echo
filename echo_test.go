package echo

import (
	"bytes"
	"encoding/json"
	"io"
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

var u = user{
	ID:   "1",
	Name: "Joe",
}

// TODO: Fix me
func TestEchoMaxParam(t *testing.T) {
	e := New()
	e.MaxParam(8)
	if e.maxParam != 8 {
		t.Errorf("max param should be 8, found %d", e.maxParam)
	}
}

func TestEchoIndex(t *testing.T) {
	e := New()
	e.Index("example/public/index.html")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoStatic(t *testing.T) {
	e := New()
	e.Static("/js", "example/public/js")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/js/main.js", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	b := new(bytes.Buffer)

	// func(*echo.Context)
	e.Use(func(c *Context) {
		b.WriteString("a")
	})

	// func(echo.HandlerFunc) echo.HandlerFunc
	e.Use(func(h HandlerFunc) HandlerFunc {
		return HandlerFunc(func(c *Context) {
			b.WriteString("b")
			h(c)
		})
	})

	// http.HandlerFunc
	e.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("c")
	}))

	// http.Handler
	e.Use(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("d")
	})))

	// func(http.Handler) http.Handler
	e.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b.WriteString("e")
			h.ServeHTTP(w, r)
		})
	})

	// func(http.ResponseWriter, *http.Request)
	e.Use(func(w http.ResponseWriter, r *http.Request) {
		b.WriteString("f")
	})

	// Route
	e.Get("/hello", func(c *Context) {
		c.String(200, "world")
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/hello", nil)
	e.ServeHTTP(w, r)
	if b.String() != "abcdef" {
		t.Errorf("buffer should be abcdef, found %s", b.String())
	}
	if w.Body.String() != "world" {
		t.Error("body should be world")
	}
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// func(*echo.Context)
	e.Get("/1", func(c *Context) {
		c.String(http.StatusOK, "1")
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/1", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "1" {
		t.Error("body should be 1")
	}

	// http.Handler/http.HandlerFunc
	e.Get("/2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
	}))
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(MethodGET, "/2", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "2" {
		t.Error("body should be 2")
	}

	// func(http.ResponseWriter, *http.Request)
	e.Get("/3", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("3"))
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(MethodGET, "/3", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "3" {
		t.Error("body should be 3")
	}
}

func TestEchoSubGroup(t *testing.T) {
	b := new(bytes.Buffer)

	e := New()
	e.Use(func(*Context) {
		b.WriteString("1")
	})
	e.Get("/users", func(*Context) {})

	s := e.Sub("/sub")
	s.Use(func(*Context) {
		b.WriteString("2")
	})
	s.Get("/home", func(*Context) {})

	g := e.Group("/group")
	g.Use(func(*Context) {
		b.WriteString("3")
	})
	g.Get("/home", func(*Context) {})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/users", nil)
	e.ServeHTTP(w, r)
	if b.String() != "1" {
		t.Errorf("should only execute middleware 1, executed %s", b.String())
	}

	b.Reset()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(MethodGET, "/sub/home", nil)
	e.ServeHTTP(w, r)
	if b.String() != "12" {
		t.Errorf("should execute middleware 1 & 2, executed %s", b.String())
	}

	b.Reset()
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(MethodGET, "/group/home", nil)
	e.ServeHTTP(w, r)
	if b.String() != "3" {
		t.Errorf("should execute middleware 3, executed %s", b.String())
	}
}

func TestEchoMethod(t *testing.T) {
	// e := New()
	// // GET
	// e.Get("/users", func(c *Context) {})
	// h, _, _ := e.Router.Find("GET", "/users")
	// if h == nil {
	// 	t.Error("should find route for GET")
	// }
}

func TestEchoServeHTTP(t *testing.T) {
	e := New()

	// OK
	e.Get("/users", func(*Context) {
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodGET, "/users", nil)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status code should be 200, found %d", w.Code)
	}

	// NotFound
	r, _ = http.NewRequest(MethodGET, "/user", nil)
	w = httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("status code should be 404, found %d", w.Code)
	}

	// NotAllowed
	// r, _ = http.NewRequest("POST", "/users", nil)
	// w = httptest.NewRecorder()
	// e.ServeHTTP(w, r)
	// if w.Code != http.StatusMethodNotAllowed {
	// 	t.Errorf("status code should be 405, found %d", w.Code)
	// }
}

func verifyUser(rd io.Reader, t *testing.T) {
	u2 := new(user)
	dec := json.NewDecoder(rd)
	err := dec.Decode(u2)
	if err != nil {
		t.Error(err)
	}
	if u2.ID != u.ID {
		t.Errorf("user id should be %s, found %s", u.ID, u2.ID)
	}
	if u2.Name != u.Name {
		t.Errorf("user name should be %s, found %s", u.Name, u2.Name)
	}
}

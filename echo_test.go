package echo

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	user struct {
		ID   string
		Name string
	}
)

var u = user{
	ID:   "1",
	Name: "Joe",
}

func TestEchoMaxParam(t *testing.T) {
	b := New()
	b.MaxParam(8)
	if b.maxParam != 8 {
		t.Errorf("max param should be 8, found %d", b.maxParam)
	}
}

func TestEchoIndex(t *testing.T) {
	b := New()
	b.Index("example/public/index.html")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoStatic(t *testing.T) {
	b := New()
	b.Static("/js", "example/public/js")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/js/main.js", nil)
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	b := New()

	// func(HandlerFunc) HandlerFunc
	b.Use(func(h HandlerFunc) HandlerFunc {
		return HandlerFunc(func(c *Context) {
			c.Request.Header.Set("a", "1")
			h(c)
		})
	})

	// http.HandlerFunc
	b.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("b", "2")
	}))

	// http.Handler
	b.Use(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("c", "3")
	})))

	// func(http.Handler) http.Handler
	b.Use(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("d", "4")
		})
	})

	// Route
	b.Get("/users", func(c *Context) {
		h := c.Request.Header.Get("a")
		if h != "1" {
			t.Errorf("header a should be 1, found %s", h)
		}
		h = c.Request.Header.Get("b")
		if h != "2" {
			t.Errorf("header b should be 2, found %s", h)
		}
		h = c.Request.Header.Get("c")
		if h != "3" {
			t.Errorf("header c should be 3, found %s", h)
		}
		h = c.Request.Header.Get("d")
		if h != "4" {
			t.Errorf("header d should be 4, found %s", h)
		}
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/users", nil)
	b.ServeHTTP(w, r)
}

func verifyUser(rd io.Reader, t *testing.T) {
	var l int64
	err := binary.Read(rd, binary.BigEndian, &l) // Body length
	if err != nil {
		t.Fatal(err)
	}
	bd := io.LimitReader(rd, l) // Body
	u2 := new(user)
	dec := json.NewDecoder(bd)
	err = dec.Decode(u2)
	if err != nil {
		t.Fatal(err)
	}
	if u2.ID != u.ID {
		t.Errorf("user id should be %s, found %s", u.ID, u2.ID)
	}
	if u2.Name != u.Name {
		t.Errorf("user name should be %s, found %s", u.Name, u2.Name)
	}
}

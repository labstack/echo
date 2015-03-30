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
	r, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoStatic(t *testing.T) {
	e := New()
	e.Static("/js", "example/public/js")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/js/main.js", nil)
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status code should be 200, found %d", w.Code)
	}
}

func TestEchoMiddleware(t *testing.T) {
	e := New()

	// func(*echo.Context)
	e.Use(func(c *Context) {
		c.Request.Header.Set("e", "5")
	})

	// func(echo.HandlerFunc) echo.HandlerFunc
	e.Use(func(h HandlerFunc) HandlerFunc {
		return HandlerFunc(func(c *Context) {
			c.Request.Header.Set("a", "1")
			h(c)
		})
	})

	// http.HandlerFunc
	e.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("b", "2")
	}))

	// http.Handler
	e.Use(http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("c", "3")
	})))

	// func(http.Handler) http.Handler
	e.Use(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("d", "4")
		})
	})

	// func(http.ResponseWriter, *http.Request)
	e.Use(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("f", "6")
	})

	// Route
	e.Get("/hello", func(c *Context) {
		if c.Request.Header.Get("a") != "1" {
			t.Error("header a should be 1")
		}
		if c.Request.Header.Get("b") != "2" {
			t.Error("header b should be 2")
		}
		if c.Request.Header.Get("c") != "3" {
			t.Error("header c should be 3")
		}
		if c.Request.Header.Get("d") != "4" {
			t.Error("header d should be 4")
		}
		if c.Request.Header.Get("e") != "5" {
			t.Error("header e should be 5")
		}
		if c.Request.Header.Get("f") != "6" {
			t.Error("header f should be 6")
		}
		c.String(200, "world")
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "world" {
		t.Errorf("body should be world")
	}
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// func(*echo.Context)
	e.Get("/1", func(c *Context) {
		c.String(http.StatusOK, "1")
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/1", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "1" {
		t.Errorf("body should be 1")
	}

	// http.Handler / http.HandlerFunc
	e.Get("/2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
	}))
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/2", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "2" {
		t.Errorf("body should be 2")
	}

	// func(http.ResponseWriter, *http.Request)
	e.Get("/3", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("3"))
	})
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/3", nil)
	e.ServeHTTP(w, r)
	if w.Body.String() != "3" {
		t.Errorf("body should be 3")
	}
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

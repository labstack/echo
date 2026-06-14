// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// nopResponseWriter is a no-op http.ResponseWriter used to isolate framework
// overhead from the cost of httptest.ResponseRecorder buffering.
type nopResponseWriter struct{ h http.Header }

func (w *nopResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *nopResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *nopResponseWriter) WriteHeader(int)             {}

func benchServe(b *testing.B, e *Echo, req *http.Request) {
	b.Helper()
	w := &nopResponseWriter{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.h = nil
		e.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTP_Static(b *testing.B) {
	e := New()
	e.GET("/users/profile", func(c *Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	benchServe(b, e, req)
}

func BenchmarkServeHTTP_Param(b *testing.B) {
	e := New()
	e.GET("/users/:id/books/:bid", func(c *Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/users/42/books/7", nil)
	benchServe(b, e, req)
}

// Exercises the global middleware chain (finding #1). Five pass-through middlewares.
func BenchmarkServeHTTP_Middleware(b *testing.B) {
	e := New()
	for i := 0; i < 5; i++ {
		e.Use(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error { return next(c) }
		})
	}
	e.GET("/users/:id", func(c *Context) error { return c.NoContent(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	benchServe(b, e, req)
}

func BenchmarkServeHTTP_String(b *testing.B) {
	e := New()
	e.GET("/", func(c *Context) error { return c.String(http.StatusOK, "Hello, World!") })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	benchServe(b, e, req)
}

func BenchmarkServeHTTP_JSON(b *testing.B) {
	e := New()
	type payload struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Tags []string
	}
	p := payload{ID: 1, Name: "Jon Snow", Tags: []string{"a", "b", "c"}}
	e.GET("/", func(c *Context) error { return c.JSON(http.StatusOK, p) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	benchServe(b, e, req)
}

// Exercises a per-request Set (as request_id/auth middleware do), measuring the store-map reuse.
func BenchmarkServeHTTP_Store(b *testing.B) {
	e := New()
	e.GET("/", func(c *Context) error {
		c.Set("user", "alice")
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	benchServe(b, e, req)
}

func BenchmarkContext_GetSet(b *testing.B) {
	e := New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), &nopResponseWriter{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", i)
		_ = c.Get("key")
	}
}

type bindTarget struct {
	ID     int    `json:"id" query:"id"`
	Name   string `json:"name" query:"name"`
	Email  string `json:"email" query:"email"`
	Age    int    `json:"age" query:"age"`
	Active bool   `json:"active" query:"active"`
}

func BenchmarkBind_JSON(b *testing.B) {
	e := New()
	body := `{"id":1,"name":"Jon Snow","email":"jon@winterfell.north","age":24,"active":true}`
	e.POST("/", func(c *Context) error {
		var t bindTarget
		return c.Bind(&t)
	})
	b.ReportAllocs()
	b.ResetTimer()
	w := &nopResponseWriter{}
	for i := 0; i < b.N; i++ {
		w.h = nil
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(HeaderContentType, MIMEApplicationJSON)
		e.ServeHTTP(w, req)
	}
}

func BenchmarkBind_Query(b *testing.B) {
	e := New()
	e.GET("/", func(c *Context) error {
		var t bindTarget
		return c.Bind(&t)
	})
	b.ReportAllocs()
	b.ResetTimer()
	w := &nopResponseWriter{}
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon&email=jon@x.io&age=24&active=true", nil)
	for i := 0; i < b.N; i++ {
		w.h = nil
		e.ServeHTTP(w, req)
	}
}

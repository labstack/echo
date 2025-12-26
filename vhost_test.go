// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVirtualHostHandler(t *testing.T) {
	okHandler := func(c *Context) error { return c.String(http.StatusOK, http.StatusText(http.StatusOK)) }
	teapotHandler := func(c *Context) error { return c.String(http.StatusTeapot, http.StatusText(http.StatusTeapot)) }
	acceptHandler := func(c *Context) error { return c.String(http.StatusAccepted, http.StatusText(http.StatusAccepted)) }
	teapotMiddleware := MiddlewareFunc(func(next HandlerFunc) HandlerFunc { return teapotHandler })

	ok := New()
	ok.GET("/", okHandler)
	ok.GET("/foo", okHandler)

	teapot := New()
	teapot.GET("/", teapotHandler)
	teapot.GET("/foo", teapotHandler)

	middle := New()
	middle.Use(teapotMiddleware)
	middle.GET("/", okHandler)
	middle.GET("/foo", okHandler)

	virtualHosts := NewVirtualHostHandler(map[string]*Echo{
		"ok.com":         ok,
		"teapot.com":     teapot,
		"middleware.com": middle,
	})
	virtualHosts.GET("/", acceptHandler)
	virtualHosts.GET("/foo", acceptHandler)

	var testCases = []struct {
		name         string
		whenHost     string
		whenPath     string
		expectBody   string
		expectStatus int
	}{
		{
			name:         "No Host Root",
			whenHost:     "",
			whenPath:     "/",
			expectBody:   http.StatusText(http.StatusAccepted),
			expectStatus: http.StatusAccepted,
		},
		{
			name:         "No Host Foo",
			whenHost:     "",
			whenPath:     "/foo",
			expectBody:   http.StatusText(http.StatusAccepted),
			expectStatus: http.StatusAccepted,
		},
		{
			name:         "OK Host Root",
			whenHost:     "ok.com",
			whenPath:     "/",
			expectBody:   http.StatusText(http.StatusOK),
			expectStatus: http.StatusOK,
		},
		{
			name:         "OK Host Foo",
			whenHost:     "ok.com",
			whenPath:     "/foo",
			expectBody:   http.StatusText(http.StatusOK),
			expectStatus: http.StatusOK,
		},
		{
			name:         "Teapot Host Root",
			whenHost:     "teapot.com",
			whenPath:     "/",
			expectBody:   http.StatusText(http.StatusTeapot),
			expectStatus: http.StatusTeapot,
		},
		{
			name:         "Teapot Host Foo",
			whenHost:     "teapot.com",
			whenPath:     "/foo",
			expectBody:   http.StatusText(http.StatusTeapot),
			expectStatus: http.StatusTeapot,
		},
		{
			name:         "Middleware Host",
			whenHost:     "middleware.com",
			whenPath:     "/",
			expectBody:   http.StatusText(http.StatusTeapot),
			expectStatus: http.StatusTeapot,
		},
		{
			name:         "Middleware Host Foo",
			whenHost:     "middleware.com",
			whenPath:     "/foo",
			expectBody:   http.StatusText(http.StatusTeapot),
			expectStatus: http.StatusTeapot,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenPath, nil)
			req.Host = tc.whenHost
			rec := httptest.NewRecorder()

			virtualHosts.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.Equal(t, tc.expectBody, rec.Body.String())
		})
	}
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"bytes"
	stdContext "context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type user struct {
	ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id" header:"id"`
	Name string `json:"name" xml:"name" form:"name" query:"name" param:"name" header:"name"`
}

const (
	userJSON                    = `{"id":1,"name":"Jon Snow"}`
	usersJSON                   = `[{"id":1,"name":"Jon Snow"}]`
	userXML                     = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm                    = `id=1&name=Jon Snow`
	invalidContent              = "invalid content"
	userJSONInvalidType         = `{"id":"1","name":"Jon Snow"}`
	userXMLConvertNumberError   = `<user><id>Number one</id><name>Jon Snow</name></user>`
	userXMLUnsupportedTypeError = `<user><>Number one</><name>Jon Snow</name></user>`
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

const userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`

var dummyQuery = url.Values{"dummy": []string{"useless"}}

func TestEcho(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Router
	assert.NotNil(t, e.Router())

	e.HTTPErrorHandler(c, errors.New("error"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestNewWithConfig(t *testing.T) {
	e := NewWithConfig(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.GET("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "Hello, World!")
	})
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, `Hello, World!`, rec.Body.String())
}

func TestEcho_StaticFS(t *testing.T) {
	var testCases = []struct {
		givenFs              fs.FS
		name                 string
		givenPrefix          string
		givenFsRoot          string
		whenURL              string
		expectHeaderLocation string
		expectBodyStartsWith string
		expectStatus         int
	}{
		{
			name:                 "ok",
			givenPrefix:          "/images",
			givenFs:              os.DirFS("./_fixture/images"),
			whenURL:              "/images/walle.png",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: string([]byte{0x89, 0x50, 0x4e, 0x47}),
		},
		{
			name:                 "ok, from sub fs",
			givenPrefix:          "/images",
			givenFs:              MustSubFS(os.DirFS("./_fixture/"), "images"),
			whenURL:              "/images/walle.png",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: string([]byte{0x89, 0x50, 0x4e, 0x47}),
		},
		{
			name:                 "No file",
			givenPrefix:          "/images",
			givenFs:              os.DirFS("_fixture/scripts"),
			whenURL:              "/images/bolt.png",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory",
			givenPrefix:          "/images",
			givenFs:              os.DirFS("_fixture/images"),
			whenURL:              "/images/",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory Redirect",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture/"),
			whenURL:              "/folder",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory Redirect with non-root path",
			givenPrefix:          "/static",
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/static",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/static/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Prefixed directory 404 (request URL without slash)",
			givenPrefix:          "/folder/", // trailing slash will intentionally not match "/folder"
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/folder", // no trailing slash
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Prefixed directory redirect (without slash redirect to slash)",
			givenPrefix:          "/folder", // no trailing slash shall match /folder and /folder/*
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/folder", // no trailing slash
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory with index.html",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending with slash)",
			givenPrefix:          "/assets/",
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending without slash)",
			givenPrefix:          "/assets",
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Sub-directory with index.html",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture"),
			whenURL:              "/folder/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "do not allow directory traversal (backslash - windows separator)",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture/"),
			whenURL:              `/..\\middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "do not allow directory traversal (slash - unix separator)",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture/"),
			whenURL:              `/../middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "open redirect vulnerability",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture/"),
			whenURL:              "/open.redirect.hackercom%2f..",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/open.redirect.hackercom/../", // location starting with `//open` would be very bad
			expectBodyStartsWith: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			tmpFs := tc.givenFs
			if tc.givenFsRoot != "" {
				tmpFs = MustSubFS(tmpFs, tc.givenFsRoot)
			}
			e.StaticFS(tc.givenPrefix, tmpFs)

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			body := rec.Body.String()
			if tc.expectBodyStartsWith != "" {
				assert.True(t, strings.HasPrefix(body, tc.expectBodyStartsWith))
			} else {
				assert.Equal(t, "", body)
			}

			if tc.expectHeaderLocation != "" {
				assert.Equal(t, tc.expectHeaderLocation, rec.Result().Header["Location"][0])
			} else {
				_, ok := rec.Result().Header["Location"]
				assert.False(t, ok)
			}
		})
	}
}

func TestEcho_FileFS(t *testing.T) {
	var testCases = []struct {
		whenFS           fs.FS
		name             string
		whenPath         string
		whenFile         string
		givenURL         string
		expectStartsWith []byte
		expectCode       int
	}{
		{
			name:             "ok",
			whenPath:         "/walle",
			whenFS:           os.DirFS("_fixture/images"),
			whenFile:         "walle.png",
			givenURL:         "/walle",
			expectCode:       http.StatusOK,
			expectStartsWith: []byte{0x89, 0x50, 0x4e},
		},
		{
			name:             "nok, requesting invalid path",
			whenPath:         "/walle",
			whenFS:           os.DirFS("_fixture/images"),
			whenFile:         "walle.png",
			givenURL:         "/walle.png",
			expectCode:       http.StatusNotFound,
			expectStartsWith: []byte(`{"message":"Not Found"}`),
		},
		{
			name:             "nok, serving not existent file from filesystem",
			whenPath:         "/walle",
			whenFS:           os.DirFS("_fixture/images"),
			whenFile:         "not-existent.png",
			givenURL:         "/walle",
			expectCode:       http.StatusNotFound,
			expectStartsWith: []byte(`{"message":"Not Found"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.FileFS(tc.whenPath, tc.whenFile, tc.whenFS)

			req := httptest.NewRequest(http.MethodGet, tc.givenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)

			body := rec.Body.Bytes()
			if len(body) > len(tc.expectStartsWith) {
				body = body[:len(tc.expectStartsWith)]
			}
			assert.Equal(t, tc.expectStartsWith, body)
		})
	}
}

func TestEcho_StaticPanic(t *testing.T) {
	var testCases = []struct {
		name      string
		givenRoot string
	}{
		{
			name:      "panics for ../",
			givenRoot: "../assets",
		},
		{
			name:      "panics for /",
			givenRoot: "/assets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Filesystem = os.DirFS("./")

			assert.Panics(t, func() {
				e.Static("../assets", tc.givenRoot)
			})
		})
	}
}

func TestEchoStaticRedirectIndex(t *testing.T) {
	e := New()

	// HandlerFunc
	ri := e.Static("/static", "_fixture")
	assert.Equal(t, http.MethodGet, ri.Method)
	assert.Equal(t, "/static*", ri.Path)
	assert.Equal(t, "GET:/static*", ri.Name)
	assert.Equal(t, []string{"*"}, ri.Parameters)

	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
	defer cancel()
	addr, err := startOnRandomPort(ctx, e)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	code, body, err := doGet(fmt.Sprintf("http://%v/static", addr))
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(body, "<!doctype html>"))
	assert.Equal(t, http.StatusOK, code)
}

func TestEchoFile(t *testing.T) {
	var testCases = []struct {
		name             string
		givenPath        string
		givenFile        string
		whenPath         string
		expectStartsWith string
		expectCode       int
	}{
		{
			name:             "ok",
			givenPath:        "/walle",
			givenFile:        "_fixture/images/walle.png",
			whenPath:         "/walle",
			expectCode:       http.StatusOK,
			expectStartsWith: string([]byte{0x89, 0x50, 0x4e}),
		},
		{
			name:             "ok with relative path",
			givenPath:        "/",
			givenFile:        "./go.mod",
			whenPath:         "/",
			expectCode:       http.StatusOK,
			expectStartsWith: "module github.com/labstack/echo/v",
		},
		{
			name:             "nok file does not exist",
			givenPath:        "/",
			givenFile:        "./this-file-does-not-exist",
			whenPath:         "/",
			expectCode:       http.StatusNotFound,
			expectStartsWith: "{\"message\":\"Not Found\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New() // we are using echo.defaultFS instance
			e.File(tc.givenPath, tc.givenFile)

			c, b := request(http.MethodGet, tc.whenPath, e)
			assert.Equal(t, tc.expectCode, c)

			if len(b) > len(tc.expectStartsWith) {
				b = b[:len(tc.expectStartsWith)]
			}
			assert.Equal(t, tc.expectStartsWith, b)
		})
	}
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			// before route match is found RouteInfo does not exist
			assert.Equal(t, RouteInfo{}, c.RouteInfo())
			buf.WriteString("-1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	// Route
	e.GET("/", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(http.MethodGet, "/", e)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return errors.New("error")
		}
	})
	e.GET("/", notFoundHandler)
	c, _ := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.GET("/ok", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(http.MethodGet, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoWrapHandler(t *testing.T) {
	e := New()

	var actualID string
	var actualPattern string
	e.GET("/:id", WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
		actualID = r.PathValue("id")
		actualPattern = r.Pattern
	})))

	req := httptest.NewRequest(http.MethodGet, "/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
	assert.Equal(t, "123", actualID)
	assert.Equal(t, "/:id", actualPattern)
}

func TestEchoWrapMiddleware(t *testing.T) {
	e := New()

	var actualID string
	var actualPattern string
	e.Use(WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualID = r.PathValue("id")
			actualPattern = r.Pattern
			h.ServeHTTP(w, r)
		})
	}))

	e.GET("/:id", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/123", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
	assert.Equal(t, "123", actualID)
	assert.Equal(t, "/:id", actualPattern)
}

func TestEchoConnect(t *testing.T) {
	e := New()

	ri := e.CONNECT("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodConnect, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodConnect+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodConnect, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoDelete(t *testing.T) {
	e := New()

	ri := e.DELETE("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodDelete, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodDelete+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodDelete, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoGet(t *testing.T) {
	e := New()

	ri := e.GET("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodGet, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodGet+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoHead(t *testing.T) {
	e := New()

	ri := e.HEAD("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodHead, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodHead+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodHead, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoOptions(t *testing.T) {
	e := New()

	ri := e.OPTIONS("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodOptions, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodOptions+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodOptions, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPatch(t *testing.T) {
	e := New()

	ri := e.PATCH("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPatch, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodPatch+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodPatch, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPost(t *testing.T) {
	e := New()

	ri := e.POST("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPost, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodPost+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodPost, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPut(t *testing.T) {
	e := New()

	ri := e.PUT("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPut, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodPut+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodPut, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoTrace(t *testing.T) {
	e := New()

	ri := e.TRACE("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodTrace, ri.Method)
	assert.Equal(t, "/", ri.Path)
	assert.Equal(t, http.MethodTrace+":/", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodTrace, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEcho_Any(t *testing.T) {
	e := New()

	ri := e.Any("/activate", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK from ANY")
	})

	assert.Equal(t, RouteAny, ri.Method)
	assert.Equal(t, "/activate", ri.Path)
	assert.Equal(t, RouteAny+":/activate", ri.Name)
	assert.Nil(t, ri.Parameters)

	status, body := request(http.MethodTrace, "/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK from ANY`, body)
}

func TestEcho_Any_hasLowerPriority(t *testing.T) {
	e := New()

	e.Any("/activate", func(c *Context) error {
		return c.String(http.StatusTeapot, "ANY")
	})
	e.GET("/activate", func(c *Context) error {
		return c.String(http.StatusLocked, "GET")
	})

	status, body := request(http.MethodTrace, "/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `ANY`, body)

	status, body = request(http.MethodGet, "/activate", e)
	assert.Equal(t, http.StatusLocked, status)
	assert.Equal(t, `GET`, body)
}

func TestEchoMatch(t *testing.T) { // JFC
	e := New()
	ris := e.Match([]string{http.MethodGet, http.MethodPost}, "/", func(c *Context) error {
		return c.String(http.StatusOK, "Match")
	})
	assert.Len(t, ris, 2)
}

func TestEchoServeHTTPPathEncoding(t *testing.T) {
	e := New()
	e.GET("/with/slash", func(c *Context) error {
		return c.String(http.StatusOK, "/with/slash")
	})
	e.GET("/:id", func(c *Context) error {
		return c.String(http.StatusOK, c.Param("id"))
	})

	var testCases = []struct {
		name         string
		whenURL      string
		expectURL    string
		expectStatus int
	}{
		{
			name:         "url with encoding is not decoded for routing",
			whenURL:      "/with%2Fslash",
			expectURL:    "with%2Fslash", // `%2F` is not decoded to `/` for routing
			expectStatus: http.StatusOK,
		},
		{
			name:         "url without encoding is used as is",
			whenURL:      "/with/slash",
			expectURL:    "/with/slash",
			expectStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.Equal(t, tc.expectURL, rec.Body.String())
		})
	}
}

func TestEchoGroup(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	e.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))
	h := func(c *Context) error {
		return c.NoContent(http.StatusOK)
	}

	//--------
	// Routes
	//--------

	e.GET("/users", h)

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.GET("", h)

	// Nested groups with middleware
	g2 := e.Group("/group2")
	g2.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})
	g3 := g2.Group("/group3")
	g3.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})
	g3.GET("", h)

	request(http.MethodGet, "/users", e)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request(http.MethodGet, "/group1", e)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request(http.MethodGet, "/group2/group3", e)
	assert.Equal(t, "023", buf.String())
}

func TestEcho_RouteNotFound(t *testing.T) {
	var testCases = []struct {
		expectRoute any
		name        string
		whenURL     string
		expectCode  int
	}{
		{
			name:        "404, route to static not found handler /a/c/xx",
			whenURL:     "/a/c/xx",
			expectRoute: "GET /a/c/xx",
			expectCode:  http.StatusNotFound,
		},
		{
			name:        "404, route to path param not found handler /a/:file",
			whenURL:     "/a/echo.exe",
			expectRoute: "GET /a/:file",
			expectCode:  http.StatusNotFound,
		},
		{
			name:        "404, route to any not found handler /*",
			whenURL:     "/b/echo.exe",
			expectRoute: "GET /*",
			expectCode:  http.StatusNotFound,
		},
		{
			name:        "200, route /a/c/df to /a/c/df",
			whenURL:     "/a/c/df",
			expectRoute: "GET /a/c/df",
			expectCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			okHandler := func(c *Context) error {
				return c.String(http.StatusOK, c.Request().Method+" "+c.Path())
			}
			notFoundHandler := func(c *Context) error {
				return c.String(http.StatusNotFound, c.Request().Method+" "+c.Path())
			}

			e.GET("/", okHandler)
			e.GET("/a/c/df", okHandler)
			e.GET("/a/b*", okHandler)
			e.PUT("/*", okHandler)

			e.RouteNotFound("/a/c/xx", notFoundHandler)  // static
			e.RouteNotFound("/a/:file", notFoundHandler) // param
			e.RouteNotFound("/*", notFoundHandler)       // any

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			assert.Equal(t, tc.expectRoute, rec.Body.String())
		})
	}
}

func TestEchoNotFound(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEchoMethodNotAllowed(t *testing.T) {
	e := New()

	e.GET("/", func(c *Context) error {
		return c.String(http.StatusOK, "Echo!")
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "OPTIONS, GET", rec.Header().Get(HeaderAllow))
}

func TestEcho_OnAddRoute(t *testing.T) {
	exampleRoute := Route{
		Method:      http.MethodGet,
		Path:        "/api/files/:id",
		Handler:     notFoundHandler,
		Middlewares: nil,
		Name:        "x",
	}

	var testCases = []struct {
		whenRoute   Route
		whenError   error
		name        string
		expectError string
		expectAdded []string
		expectLen   int
	}{
		{
			name:        "ok",
			whenRoute:   exampleRoute,
			whenError:   nil,
			expectAdded: []string{"/static", "/api/files/:id"},
			expectError: "",
			expectLen:   2,
		},
		{
			name:        "nok, error is returned",
			whenRoute:   exampleRoute,
			whenError:   errors.New("nope"),
			expectAdded: []string{"/static"},
			expectError: "nope",
			expectLen:   1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e := New()

			added := make([]string, 0)
			cnt := 0
			e.OnAddRoute = func(route Route) error {
				if cnt > 0 && tc.whenError != nil { // we want to GET /static to succeed for nok tests
					return tc.whenError
				}
				cnt++
				added = append(added, route.Path)
				return nil
			}

			e.GET("/static", notFoundHandler)

			var err error
			_, err = e.AddRoute(tc.whenRoute)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, e.Router().Routes(), tc.expectLen)
			assert.Equal(t, tc.expectAdded, added)
		})
	}
}

func TestEchoContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(Context), c)
	e.ReleaseContext(c)
}

func TestPreMiddlewares(t *testing.T) {
	e := New()
	assert.Equal(t, 0, len(e.PreMiddlewares()))

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return next(c)
		}
	})

	assert.Equal(t, 1, len(e.PreMiddlewares()))
}

func TestMiddlewares(t *testing.T) {
	e := New()
	assert.Equal(t, 0, len(e.Middlewares()))

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			return next(c)
		}
	})

	assert.Equal(t, 1, len(e.Middlewares()))
}

func TestEcho_Start(t *testing.T) {
	e := New()
	e.GET("/", func(c *Context) error {
		return c.String(http.StatusTeapot, "OK")
	})
	rndPort, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer rndPort.Close()
	errChan := make(chan error, 1)
	go func() {
		errChan <- e.Start(rndPort.Addr().String())
	}()

	select {
	case <-time.After(250 * time.Millisecond):
		t.Fatal("start did not error out")
	case err := <-errChan:
		expectContains := "bind: address already in use"
		if runtime.GOOS == "windows" {
			expectContains = "bind: Only one usage of each socket address"
		}
		assert.Contains(t, err.Error(), expectContains)
	}
}

func request(method, path string, e *Echo) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

type customError struct {
	Code    int
	Message string
}

func (ce *customError) StatusCode() int {
	return ce.Code
}

func (ce *customError) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"x":"%v"}`, ce.Message)), nil
}

func (ce *customError) Error() string {
	return ce.Message
}

func TestDefaultHTTPErrorHandler(t *testing.T) {
	var testCases = []struct {
		whenError        error
		name             string
		whenMethod       string
		expectBody       string
		expectLogged     string
		expectStatus     int
		givenExposeError bool
		givenLoggerFunc  bool
	}{
		{
			name:             "ok, expose error = true, HTTPError, no wrapped err",
			givenExposeError: true,
			whenError:        &HTTPError{Code: http.StatusTeapot, Message: "my_error"},
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"message":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + wrapped error",
			givenExposeError: true,
			whenError:        HTTPError{Code: http.StatusTeapot, Message: "my_error"}.Wrap(errors.New("internal_error")),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"error":"internal_error","message":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + wrapped HTTPError",
			givenExposeError: true,
			whenError:        HTTPError{Code: http.StatusTeapot, Message: "my_error"}.Wrap(&HTTPError{Code: http.StatusTeapot, Message: "early_error"}),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"error":"code=418, message=early_error","message":"my_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError",
			whenError:    &HTTPError{Code: http.StatusTeapot, Message: "my_error"},
			expectStatus: http.StatusTeapot,
			expectBody:   `{"message":"my_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError, no message",
			whenError:    &HTTPError{Code: http.StatusTeapot, Message: ""},
			expectStatus: http.StatusTeapot,
			expectBody:   `{"message":"I'm a teapot"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError + internal HTTPError",
			whenError:    HTTPError{Code: http.StatusTooEarly, Message: "my_error"}.Wrap(&HTTPError{Code: http.StatusTeapot, Message: "early_error"}),
			expectStatus: http.StatusTooEarly,
			expectBody:   `{"message":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, Error",
			givenExposeError: true,
			whenError:        fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus:     http.StatusInternalServerError,
			expectBody:       `{"error":"my errors wraps: internal_error","message":"Internal Server Error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, Error",
			whenError:    fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus: http.StatusInternalServerError,
			expectBody:   `{"message":"Internal Server Error"}` + "\n",
		},
		{
			name:             "ok, http.HEAD, expose error = true, Error",
			givenExposeError: true,
			whenMethod:       http.MethodHead,
			whenError:        fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus:     http.StatusInternalServerError,
			expectBody:       ``,
		},
		{
			name:         "ok, custom error implement MarshalJSON + HTTPStatusCoder",
			whenMethod:   http.MethodGet,
			whenError:    &customError{Code: http.StatusTeapot, Message: "custom error msg"},
			expectStatus: http.StatusTeapot,
			expectBody:   `{"x":"custom error msg"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			e := New()
			e.Logger = slog.New(slog.DiscardHandler)
			e.Any("/path", func(c *Context) error {
				return tc.whenError
			})

			e.HTTPErrorHandler = DefaultHTTPErrorHandler(tc.givenExposeError)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			c, b := request(method, "/path", e)

			assert.Equal(t, tc.expectStatus, c)
			assert.Equal(t, tc.expectBody, b)
			assert.Equal(t, tc.expectLogged, buf.String())
		})
	}
}

func TestDefaultHTTPErrorHandler_CommitedResponse(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)

	c.orgResponse.Committed = true
	errHandler := DefaultHTTPErrorHandler(false)

	errHandler(c, errors.New("my_error"))
	assert.Equal(t, http.StatusOK, resp.Code)
}

func benchmarkEchoRoutes(b *testing.B, routes []testRoute) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	u := req.URL
	w := httptest.NewRecorder()

	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		e.Add(route.Method, route.Path, func(c *Context) error {
			return nil
		})
	}

	// Find routes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			req.Method = route.Method
			u.Path = route.Path
			e.ServeHTTP(w, req)
		}
	}
}

func BenchmarkEchoStaticRoutes(b *testing.B) {
	benchmarkEchoRoutes(b, staticRoutes)
}

func BenchmarkEchoStaticRoutesMisses(b *testing.B) {
	benchmarkEchoRoutes(b, staticRoutes)
}

func BenchmarkEchoGitHubAPI(b *testing.B) {
	benchmarkEchoRoutes(b, gitHubAPI)
}

func BenchmarkEchoGitHubAPIMisses(b *testing.B) {
	benchmarkEchoRoutes(b, gitHubAPI)
}

func BenchmarkEchoParseAPI(b *testing.B) {
	benchmarkEchoRoutes(b, parseAPI)
}

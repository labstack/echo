package echo

import (
	"bytes"
	stdContext "context"
	"errors"
	"fmt"
	"io/fs"
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

func TestEcho_StaticFS(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenPrefix          string
		givenFs              fs.FS
		givenFsRoot          string
		whenURL              string
		expectStatus         int
		expectHeaderLocation string
		expectBodyStartsWith string
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
		name             string
		whenPath         string
		whenFile         string
		whenFS           fs.FS
		givenURL         string
		expectCode       int
		expectStartsWith []byte
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
		name        string
		givenRoot   string
		expectError string
	}{
		{
			name:        "panics for ../",
			givenRoot:   "../assets",
			expectError: "can not create sub FS, invalid root given, err: sub ../assets: invalid name",
		},
		{
			name:        "panics for /",
			givenRoot:   "/assets",
			expectError: "can not create sub FS, invalid root given, err: sub /assets: invalid name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Filesystem = os.DirFS("./")

			assert.PanicsWithError(t, tc.expectError, func() {
				e.Static("../assets", tc.givenRoot)
			})
		})
	}
}

func TestEchoStaticRedirectIndex(t *testing.T) {
	e := New()

	// HandlerFunc
	ri := e.Static("/static", "_fixture")
	assert.Equal(t, http.MethodGet, ri.Method())
	assert.Equal(t, "/static*", ri.Path())
	assert.Equal(t, "GET:/static*", ri.Name())
	assert.Equal(t, []string{"*"}, ri.Params())

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
		expectCode       int
		expectStartsWith string
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
		return func(c Context) error {
			// before route match is found RouteInfo does not exist
			assert.Equal(t, nil, c.RouteInfo())
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

	c, b := request(http.MethodGet, "/", e)
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
	e.GET("/", notFoundHandler)
	c, _ := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(http.MethodGet, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoWrapHandler(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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

func TestEchoGet_routeInfoIsImmutable(t *testing.T) {
	e := New()
	ri := e.GET("/test", handlerFunc)
	assert.Equal(t, "GET:/test", ri.Name())

	riFromRouter, err := e.Router().Routes().FindByMethodPath(http.MethodGet, "/test")
	assert.NoError(t, err)
	assert.Equal(t, "GET:/test", riFromRouter.Name())

	rInfo := ri.(routeInfo)
	rInfo.name = "changed" // this change should not change other returned values

	assert.Equal(t, "GET:/test", ri.Name())

	riFromRouter, err = e.Router().Routes().FindByMethodPath(http.MethodGet, "/test")
	assert.NoError(t, err)
	assert.Equal(t, "GET:/test", riFromRouter.Name())
}

func TestEchoConnect(t *testing.T) {
	e := New()

	ri := e.CONNECT("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodConnect, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodConnect+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodConnect, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoDelete(t *testing.T) {
	e := New()

	ri := e.DELETE("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodDelete, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodDelete+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodDelete, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoGet(t *testing.T) {
	e := New()

	ri := e.GET("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodGet, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodGet+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoHead(t *testing.T) {
	e := New()

	ri := e.HEAD("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodHead, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodHead+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodHead, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoOptions(t *testing.T) {
	e := New()

	ri := e.OPTIONS("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodOptions, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodOptions+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodOptions, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPatch(t *testing.T) {
	e := New()

	ri := e.PATCH("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPatch, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodPatch+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPatch, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPost(t *testing.T) {
	e := New()

	ri := e.POST("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPost, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodPost+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPost, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoPut(t *testing.T) {
	e := New()

	ri := e.PUT("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPut, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodPut+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPut, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoTrace(t *testing.T) {
	e := New()

	ri := e.TRACE("/", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodTrace, ri.Method())
	assert.Equal(t, "/", ri.Path())
	assert.Equal(t, http.MethodTrace+":/", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodTrace, "/", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OK", body)
}

func TestEchoAny(t *testing.T) { // JFC
	e := New()
	ris := e.Any("/", func(c Context) error {
		return c.String(http.StatusOK, "Any")
	})
	assert.Len(t, ris, 11)
}

func TestEchoMatch(t *testing.T) { // JFC
	e := New()
	ris := e.Match([]string{http.MethodGet, http.MethodPost}, "/", func(c Context) error {
		return c.String(http.StatusOK, "Match")
	})
	assert.Len(t, ris, 2)
}

func TestEcho_Routers_HandleHostsProperly(t *testing.T) {
	e := New()
	h := e.Host("route.com")
	routes := []*Route{
		{Method: http.MethodGet, Path: "/users/:user/events"},
		{Method: http.MethodGet, Path: "/users/:user/events/public"},
		{Method: http.MethodPost, Path: "/repos/:owner/:repo/git/refs"},
		{Method: http.MethodPost, Path: "/repos/:owner/:repo/git/tags"},
	}
	for _, r := range routes {
		h.Add(r.Method, r.Path, func(c Context) error {
			return c.String(http.StatusOK, "OK")
		})
	}

	routers := e.Routers()

	routeCom, ok := routers["route.com"]
	assert.True(t, ok)

	if assert.Equal(t, len(routes), len(routeCom.Routes())) {
		for _, r := range routeCom.Routes() {
			found := false
			for _, rr := range routes {
				if r.Method() == rr.Method && r.Path() == rr.Path {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Route %s %s not found", r.Method(), r.Path())
			}
		}
	}
}

func TestEchoServeHTTPPathEncoding(t *testing.T) {
	e := New()
	e.GET("/with/slash", func(c Context) error {
		return c.String(http.StatusOK, "/with/slash")
	})
	e.GET("/:id", func(c Context) error {
		return c.String(http.StatusOK, c.PathParam("id"))
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

func TestEchoHost(t *testing.T) {
	okHandler := func(c Context) error { return c.String(http.StatusOK, http.StatusText(http.StatusOK)) }
	teapotHandler := func(c Context) error { return c.String(http.StatusTeapot, http.StatusText(http.StatusTeapot)) }
	acceptHandler := func(c Context) error { return c.String(http.StatusAccepted, http.StatusText(http.StatusAccepted)) }
	teapotMiddleware := MiddlewareFunc(func(next HandlerFunc) HandlerFunc { return teapotHandler })

	e := New()
	e.GET("/", acceptHandler)
	e.GET("/foo", acceptHandler)

	ok := e.Host("ok.com")
	ok.GET("/", okHandler)
	ok.GET("/foo", okHandler)

	teapot := e.Host("teapot.com")
	teapot.GET("/", teapotHandler)
	teapot.GET("/foo", teapotHandler)

	middle := e.Host("middleware.com", teapotMiddleware)
	middle.GET("/", okHandler)
	middle.GET("/foo", okHandler)

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

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.Equal(t, tc.expectBody, rec.Body.String())
		})
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
		name        string
		whenURL     string
		expectRoute interface{}
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

			okHandler := func(c Context) error {
				return c.String(http.StatusOK, c.Request().Method+" "+c.Path())
			}
			notFoundHandler := func(c Context) error {
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

	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "Echo!")
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "OPTIONS, GET", rec.Header().Get(HeaderAllow))
}

func TestEcho_OnAddRoute(t *testing.T) {
	type rr struct {
		host string
		path string
	}
	exampleRoute := Route{
		Method:      http.MethodGet,
		Path:        "/api/files/:id",
		Handler:     notFoundHandler,
		Middlewares: nil,
		Name:        "x",
	}

	var testCases = []struct {
		name        string
		whenHost    string
		whenRoute   Routable
		whenError   error
		expectLen   int
		expectAdded []rr
		expectError string
	}{
		{
			name:      "ok, for default host",
			whenHost:  "",
			whenRoute: exampleRoute,
			whenError: nil,
			expectAdded: []rr{
				{host: "", path: "/static"},
				{host: "", path: "/api/files/:id"},
			},
			expectError: "",
			expectLen:   2,
		},
		{
			name:      "ok, for specific host",
			whenHost:  "test.com",
			whenRoute: exampleRoute,
			whenError: nil,
			expectAdded: []rr{
				{host: "", path: "/static"},
				{host: "test.com", path: "/api/files/:id"},
			},
			expectError: "",
			expectLen:   1,
		},
		{
			name:      "nok, error is returned",
			whenHost:  "test.com",
			whenRoute: exampleRoute,
			whenError: errors.New("nope"),
			expectAdded: []rr{
				{host: "", path: "/static"},
			},
			expectError: "nope",
			expectLen:   0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e := New()

			added := make([]rr, 0)
			cnt := 0
			e.OnAddRoute = func(host string, route Routable) error {
				if cnt > 0 && tc.whenError != nil { // we want to GET /static to succeed for nok tests
					return tc.whenError
				}
				cnt++
				added = append(added, rr{
					host: host,
					path: route.ToRoute().Path,
				})
				return nil
			}

			e.GET("/static", notFoundHandler)

			var err error
			if tc.whenHost != "" {
				_, err = e.Host(tc.whenHost).AddRoute(tc.whenRoute)
			} else {
				_, err = e.AddRoute(tc.whenRoute)
			}

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			r, _ := e.RouterFor(tc.whenHost)
			assert.Len(t, r.Routes(), tc.expectLen)
			assert.Equal(t, tc.expectAdded, added)
		})
	}
}

func TestEcho_RouterFor(t *testing.T) {
	var testCases = []struct {
		name      string
		whenHost  string
		expectLen int
		expectOk  bool
	}{
		{
			name:      "ok, default host",
			whenHost:  "",
			expectLen: 2,
			expectOk:  true,
		},
		{
			name:      "ok, specific host with routes",
			whenHost:  "test.com",
			expectLen: 1,
			expectOk:  true,
		},
		{
			name:      "ok, non-existent host",
			whenHost:  "oups.com",
			expectLen: 0,
			expectOk:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/1", notFoundHandler)
			e.GET("/2", notFoundHandler)
			e.Host("test.com").GET("/3", notFoundHandler)

			r, ok := e.RouterFor(tc.whenHost)
			assert.Equal(t, tc.expectOk, ok)
			if tc.expectLen > 0 {
				assert.Len(t, r.Routes(), tc.expectLen)
			} else {
				assert.Nil(t, r)
			}
		})
	}
}

func TestEchoContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(DefaultContext), c)
	e.ReleaseContext(c)
}

func TestEcho_Start(t *testing.T) {
	e := New()
	e.GET("/", func(c Context) error {
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

func TestDefaultHTTPErrorHandler(t *testing.T) {
	var testCases = []struct {
		name             string
		givenExposeError bool
		givenLoggerFunc  bool
		whenMethod       string
		whenError        error
		expectBody       string
		expectStatus     int
		expectLogged     string
	}{
		{
			name:             "ok, expose error = true, HTTPError",
			givenExposeError: true,
			whenError:        NewHTTPError(http.StatusTeapot, "my_error"),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"error":"code=418, message=my_error","message":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + internal error",
			givenExposeError: true,
			whenError:        NewHTTPError(http.StatusTeapot, "my_error").WithInternal(errors.New("internal_error")),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"error":"code=418, message=my_error, internal=internal_error","message":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + internal HTTPError",
			givenExposeError: true,
			whenError:        NewHTTPError(http.StatusTeapot, "my_error").WithInternal(NewHTTPError(http.StatusTooEarly, "early_error")),
			expectStatus:     http.StatusTooEarly,
			expectBody:       `{"error":"code=418, message=my_error, internal=code=425, message=early_error","message":"early_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError",
			whenError:    NewHTTPError(http.StatusTeapot, "my_error"),
			expectStatus: http.StatusTeapot,
			expectBody:   `{"message":"my_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError + internal HTTPError",
			whenError:    NewHTTPError(http.StatusTeapot, "my_error").WithInternal(NewHTTPError(http.StatusTooEarly, "early_error")),
			expectStatus: http.StatusTooEarly,
			expectBody:   `{"message":"early_error"}` + "\n",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			e := New()
			e.Logger = &jsonLogger{writer: buf}
			e.Any("/path", func(c Context) error {
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

type myCustomContext struct {
	DefaultContext
}

func (c *myCustomContext) QueryParam(name string) string {
	return "prefix_" + c.DefaultContext.QueryParam(name)
}

func TestEcho_customContext(t *testing.T) {
	e := New()
	e.NewContextFunc = func(ec *Echo, pathParamAllocSize int) ServableContext {
		return &myCustomContext{
			DefaultContext: *NewDefaultContext(ec, pathParamAllocSize),
		}
	}

	e.GET("/info/:id/:file", func(c Context) error {
		return c.String(http.StatusTeapot, c.QueryParam("param"))
	})

	status, body := request(http.MethodGet, "/info/1/a.csv?param=123", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "prefix_123", body)
}

func benchmarkEchoRoutes(b *testing.B, routes []testRoute) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	u := req.URL
	w := httptest.NewRecorder()

	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		e.Add(route.Method, route.Path, func(c Context) error {
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

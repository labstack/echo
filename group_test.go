package echo

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroup_withoutRouteWillNotExecuteMiddleware(t *testing.T) {
	e := New()

	called := false
	mw := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			called = true
			return c.NoContent(http.StatusTeapot)
		}
	}
	// even though group has middleware it will not be executed when there are no routes under that group
	_ = e.Group("/group", mw)

	status, body := request(http.MethodGet, "/group/nope", e)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, `{"message":"Not Found"}`+"\n", body)

	assert.False(t, called)
}

func TestGroup_withRoutesWillNotExecuteMiddlewareFor404(t *testing.T) {
	e := New()

	called := false
	mw := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			called = true
			return c.NoContent(http.StatusTeapot)
		}
	}
	// even though group has middleware and routes when we have no match on some route the middlewares for that
	// group will not be executed
	g := e.Group("/group", mw)
	g.GET("/yes", handlerFunc)

	status, body := request(http.MethodGet, "/group/nope", e)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, `{"message":"Not Found"}`+"\n", body)

	assert.False(t, called)
}

func TestGroup_multiLevelGroup(t *testing.T) {
	e := New()

	api := e.Group("/api")
	users := api.Group("/users")
	users.GET("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	status, body := request(http.MethodGet, "/api/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroupFile(t *testing.T) {
	e := New()
	g := e.Group("/group")
	g.File("/walle", "_fixture/images/walle.png")
	expectedData, err := ioutil.ReadFile("_fixture/images/walle.png")
	assert.Nil(t, err)
	req := httptest.NewRequest(http.MethodGet, "/group/walle", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expectedData, rec.Body.Bytes())
}

func TestGroupRouteMiddleware(t *testing.T) {
	// Ensure middleware slices are not re-used
	e := New()
	g := e.Group("/group")
	h := func(Context) error { return nil }
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m4 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(404)
		}
	}
	m5 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(405)
		}
	}
	g.Use(m1, m2, m3)
	g.GET("/404", h, m4)
	g.GET("/405", h, m5)

	c, _ := request(http.MethodGet, "/group/404", e)
	assert.Equal(t, 404, c)
	c, _ = request(http.MethodGet, "/group/405", e)
	assert.Equal(t, 405, c)
}

func TestGroupRouteMiddlewareWithMatchAny(t *testing.T) {
	// Ensure middleware and match any routes do not conflict
	e := New()
	g := e.Group("/group")
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, c.RouteInfo().Path())
		}
	}
	h := func(c Context) error {
		return c.String(http.StatusOK, c.RouteInfo().Path())
	}
	g.Use(m1)
	g.GET("/help", h, m2)
	g.GET("/*", h, m2)
	g.GET("", h, m2)
	e.GET("unrelated", h, m2)
	e.GET("*", h, m2)

	_, m := request(http.MethodGet, "/group/help", e)
	assert.Equal(t, "/group/help", m)
	_, m = request(http.MethodGet, "/group/help/other", e)
	assert.Equal(t, "/group/*", m)
	_, m = request(http.MethodGet, "/group/404", e)
	assert.Equal(t, "/group/*", m)
	_, m = request(http.MethodGet, "/group", e)
	assert.Equal(t, "/group", m)
	_, m = request(http.MethodGet, "/other", e)
	assert.Equal(t, "/*", m)
	_, m = request(http.MethodGet, "/", e)
	assert.Equal(t, "/*", m)

}

func TestGroup_CONNECT(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.CONNECT("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodConnect, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodConnect+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodConnect, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_DELETE(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.DELETE("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodDelete, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodDelete+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodDelete, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_HEAD(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.HEAD("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodHead, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodHead+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodHead, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_OPTIONS(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.OPTIONS("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodOptions, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodOptions+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodOptions, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_PATCH(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.PATCH("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPatch, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodPatch+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPatch, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_POST(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.POST("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPost, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodPost+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPost, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_PUT(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.PUT("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodPut, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodPut+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodPut, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_TRACE(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ri := users.TRACE("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})

	assert.Equal(t, http.MethodTrace, ri.Method())
	assert.Equal(t, "/users/activate", ri.Path())
	assert.Equal(t, http.MethodTrace+":/users/activate", ri.Name())
	assert.Nil(t, ri.Params())

	status, body := request(http.MethodTrace, "/users/activate", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, `OK`, body)
}

func TestGroup_Any(t *testing.T) {
	e := New()

	users := e.Group("/users")
	ris := users.Any("/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})
	assert.Len(t, ris, 11)

	for _, m := range methods {
		status, body := request(m, "/users/activate", e)
		assert.Equal(t, http.StatusTeapot, status)
		assert.Equal(t, `OK`, body)
	}
}

func TestGroup_AnyWithErrors(t *testing.T) {
	e := New()

	users := e.Group("/users")
	users.GET("/activate", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	errs := func() (errs []error) {
		defer func() {
			if r := recover(); r != nil {
				if tmpErr, ok := r.([]error); ok {
					errs = tmpErr
					return
				}
				panic(r)
			}
		}()

		users.Any("/activate", func(c Context) error {
			return c.String(http.StatusTeapot, "OK")
		})
		return nil
	}()
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "GET /users/activate: adding duplicate route (same method+path) is not allowed")

	for _, m := range methods {
		status, body := request(m, "/users/activate", e)

		expect := http.StatusTeapot
		if m == http.MethodGet {
			expect = http.StatusOK
		}
		assert.Equal(t, expect, status)
		assert.Equal(t, `OK`, body)
	}
}

func TestGroup_Match(t *testing.T) {
	e := New()

	myMethods := []string{http.MethodGet, http.MethodPost}
	users := e.Group("/users")
	ris := users.Match(myMethods, "/activate", func(c Context) error {
		return c.String(http.StatusTeapot, "OK")
	})
	assert.Len(t, ris, 2)

	for _, m := range myMethods {
		status, body := request(m, "/users/activate", e)
		assert.Equal(t, http.StatusTeapot, status)
		assert.Equal(t, `OK`, body)
	}
}

func TestGroup_MatchWithErrors(t *testing.T) {
	e := New()

	users := e.Group("/users")
	users.GET("/activate", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})
	myMethods := []string{http.MethodGet, http.MethodPost}

	errs := func() (errs []error) {
		defer func() {
			if r := recover(); r != nil {
				if tmpErr, ok := r.([]error); ok {
					errs = tmpErr
					return
				}
				panic(r)
			}
		}()

		users.Match(myMethods, "/activate", func(c Context) error {
			return c.String(http.StatusTeapot, "OK")
		})
		return nil
	}()
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "GET /users/activate: adding duplicate route (same method+path) is not allowed")

	for _, m := range myMethods {
		status, body := request(m, "/users/activate", e)

		expect := http.StatusTeapot
		if m == http.MethodGet {
			expect = http.StatusOK
		}
		assert.Equal(t, expect, status)
		assert.Equal(t, `OK`, body)
	}
}

func TestGroup_Static(t *testing.T) {
	e := New()

	g := e.Group("/books")
	ri := g.Static("/download", "_fixture")
	assert.Equal(t, http.MethodGet, ri.Method())
	assert.Equal(t, "/books/download*", ri.Path())
	assert.Equal(t, "GET:/books/download*", ri.Name())
	assert.Equal(t, []string{"*"}, ri.Params())

	req := httptest.NewRequest(http.MethodGet, "/books/download/index.html", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.True(t, strings.HasPrefix(body, "<!doctype html>"))
}

func TestGroup_StaticMultiTest(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenPrefix          string
		givenRoot            string
		whenURL              string
		expectStatus         int
		expectHeaderLocation string
		expectBodyStartsWith string
	}{
		{
			name:                 "ok",
			givenPrefix:          "/images",
			givenRoot:            "_fixture/images",
			whenURL:              "/test/images/walle.png",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: string([]byte{0x89, 0x50, 0x4e, 0x47}),
		},
		{
			name:                 "ok, without prefix",
			givenPrefix:          "",
			givenRoot:            "_fixture/images",
			whenURL:              "/testwalle.png", // `/test` + `*` creates route `/test*` witch matches `/testwalle.png`
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: string([]byte{0x89, 0x50, 0x4e, 0x47}),
		},
		{
			name:                 "nok, without prefix does not serve dir index",
			givenPrefix:          "",
			givenRoot:            "_fixture/images",
			whenURL:              "/test/", // `/test` + `*` creates route `/test*`
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "No file",
			givenPrefix:          "/images",
			givenRoot:            "_fixture/scripts",
			whenURL:              "/test/images/bolt.png",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory",
			givenPrefix:          "/images",
			givenRoot:            "_fixture/images",
			whenURL:              "/test/images/",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory Redirect",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/test/folder",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/test/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory Redirect with non-root path",
			givenPrefix:          "/static",
			givenRoot:            "_fixture",
			whenURL:              "/test/static",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/test/static/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Prefixed directory 404 (request URL without slash)",
			givenPrefix:          "/folder/", // trailing slash will intentionally not match "/folder"
			givenRoot:            "_fixture",
			whenURL:              "/test/folder", // no trailing slash
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Prefixed directory redirect (without slash redirect to slash)",
			givenPrefix:          "/folder", // no trailing slash shall match /folder and /folder/*
			givenRoot:            "_fixture",
			whenURL:              "/test/folder", // no trailing slash
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/test/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory with index.html",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/test/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending with slash)",
			givenPrefix:          "/assets/",
			givenRoot:            "_fixture",
			whenURL:              "/test/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending without slash)",
			givenPrefix:          "/assets",
			givenRoot:            "_fixture",
			whenURL:              "/test/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Sub-directory with index.html",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/test/folder/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "do not allow directory traversal (backslash - windows separator)",
			givenPrefix:          "/",
			givenRoot:            "_fixture/",
			whenURL:              `/test/..\\middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "do not allow directory traversal (slash - unix separator)",
			givenPrefix:          "/",
			givenRoot:            "_fixture/",
			whenURL:              `/test/../middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			g := e.Group("/test")
			g.Static(tc.givenPrefix, tc.givenRoot)

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

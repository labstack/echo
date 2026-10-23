// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatic_useCaseForApiAndSPAs(t *testing.T) {
	e := echo.New()

	// serve single page application (SPA) files from server root
	e.Use(StaticWithConfig(StaticConfig{
		Root: "testdata/dist/public",
	}))

	// all requests to `/api/*` will end up in echo handlers (assuming there is not `api` folder and files)
	api := e.Group("/api")
	users := api.Group("/users")
	users.GET("/info", func(c *echo.Context) error {
		return c.String(http.StatusOK, "users info")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/info", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "users info", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<h1>Hello from index</h1>\n")

}

func TestStaticHTML5PreservesMatchedHandlerNotFound(t *testing.T) {
	var testCases = []struct {
		name           string
		buildEcho      func() *echo.Echo
		whenURL        string
		expectCode     int
		expectJSONEq   string
		expectContains string
	}{
		{
			name: "ok, router-matched handler 404 is preserved instead of serving index",
			buildEcho: func() *echo.Echo {
				e := echo.New()
				e.Use(StaticWithConfig(StaticConfig{
					Root:  "testdata/dist/public",
					HTML5: true,
				}))
				e.GET("/api/users/:id", func(c *echo.Context) error {
					return echo.NewHTTPError(http.StatusNotFound, "user not found")
				})
				return e
			},
			whenURL:      "/api/users/42",
			expectCode:   http.StatusNotFound,
			expectJSONEq: `{"message":"user not found"}`,
		},
		{
			name: "ok, router-level 404 for group without matched route still serves index",
			buildEcho: func() *echo.Echo {
				e := echo.New()
				e.Group("/app", StaticWithConfig(StaticConfig{
					Root:  "testdata/dist/public",
					HTML5: true,
				}))
				return e
			},
			whenURL:        "/app/dashboard",
			expectCode:     http.StatusOK,
			expectContains: "<h1>Hello from index</h1>\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.buildEcho()

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			if tc.expectJSONEq != "" {
				assert.JSONEq(t, tc.expectJSONEq, rec.Body.String())
			}
			if tc.expectContains != "" {
				assert.Contains(t, rec.Body.String(), tc.expectContains)
			}
		})
	}
}

func TestStatic(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenConfig          *StaticConfig
		givenAttachedToGroup string
		whenURL              string
		expectContains       string
		expectNotContains    string
		expectLength         string
		expectCode           int
	}{
		{
			name:           "ok, serve index with Echo message",
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "<h1>Hello from index</h1>",
		},
		{
			name:           "ok, serve file from subdirectory",
			whenURL:        "/assets/readme.md",
			expectCode:     http.StatusOK,
			expectContains: "This directory is used for the static middleware test",
		},
		{
			name: "ok, when html5 mode serve index for any static file that does not exist",
			givenConfig: &StaticConfig{
				Root:  "testdata/dist/public",
				HTML5: true,
			},
			whenURL:        "/random",
			expectCode:     http.StatusOK,
			expectContains: "<h1>Hello from index</h1>",
		},
		{
			name: "ok, when html5 mode serve route html next to a directory",
			givenConfig: &StaticConfig{
				Root:  "testdata/dist/public",
				HTML5: true,
			},
			whenURL:        "/camera",
			expectCode:     http.StatusOK,
			expectContains: "<h1>Camera page</h1>",
		},
		{
			name: "ok, serve index as directory index listing files directory",
			givenConfig: &StaticConfig{
				Root:   "testdata/dist/public/assets",
				Browse: true,
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: `<a class="file" href="readme.md">readme.md</a>`,
		},
		{
			name: "ok, serve directory index with IgnoreBase and browse",
			givenConfig: &StaticConfig{
				Root:       "testdata/dist/public/assets/", // <-- last `assets/` is overlapping with group path and needs to be ignored
				IgnoreBase: true,
				Browse:     true,
			},
			givenAttachedToGroup: "/assets",
			whenURL:              "/assets/",
			expectCode:           http.StatusOK,
			expectContains:       `<a class="file" href="readme.md">readme.md</a>`,
		},
		{
			name: "ok, serve file with IgnoreBase",
			givenConfig: &StaticConfig{
				Root:       "testdata/dist/public/assets", // <-- last `assets/` is overlapping with group path and needs to be ignored
				IgnoreBase: true,
				Browse:     true,
			},
			givenAttachedToGroup: "/assets",
			whenURL:              "/assets/readme.md",
			expectCode:           http.StatusOK,
			expectContains:       "This directory is used for the static middleware test",
		},
		{
			name:           "nok, file not found",
			whenURL:        "/none",
			expectCode:     http.StatusNotFound,
			expectContains: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:           "ok, when no file then a handler will care of the request",
			whenURL:        "/regular-handler",
			expectCode:     http.StatusOK,
			expectContains: "ok",
		},
		{
			name: "ok, skip middleware and serve handler",
			givenConfig: &StaticConfig{
				Root: "testdata/dist/public",
				Skipper: func(c *echo.Context) bool {
					return true
				},
			},
			whenURL:        "/walle.png",
			expectCode:     http.StatusTeapot,
			expectContains: "walle",
		},
		{
			name: "nok, when html5 fail if the index file does not exist",
			givenConfig: &StaticConfig{
				Root:  "testdata/dist/public",
				HTML5: true,
				Index: "missing.html", // that folder contains `index.html`
			},
			whenURL:    "/random",
			expectCode: http.StatusInternalServerError,
		},
		{
			name: "ok, serve from http.FileSystem",
			givenConfig: &StaticConfig{
				Root:       "public",
				Filesystem: os.DirFS("testdata/dist"),
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "<h1>Hello from index</h1>",
		},
		{
			name:              "nok, do not allow directory traversal (backslash - windows separator)",
			whenURL:           `/..\\private.txt`,
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok,do not allow directory traversal (slash - unix separator)",
			whenURL:           `/../private.txt`,
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, URL encoded path traversal (single encoding, slash - unix separator)",
			whenURL:           "/%2e%2e%2fprivate.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, URL encoded path traversal (single encoding, backslash - windows separator)",
			whenURL:           "/%2e%2e%5cprivate.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, URL encoded path traversal (double encoding, slash - unix separator)",
			whenURL:           "/%252e%252e%252fprivate.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, URL encoded path traversal (double encoding, backslash - windows separator)",
			whenURL:           "/%252e%252e%255cprivate.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, URL encoded path traversal (mixed encoding)",
			whenURL:           "/%2e%2e/private.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, backslash URL encoded",
			whenURL:           "/..%5c..%5cprivate.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		//{ // Under windows, %00 gets cleaned out by `http.ReadRequest` making this test to fail with different code
		//	name:           "nok, null byte injection",
		//	whenURL:        "/index.html%00.jpg",
		//	expectCode:     http.StatusInternalServerError,
		//	expectContains: "{\"message\":\"Internal Server Error\"}\n",
		//},
		{
			name:              "nok, mixed backslash and forward slash traversal",
			whenURL:           "/..\\../private.txt",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
		{
			name:              "nok, trailing dots (Windows edge case)",
			whenURL:           "/../private.txt...",
			expectCode:        http.StatusNotFound,
			expectContains:    "{\"message\":\"Not Found\"}\n",
			expectNotContains: `private file`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			config := StaticConfig{Root: "testdata/dist/public"}
			if tc.givenConfig != nil {
				config = *tc.givenConfig
			}
			middlewareFunc := StaticWithConfig(config)
			if tc.givenAttachedToGroup != "" {
				// middleware is attached to group
				subGroup := e.Group(tc.givenAttachedToGroup, middlewareFunc)
				// group without http handlers (routes) does not do anything.
				// Request is matched against http handlers (routes) that have group middleware attached to them
				subGroup.GET("", func(c *echo.Context) error { return echo.ErrNotFound })
				subGroup.GET("/*", func(c *echo.Context) error { return echo.ErrNotFound })
			} else {
				// middleware is on root level
				e.Use(middlewareFunc)
				e.GET("/regular-handler", func(c *echo.Context) error {
					return c.String(http.StatusOK, "ok")
				})
				e.GET("/walle.png", func(c *echo.Context) error {
					return c.String(http.StatusTeapot, "walle")
				})
			}

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			responseBody := rec.Body.String()
			if tc.expectContains != "" {
				assert.Contains(t, responseBody, tc.expectContains)
			}
			if tc.expectNotContains != "" {
				assert.NotContains(t, responseBody, tc.expectNotContains)
			}
			if tc.expectLength != "" {
				assert.Equal(t, tc.expectLength, rec.Header().Get(echo.HeaderContentLength))
			}
		})
	}
}

func TestStaticMiddlewareAndRouterInconsistentEscaping(t *testing.T) {
	var testCases = []struct {
		name                  string
		givenConfig           StaticConfig
		whenURL               string
		expectCode            int
		expectBodyContains    string
		expectBodyNotContains string
	}{
		{
			name:               "ok, normal file is served",
			givenConfig:        StaticConfig{Root: "testdata/dist/public"},
			whenURL:            "/test.txt",
			expectCode:         http.StatusOK,
			expectBodyContains: "test",
		},
		{
			name:        "ok, direct request to restricted path is blocked by ACL route",
			givenConfig: StaticConfig{Root: "testdata/dist/public"},
			whenURL:     "/admin/private.txt",
			expectCode:  http.StatusForbidden,
		},
		{
			// With EnablePathUnescaping=false (default/safe), the wildcard param "admin%2fprivate.txt"
			// is NOT decoded, so the FS lookup is for literal "admin%2fprivate.txt" which does
			// not exist → falls through to the /* handler → 404. ACL is not bypassed.
			name:                  "ok, encoded slash returns 404 with default safe config (EnablePathUnescaping=false)",
			givenConfig:           StaticConfig{Root: "testdata/dist/public"},
			whenURL:               "/admin%2fprivate.txt",
			expectCode:            http.StatusNotFound,
			expectBodyNotContains: "private file",
		},
		{
			// With EnablePathUnescaping=true, the wildcard param "admin%2fprivate.txt" IS decoded
			// to "admin/private.txt". The router already routed to /* (encoded %2f prevented matching
			// /admin/*), so the ACL guard never ran. The file is served — ACL bypass.
			// Only use EnablePathUnescaping: true when not relying on route-based ACL guards.
			name:               "nok, encoded slash bypasses ACL when EnablePathUnescaping=true",
			givenConfig:        StaticConfig{Root: "testdata/dist/public", EnablePathUnescaping: true},
			whenURL:            "/admin%2fprivate.txt",
			expectCode:         http.StatusOK,
			expectBodyContains: "dist/public/admin/private.txt - private file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			// Global middleware runs for all matched routes. The /* wildcard route ensures
			// the middleware takes the c.Param("*") branch (c.Path() ends with "*").
			e.Use(StaticWithConfig(tc.givenConfig))
			e.GET("/*", func(c *echo.Context) error { return echo.ErrNotFound })
			// ACL guard: requests with a literal /admin/ prefix are forbidden.
			e.GET("/admin/*", func(c *echo.Context) error { return echo.ErrForbidden })

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			body := rec.Body.String()
			if tc.expectBodyContains != "" {
				assert.Contains(t, body, tc.expectBodyContains)
			}
			if tc.expectBodyNotContains != "" {
				assert.NotContains(t, body, tc.expectBodyNotContains)
			}
		})
	}
}

// Regression for GHSA-vfp3-v2gw-7wfq: the static middleware mounted on a group
// must not let an encoded separator in the wildcard bypass route-level middleware
// and disclose a file the matched route never authorized.
func TestStatic_EncodedSeparatorDoesNotBypassRoute(t *testing.T) {
	fsys := fstest.MapFS{
		"admin/secret.txt": {Data: []byte("TOP-SECRET")},
		"index.html":       {Data: []byte("public")},
	}
	e := echo.New()
	g := e.Group("/files", StaticWithConfig(StaticConfig{Filesystem: fsys}))
	g.GET("/*", func(c *echo.Context) error { return echo.ErrNotFound })

	cases := []struct {
		target   string
		wantCode int
	}{
		{"/files/index.html", http.StatusOK},
		{"/files/admin%2Fsecret.txt", http.StatusNotFound},
		{"/files/admin%5Csecret.txt", http.StatusNotFound},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.target, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, tc.wantCode, rec.Code, "GET %s", tc.target)
		assert.NotContains(t, rec.Body.String(), "TOP-SECRET", "GET %s leaked protected file", tc.target)
	}
}

func TestMustStaticWithConfig_panicsInvalidDirListTemplate(t *testing.T) {
	assert.Panics(t, func() {
		StaticWithConfig(StaticConfig{DirectoryListTemplate: `{{}`})
	})
}

func TestFormat(t *testing.T) {
	var testCases = []struct {
		name   string
		when   int64
		expect string
	}{
		{
			name:   "byte",
			when:   0,
			expect: "0",
		},
		{
			name:   "bytes",
			when:   515,
			expect: "515B",
		},
		{
			name:   "KB",
			when:   31323,
			expect: "30.59KB",
		},
		{
			name:   "MB",
			when:   13231323,
			expect: "12.62MB",
		},
		{
			name:   "GB",
			when:   7323232398,
			expect: "6.82GB",
		},
		{
			name:   "TB",
			when:   1_099_511_627_776,
			expect: "1.00TB",
		},
		{
			name:   "PB",
			when:   9923232398434432,
			expect: "8.81PB",
		},
		{
			// test with 7EB because of https://github.com/labstack/gommon/pull/38 and https://github.com/labstack/gommon/pull/43
			//
			// 8 exbi equals 2^64, therefore it cannot be stored in int64. The tests use
			// the fact that on x86_64 the following expressions holds true:
			// int64(0) - 1 == math.MaxInt64.
			//
			// However, this is not true for other platforms, specifically aarch64, s390x
			// and ppc64le.
			name:   "EB",
			when:   8070450532247929000,
			expect: "7.00EB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := format(tc.when)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestStatic_CustomFS(t *testing.T) {
	var testCases = []struct {
		name           string
		filesystem     fs.FS
		root           string
		whenURL        string
		expectContains string
		expectCode     int
	}{
		{
			name:           "ok, serve index with Echo message",
			whenURL:        "/",
			filesystem:     os.DirFS("../_fixture"),
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},

		{
			name:           "ok, serve index with Echo message",
			whenURL:        "/_fixture/",
			filesystem:     os.DirFS(".."),
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
		{
			name:    "ok, serve file from map fs",
			whenURL: "/file.txt",
			filesystem: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("file.txt is ok")},
			},
			expectCode:     http.StatusOK,
			expectContains: "file.txt is ok",
		},
		{
			name:       "nok, missing file in map fs",
			whenURL:    "/file.txt",
			expectCode: http.StatusNotFound,
			filesystem: fstest.MapFS{
				"file2.txt": &fstest.MapFile{Data: []byte("file2.txt is ok")},
			},
		},
		{
			name:    "nok, file is not a subpath of root",
			whenURL: `/../../secret.txt`,
			root:    "/nested/folder",
			filesystem: fstest.MapFS{
				"secret.txt": &fstest.MapFile{Data: []byte("this is a secret")},
			},
			expectCode: http.StatusNotFound,
		},
		{
			name:       "nok, backslash is forbidden",
			whenURL:    `/..\..\secret.txt`,
			expectCode: http.StatusNotFound,
			root:       "/nested/folder",
			filesystem: fstest.MapFS{
				"secret.txt": &fstest.MapFile{Data: []byte("this is a secret")},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			config := StaticConfig{
				Root:       ".",
				Filesystem: tc.filesystem,
			}

			if tc.root != "" {
				config.Root = tc.root
			}

			middlewareFunc, err := config.ToMiddleware()
			assert.NoError(t, err)

			e.Use(middlewareFunc)

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			if tc.expectContains != "" {
				responseBody := rec.Body.String()
				assert.Contains(t, responseBody, tc.expectContains)
			}
		})
	}
}

func TestStatic_DirectoryBrowsing(t *testing.T) {
	var testCases = []struct {
		name              string
		givenConfig       StaticConfig
		whenURL           string
		expectContains    string
		expectNotContains []string
		expectCode        int
	}{
		{
			name: "ok, should return index.html contents from Root=public folder",
			givenConfig: StaticConfig{
				Root:       "public",
				Filesystem: os.DirFS("../_fixture/dist"),
				Browse:     true,
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: `<h1>Hello from index</h1>`,
		},
		{
			name: "ok, should return only subfolder folder listing from Root=public/assets",
			givenConfig: StaticConfig{
				Root:       "public",
				Filesystem: os.DirFS("../_fixture/dist"),
				Browse:     true,
			},
			whenURL:        "/assets",
			expectCode:     http.StatusOK,
			expectContains: `<a class="file" href="readme.md">readme.md</a>`,
			expectNotContains: []string{
				`<h1>Hello from index</h1>`, // should see the listing, not index.html contents
				`private.txt`,               // file from the parent folder
				`subfolder.md`,              // file from subfolder
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			middlewareFunc, err := tc.givenConfig.ToMiddleware()
			assert.NoError(t, err)

			e.Use(middlewareFunc)

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)

			responseBody := rec.Body.String()
			if tc.expectContains != "" {
				assert.Contains(t, responseBody, tc.expectContains, "body should contain: "+tc.expectContains)
			}
			for _, notContains := range tc.expectNotContains {
				assert.NotContains(t, responseBody, notContains, "body should NOT contain: "+notContains)
			}
		})
	}
}

func TestStatic_resolvesRoutedPath(t *testing.T) {
	// GHSA-375p-5qhx-8wq4 (encoded paths) and GHSA-3pmx-cf9f-34xr (dot segments): the static middleware must resolve
	// files from the same path that the router matched, so encoding or dot segments do not reach a different file.
	// Note: middleware registered with e.Use runs before group middleware, so the `/admin` group guard below does not
	// protect files served by it (see Static docs); TestStatic_routeLevelGuard covers a guardable setup.
	fsys := fstest.MapFS{
		"admin/secret.txt":       {Data: []byte("SECRET")},
		"public/index.html":      {Data: []byte("public")},
		"hello world.txt":        {Data: []byte("hello")},
		"public/admin/other.txt": {Data: []byte("other")},
	}
	var testCases = []struct {
		name         string
		whenPath     string // set as req.URL.Path
		whenRawPath  string // set as req.URL.RawPath
		expectStatus int
		expectBody   string
	}{
		{name: "encoded slash", whenPath: "/admin/secret.txt", whenRawPath: "/admin%2Fsecret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded slash lowercase", whenPath: "/admin/secret.txt", whenRawPath: "/admin%2fsecret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded letter", whenPath: "/admin/secret.txt", whenRawPath: "/%61dmin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded dot segment", whenPath: "/public/../admin/secret.txt", whenRawPath: "/public/%2E%2E/admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded slashes and dots", whenPath: "/public/../admin/secret.txt", whenRawPath: "/public%2F..%2Fadmin%2Fsecret.txt", expectStatus: http.StatusNotFound},
		{name: "parent dot segment", whenPath: "/public/../admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "current dot segment", whenPath: "/./admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "empty segment", whenPath: "//admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "canonical escape is still served", whenPath: "/hello world.txt", expectStatus: http.StatusOK, expectBody: "hello"},
		{name: "plain public file is still served", whenPath: "/public/index.html", expectStatus: http.StatusOK, expectBody: "public"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(StaticWithConfig(StaticConfig{Filesystem: fsys}))
			admin := e.Group("/admin", func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c *echo.Context) error {
					return echo.ErrForbidden
				}
			})
			admin.GET("/*", func(c *echo.Context) error {
				return c.String(http.StatusOK, "admin")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL.Path = tc.whenPath
			req.URL.RawPath = tc.whenRawPath
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.NotContains(t, rec.Body.String(), "SECRET")
			if tc.expectBody != "" {
				assert.Equal(t, tc.expectBody, rec.Body.String())
			}
		})
	}
}

func TestStatic_routeLevelGuard(t *testing.T) {
	fsys := fstest.MapFS{
		"admin/secret.txt": {Data: []byte("SECRET")},
		"hello world.txt":  {Data: []byte("hello")},
		"x/file.txt":       {Data: []byte("x")},
	}
	var testCases = []struct {
		name                      string
		givenEnablePathUnescaping bool
		whenPath                  string // set as req.URL.Path
		whenRawPath               string // set as req.URL.RawPath
		expectStatus              int
	}{
		{name: "guarded path", whenPath: "/admin/secret.txt", expectStatus: http.StatusForbidden},
		{name: "guarded path repeated prefix", whenPath: "/admin/admin/secret.txt", expectStatus: http.StatusForbidden},
		{name: "encoded slash", whenPath: "/admin/secret.txt", whenRawPath: "/admin%2Fsecret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded letter", whenPath: "/admin/secret.txt", whenRawPath: "/%61dmin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "parent dot segment", whenPath: "/x/../admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "empty segment", whenPath: "//admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded dots with unescaping", givenEnablePathUnescaping: true, whenPath: "/x/../admin/secret.txt", whenRawPath: "/x/%2e%2e/admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "encoded dot slash with unescaping", givenEnablePathUnescaping: true, whenPath: "/./admin/secret.txt", whenRawPath: "/.%2Fadmin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "public file", whenPath: "/hello world.txt", expectStatus: http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			static := StaticWithConfig(StaticConfig{Filesystem: fsys, EnablePathUnescaping: tc.givenEnablePathUnescaping})
			e.GET("/*", func(c *echo.Context) error {
				return echo.ErrNotFound
			}, static)
			admin := e.Group("/admin", func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c *echo.Context) error {
					return echo.ErrForbidden
				}
			})
			admin.GET("/*", func(c *echo.Context) error {
				return c.String(http.StatusOK, "admin")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL.Path = tc.whenPath
			req.URL.RawPath = tc.whenRawPath
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.NotContains(t, rec.Body.String(), "SECRET")
		})
	}
}

func TestStatic_HTML5WithUncleanPath(t *testing.T) {
	// A path with an empty or dot segment is not served as a file, but HTML5 mode still serves the index for it.
	e := echo.New()
	e.Use(StaticWithConfig(StaticConfig{
		Filesystem: fstest.MapFS{"index.html": {Data: []byte("spa")}},
		HTML5:      true,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Path = "/app//route"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "spa", rec.Body.String())
}

// nonValidatingFS is a custom fs.FS that does not enforce fs.ValidPath, so a name with ".." escapes its root. Echo must
// never pass such a name to a user supplied filesystem.
type nonValidatingFS struct{ root string }

func (f nonValidatingFS) Open(name string) (fs.File, error) {
	// treat a backslash as a separator on every OS, like filepath.Join does on Windows
	return os.Open(filepath.Join(f.root, filepath.FromSlash(strings.ReplaceAll(name, `\`, "/"))))
}

func TestStatic_nonValidatingCustomFSCannotEscapeRoot(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "public"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "public", "index.txt"), []byte("public"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o644))

	targets := []string{
		"/../secret.txt",
		"/%2e%2e/secret.txt",
		"/..%2fsecret.txt",
		"/sub/../../secret.txt",
		"/..%5csecret.txt",
		"/..%5Csecret.txt",
		"/a%5C..%5C..%5Csecret.txt",
		`/..\secret.txt`,
	}
	for _, group := range []string{"", "/static"} {
		for _, unescape := range []bool{false, true} {
			e := echo.New()
			mw := StaticWithConfig(StaticConfig{
				Filesystem:           nonValidatingFS{root: filepath.Join(dir, "public")},
				EnablePathUnescaping: unescape,
			})
			if group == "" {
				e.Use(mw)
			} else {
				e.Group(group, mw)
			}

			for _, target := range targets {
				req := httptest.NewRequest(http.MethodGet, group+target, nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusNotFound, rec.Code, "%s unescape=%v", group+target, unescape)
				assert.NotContains(t, rec.Body.String(), "secret", "%s unescape=%v", group+target, unescape)
			}

			req := httptest.NewRequest(http.MethodGet, group+"/index.txt", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code, "group=%q unescape=%v", group, unescape)
			assert.Equal(t, "public", rec.Body.String())
		}
	}
}

func TestHasDotOrEmptySegment(t *testing.T) {
	var testCases = []struct {
		path   string
		expect bool
	}{
		{path: "", expect: false},
		{path: "/", expect: false},
		{path: "/index.html", expect: false},
		{path: "/css/app.css", expect: false},
		{path: "/..", expect: true},
		{path: "/a/../b", expect: true},
		{path: "/a/./b", expect: true},
		{path: "/a//b", expect: true},
		{path: `/..\secret.txt`, expect: true},
		{path: `/a\..\b`, expect: true},
		{path: `/.\secret.txt`, expect: true},
		{path: `/\..`, expect: true},
		{path: `/..\`, expect: true},
		{path: `/dir\file.txt`, expect: false},
		{path: "/...", expect: false},
		{path: "/..foo", expect: false},
		{path: `/a\\b`, expect: false},
		{path: "/..%2fsecret.txt", expect: false}, // still encoded, only unsafe once unescaped
	}
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expect, hasDotOrEmptySegment(tc.path))
		})
	}
}

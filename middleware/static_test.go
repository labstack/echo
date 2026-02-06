// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
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

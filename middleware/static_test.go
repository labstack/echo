package middleware

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestStatic_useCaseForApiAndSPAs(t *testing.T) {
	e := echo.New()

	// serve single page application (SPA) files from server root
	e.Use(StaticWithConfig(StaticConfig{
		Root: ".",
		// by default Echo filesystem is fixed to `./` but this does not allow `../` (moving up in folder structure past filesystem root)
		Filesystem: os.DirFS("../_fixture"),
	}))

	// all requests to `/api/*` will end up in echo handlers (assuming there is not `api` folder and files)
	api := e.Group("/api")
	users := api.Group("/users")
	users.GET("/info", func(c echo.Context) error {
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
	assert.Contains(t, rec.Body.String(), "<title>Echo</title>")

}

func TestStatic(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenConfig          *StaticConfig
		givenAttachedToGroup string
		whenURL              string
		expectContains       string
		expectLength         string
		expectCode           int
	}{
		{
			name:           "ok, serve index with Echo message",
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
		{
			name:         "ok, serve file from subdirectory",
			whenURL:      "/images/walle.png",
			expectCode:   http.StatusOK,
			expectLength: "219885",
		},
		{
			name: "ok, when html5 mode serve index for any static file that does not exist",
			givenConfig: &StaticConfig{
				Root:  "_fixture",
				HTML5: true,
			},
			whenURL:        "/random",
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
		{
			name: "ok, serve index as directory index listing files directory",
			givenConfig: &StaticConfig{
				Root:   "_fixture/certs",
				Browse: true,
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "cert.pem",
		},
		{
			name: "ok, serve directory index with IgnoreBase and browse",
			givenConfig: &StaticConfig{
				Root:       "_fixture/_fixture/", // <-- last `_fixture/` is overlapping with group path and needs to be ignored
				IgnoreBase: true,
				Browse:     true,
			},
			givenAttachedToGroup: "/_fixture",
			whenURL:              "/_fixture/",
			expectCode:           http.StatusOK,
			expectContains:       `<a class="file" href="README.md">README.md</a>`,
		},
		{
			name: "ok, serve file with IgnoreBase",
			givenConfig: &StaticConfig{
				Root:       "_fixture/_fixture/", // <-- last `_fixture/` is overlapping with group path and needs to be ignored
				IgnoreBase: true,
				Browse:     true,
			},
			givenAttachedToGroup: "/_fixture",
			whenURL:              "/_fixture/README.md",
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
			name:           "nok, do not allow directory traversal (backslash - windows separator)",
			whenURL:        `/..\\middleware/basic_auth.go`,
			expectCode:     http.StatusNotFound,
			expectContains: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:           "nok,do not allow directory traversal (slash - unix separator)",
			whenURL:        `/../middleware/basic_auth.go`,
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
				Root: "_fixture/images/",
				Skipper: func(c echo.Context) bool {
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
				Root:  "_fixture",
				HTML5: true,
				Index: "missing.html",
			},
			whenURL:    "/random",
			expectCode: http.StatusInternalServerError,
		},
		{
			name: "ok, serve from http.FileSystem",
			givenConfig: &StaticConfig{
				Root:       "_fixture",
				Filesystem: os.DirFS(".."),
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Filesystem = os.DirFS("../")

			config := StaticConfig{Root: "_fixture"}
			if tc.givenConfig != nil {
				config = *tc.givenConfig
			}
			middlewareFunc := StaticWithConfig(config)
			if tc.givenAttachedToGroup != "" {
				// middleware is attached to group
				subGroup := e.Group(tc.givenAttachedToGroup, middlewareFunc)
				// group without http handlers (routes) does not do anything.
				// Request is matched against http handlers (routes) that have group middleware attached to them
				subGroup.GET("", func(c echo.Context) error { return echo.ErrNotFound })
				subGroup.GET("/*", func(c echo.Context) error { return echo.ErrNotFound })
			} else {
				// middleware is on root level
				e.Use(middlewareFunc)
				e.GET("/regular-handler", func(c echo.Context) error {
					return c.String(http.StatusOK, "ok")
				})
				e.GET("/walle.png", func(c echo.Context) error {
					return c.String(http.StatusTeapot, "walle")
				})
			}

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectCode, rec.Code)
			if tc.expectContains != "" {
				responseBody := rec.Body.String()
				assert.Contains(t, responseBody, tc.expectContains)
			}
			if tc.expectLength != "" {
				assert.Equal(t, rec.Header().Get(echo.HeaderContentLength), tc.expectLength)
			}
		})
	}
}

func TestStatic_GroupWithStatic(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenGroup           string
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
			whenURL:              "/group/images/walle.png",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: string([]byte{0x89, 0x50, 0x4e, 0x47}),
		},
		{
			name:                 "No file",
			givenPrefix:          "/images",
			givenRoot:            "_fixture/scripts",
			whenURL:              "/group/images/bolt.png",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory not found (no trailing slash)",
			givenPrefix:          "/images",
			givenRoot:            "_fixture/images",
			whenURL:              "/group/images/",
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Directory redirect",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/group/folder",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/group/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory redirect",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/group/folder%2f..",
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/group/folder/../",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Prefixed directory 404 (request URL without slash)",
			givenGroup:           "_fixture",
			givenPrefix:          "/folder/", // trailing slash will intentionally not match "/folder"
			givenRoot:            "_fixture",
			whenURL:              "/_fixture/folder", // no trailing slash
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "Prefixed directory redirect (without slash redirect to slash)",
			givenGroup:           "_fixture",
			givenPrefix:          "/folder", // no trailing slash shall match /folder and /folder/*
			givenRoot:            "_fixture",
			whenURL:              "/_fixture/folder", // no trailing slash
			expectStatus:         http.StatusMovedPermanently,
			expectHeaderLocation: "/_fixture/folder/",
			expectBodyStartsWith: "",
		},
		{
			name:                 "Directory with index.html",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/group/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending with slash)",
			givenPrefix:          "/assets/",
			givenRoot:            "_fixture",
			whenURL:              "/group/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Prefixed directory with index.html (prefix ending without slash)",
			givenPrefix:          "/assets",
			givenRoot:            "_fixture",
			whenURL:              "/group/assets/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "Sub-directory with index.html",
			givenPrefix:          "/",
			givenRoot:            "_fixture",
			whenURL:              "/group/folder/",
			expectStatus:         http.StatusOK,
			expectBodyStartsWith: "<!doctype html>",
		},
		{
			name:                 "do not allow directory traversal (backslash - windows separator)",
			givenPrefix:          "/",
			givenRoot:            "_fixture/",
			whenURL:              `/group/..\\middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                 "do not allow directory traversal (slash - unix separator)",
			givenPrefix:          "/",
			givenRoot:            "_fixture/",
			whenURL:              `/group/../middleware/basic_auth.go`,
			expectStatus:         http.StatusNotFound,
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Filesystem = os.DirFS("../") // so we can access test files

			group := "/group"
			if tc.givenGroup != "" {
				group = tc.givenGroup
			}
			g := e.Group(group)
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
				assert.Equal(t, tc.expectHeaderLocation, rec.Header().Get(echo.HeaderLocation))
			} else {
				_, ok := rec.Result().Header[echo.HeaderLocation]
				assert.False(t, ok)
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

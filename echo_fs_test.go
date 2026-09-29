// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEcho_StaticFS(t *testing.T) {
	var testCases = []struct {
		name                                 string
		givenPrefix                          string
		givenFs                              fs.FS
		givenFsRoot                          string
		givenEnablePathUnescapingStaticFiles bool
		whenURL                              string
		expectStatus                         int
		expectHeaderLocation                 string
		expectBodyStartsWith                 string
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
			name:                 "do not unescape path variables by default",
			givenPrefix:          "/",
			givenFs:              os.DirFS("_fixture/"),
			whenURL:              "/open.redirect.hackercom%2f..",
			expectStatus:         http.StatusNotFound,
			expectHeaderLocation: "",
			expectBodyStartsWith: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                                 "do not accept encoded dots in path (%2E%2E is `..`) to traverse within filesystem boundary",
			givenPrefix:                          "/",
			givenFs:                              os.DirFS("_fixture/"),
			givenEnablePathUnescapingStaticFiles: false,
			whenURL:                              `/folder/%2E%2E/index.html`, // `/folder/../index.html`
			expectStatus:                         http.StatusNotFound,
			expectBodyStartsWith:                 "{\"message\":\"Not Found\"}\n",
		},
		{
			name:                                 "nok, encoded dots (%2E%2E is `..`) are rejected after unescaping (GHSA-3pmx-cf9f-34xr)",
			givenPrefix:                          "/",
			givenFs:                              os.DirFS("_fixture/"),
			givenEnablePathUnescapingStaticFiles: true,
			whenURL:                              `/folder/%2E%2E/index.html`, // `/folder/../index.html`
			expectStatus:                         http.StatusNotFound,
			expectBodyStartsWith:                 "{\"message\":\"Not Found\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.EnablePathUnescapingStaticFiles = tc.givenEnablePathUnescapingStaticFiles

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

func TestSanitizeURI(t *testing.T) {
	var testCases = []struct {
		whenURI string
		expect  string
	}{
		{whenURI: "/path/", expect: "/path/"},
		{whenURI: "//example.com/", expect: "/example.com/"},
		{whenURI: "/\t/example.com/", expect: "/%09/example.com/"},
		{whenURI: "/\t\\example.com/", expect: "/%09\\example.com/"},
		{whenURI: "/\x7f/example.com/", expect: "/%7F/example.com/"},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURI, func(t *testing.T) {
			assert.Equal(t, tc.expect, sanitizeURI(tc.whenURI))
		})
	}
}

func TestEcho_StaticFS_dotAndEmptySegments(t *testing.T) {
	// GHSA-3pmx-cf9f-34xr: paths with ".", ".." or empty segments must not reach a file under a guarded route.
	fsys := fstest.MapFS{
		"admin/secret.txt":  {Data: []byte("SECRET")},
		"public/index.html": {Data: []byte("public")},
	}
	var testCases = []struct {
		name         string
		whenPath     string
		expectStatus int
	}{
		{name: "guarded route", whenPath: "/admin/secret.txt", expectStatus: http.StatusForbidden},
		{name: "parent dot segment", whenPath: "/assets/../admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "current dot segment", whenPath: "/./admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "nested parent dot segments", whenPath: "/a/b/../../admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "empty segment", whenPath: "//admin/secret.txt", expectStatus: http.StatusNotFound},
		{name: "public file is still served", whenPath: "/public/index.html", expectStatus: http.StatusOK},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.StaticFS("/", fsys)
			e.GET("/admin/*", func(c Context) error {
				return ErrForbidden
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL.Path = tc.whenPath
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.NotContains(t, rec.Body.String(), "SECRET")
		})
	}
}

func TestStaticDirectoryHandler_encodedDotsWithPathUnescaping(t *testing.T) {
	// GHSA-3pmx-cf9f-34xr: with EnablePathUnescapingStaticFiles, dot segments created by unescaping must be rejected too.
	fsys := fstest.MapFS{"admin/secret.txt": {Data: []byte("SECRET")}, "x/file.txt": {Data: []byte("x")}}
	for _, u := range []string{"/x/%2e%2e/admin/secret.txt", "/.%2Fadmin/secret.txt", "/x/%252e%252e/admin/secret.txt"} {
		t.Run(u, func(t *testing.T) {
			e := New()
			e.EnablePathUnescapingStaticFiles = true
			e.StaticFS("/", fsys)
			e.GET("/admin/*", func(c Context) error {
				return ErrForbidden
			})

			req := httptest.NewRequest(http.MethodGet, u, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.NotContains(t, rec.Body.String(), "SECRET")
		})
	}
}

// nonValidatingDirFS is a custom fs.FS that does not enforce fs.ValidPath, so a name with ".." escapes its root. Echo
// must never pass such a name to a user supplied filesystem.
type nonValidatingDirFS struct{ root string }

func (f nonValidatingDirFS) Open(name string) (fs.File, error) {
	// treat a backslash as a separator on every OS, like filepath.Join does on Windows
	return os.Open(filepath.Join(f.root, filepath.FromSlash(strings.ReplaceAll(name, `\`, "/"))))
}

func TestEcho_StaticFS_nonValidatingCustomFSCannotEscapeRoot(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "public"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "public", "index.txt"), []byte("public"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o644))

	for _, unescape := range []bool{false, true} {
		e := New()
		e.EnablePathUnescapingStaticFiles = unescape
		e.StaticFS("/", nonValidatingDirFS{root: filepath.Join(dir, "public")})

		for _, target := range []string{
			"/../secret.txt",
			"/%2e%2e/secret.txt",
			"/..%2fsecret.txt",
			"/sub/../../secret.txt",
			"/..%5csecret.txt",
			"/..%5Csecret.txt",
			"/a%5C..%5C..%5Csecret.txt",
			`/..\secret.txt`,
		} {
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code, "%s unescape=%v", target, unescape)
			assert.NotContains(t, rec.Body.String(), "secret", "%s unescape=%v", target, unescape)
		}

		req := httptest.NewRequest(http.MethodGet, "/index.txt", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public", rec.Body.String())
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

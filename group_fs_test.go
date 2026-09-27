// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
)

func TestGroup_FileFS(t *testing.T) {
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
			givenURL:         "/assets/walle",
			expectCode:       http.StatusOK,
			expectStartsWith: []byte{0x89, 0x50, 0x4e},
		},
		{
			name:             "nok, requesting invalid path",
			whenPath:         "/walle",
			whenFS:           os.DirFS("_fixture/images"),
			whenFile:         "walle.png",
			givenURL:         "/assets/walle.png",
			expectCode:       http.StatusNotFound,
			expectStartsWith: []byte(`{"message":"Not Found"}`),
		},
		{
			name:             "nok, serving not existent file from filesystem",
			whenPath:         "/walle",
			whenFS:           os.DirFS("_fixture/images"),
			whenFile:         "not-existent.png",
			givenURL:         "/assets/walle",
			expectCode:       http.StatusNotFound,
			expectStartsWith: []byte(`{"message":"Not Found"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			g := e.Group("/assets")
			g.FileFS(tc.whenPath, tc.whenFile, tc.whenFS)

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

func TestGroup_StaticPanic(t *testing.T) {
	var testCases = []struct {
		name      string
		givenRoot string
	}{
		{
			name:      "panics for ../",
			givenRoot: "../images",
		},
		{
			name:      "panics for /",
			givenRoot: "/images",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Filesystem = os.DirFS("./")

			g := e.Group("/assets")

			assert.Panics(t, func() {
				g.Static("/images", tc.givenRoot)
			})
		})
	}
}

func TestGroup_StaticFS_dotAndEmptySegments(t *testing.T) {
	// GHSA-3pmx-cf9f-34xr: Group.StaticFS uses StaticDirectoryHandler, which must not serve paths with dot or empty
	// segments that path.Clean() would resolve under a guarded route.
	e := New()
	g := e.Group("/g")
	g.StaticFS("/", fstest.MapFS{"admin/secret.txt": {Data: []byte("SECRET")}, "x/file.txt": {Data: []byte("x")}})
	g.GET("/admin/*", func(c Context) error {
		return ErrForbidden
	})

	for _, p := range []string{"/g/x/../admin/secret.txt", "/g/./admin/secret.txt", "/g//admin/secret.txt"} {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL.Path = p
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
			assert.NotContains(t, rec.Body.String(), "SECRET")
		})
	}
	req := httptest.NewRequest(http.MethodGet, "/g/x/file.txt", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "x", rec.Body.String())
}

// +build go1.16

package middleware

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

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
				Filesystem: http.FS(tc.filesystem),
			}

			if tc.root != "" {
				config.Root = tc.root
			}

			middlewareFunc := StaticWithConfig(config)
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

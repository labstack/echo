package middleware

import (
	"github.com/siyual-park/echo-slim/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
				Root:  "../_fixture",
				HTML5: true,
			},
			whenURL:        "/random",
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
		{
			name: "ok, serve index as directory index listing files directory",
			givenConfig: &StaticConfig{
				Root:   "../_fixture/certs",
				Browse: true,
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "cert.pem",
		},
		{
			name: "ok, serve directory index with IgnoreBase and browse",
			givenConfig: &StaticConfig{
				Root:       "../_fixture/_fixture/", // <-- last `_fixture/` is overlapping with group path and needs to be ignored
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
				Root:       "../_fixture/_fixture/", // <-- last `_fixture/` is overlapping with group path and needs to be ignored
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
			name:           "nok, do not allow directory traversal (slash - unix separator)",
			whenURL:        `/../middleware/basic_auth.go`,
			expectCode:     http.StatusNotFound,
			expectContains: "{\"message\":\"Not Found\"}\n",
		},
		{
			name:           "ok, do not serve file, when a handler took care of the request",
			whenURL:        "/regular-handler",
			expectCode:     http.StatusOK,
			expectContains: "ok",
		},
		{
			name: "nok, when html5 fail if the index file does not exist",
			givenConfig: &StaticConfig{
				Root:  "../_fixture",
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
				Filesystem: http.Dir(".."),
			},
			whenURL:        "/",
			expectCode:     http.StatusOK,
			expectContains: "<title>Echo</title>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			r := NewRouter()

			e.Use(r.Routes())

			config := StaticConfig{Root: "../_fixture"}
			if tc.givenConfig != nil {
				config = *tc.givenConfig
			}
			middlewareFunc := StaticWithConfig(config)
			if tc.givenAttachedToGroup != "" {
				r.Any(tc.givenAttachedToGroup+"/*", middlewareFunc, func(next echo.HandlerFunc) echo.HandlerFunc {
					return echo.NotFoundHandler
				})
			} else {
				// middleware is on root level
				e.Use(middlewareFunc, func(next echo.HandlerFunc) echo.HandlerFunc {
					return echo.NotFoundHandler
				})
				r.GET("/regular-handler", func(next echo.HandlerFunc) echo.HandlerFunc {
					return func(c echo.Context) error {
						return c.String(http.StatusOK, "ok")
					}
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

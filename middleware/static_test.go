package middleware

import (
	"github.com/labstack/echo/v4"
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
			name:           "nok,do not allow directory traversal (slash - unix separator)",
			whenURL:        `/../middleware/basic_auth.go`,
			expectCode:     http.StatusNotFound,
			expectContains: "{\"message\":\"Not Found\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			config := StaticConfig{Root: "../_fixture"}
			if tc.givenConfig != nil {
				config = *tc.givenConfig
			}
			middlewareFunc := StaticWithConfig(config)
			if tc.givenAttachedToGroup != "" {
				// middleware is attached to group
				subGroup := e.Group(tc.givenAttachedToGroup, middlewareFunc)
				// group without http handlers (routes) does not do anything.
				// Request is matched against http handlers (routes) that have group middleware attached to them
				subGroup.GET("", echo.NotFoundHandler)
				subGroup.GET("/*", echo.NotFoundHandler)
			} else {
				// middleware is on root level
				e.Use(middlewareFunc)
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

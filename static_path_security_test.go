// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func TestStaticRejectsEncodedEmptySegments(t *testing.T) {
	for _, mode := range []string{"Static", "StaticFS", "prefix", "Group.Static", "Group.StaticFS", "middleware", "group middleware"} {
		for _, decodedMatching := range []bool{false, true} {
			for _, unescapeParams := range []bool{false, true} {
				for _, unescapeFiles := range []bool{false, true} {
					name := fmt.Sprintf("%s/decoded=%t/params=%t/files=%t", mode, decodedMatching, unescapeParams, unescapeFiles)
					t.Run(name, func(t *testing.T) {
						filesystem := fstest.MapFS{
							"public/text.txt":           {Data: []byte("PUBLIC")},
							"public/admin/private.txt":  {Data: []byte("PRIVATE_MARKER")},
							"public/100%.txt":           {Data: []byte("PERCENT")},
							"public/literal%2e%2e.txt":  {Data: []byte("ENCODED_NAME")},
							"public/%2e%2e/literal.txt": {Data: []byte("ENCODED_DIRECTORY")},
							"outside.txt":               {Data: []byte("OUTSIDE_MARKER")},
						}
						e := echo.NewWithConfig(echo.Config{
							Router: echo.NewRouter(echo.RouterConfig{
								UseEscapedPathForMatching: decodedMatching,
								UnescapePathParamValues:   unescapeParams,
							}),
							Filesystem:                      filesystem,
							EnablePathUnescapingStaticFiles: unescapeFiles,
						})
						guard := func(c *echo.Context) error { return echo.ErrForbidden }
						prefix := ""
						switch mode {
						case "Static":
							e.Static("/", "public")
						case "StaticFS":
							e.StaticFS("/", echo.MustSubFS(filesystem, "public"))
						case "prefix":
							prefix = "/assets"
							e.Static(prefix, "public")
						case "Group.Static":
							prefix = "/assets"
							e.Group(prefix).Static("/", "public")
						case "Group.StaticFS":
							prefix = "/assets"
							e.Group(prefix).StaticFS("/", echo.MustSubFS(filesystem, "public"))
						case "middleware":
							e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
								Root: "public", Filesystem: filesystem, EnablePathUnescaping: unescapeFiles,
							}))
						case "group middleware":
							prefix = "/assets"
							group := e.Group(prefix)
							group.Use(middleware.StaticWithConfig(middleware.StaticConfig{
								Root: "public", Filesystem: filesystem, EnablePathUnescaping: unescapeFiles,
							}))
							group.GET("/*", func(c *echo.Context) error { return echo.ErrNotFound })
							group.GET("/admin/*", guard)
						}
						if mode != "group middleware" {
							e.GET(prefix+"/admin/*", guard)
						}
						for _, tc := range []struct {
							path   string
							status int
							body   string
						}{
							{path: "/text.txt", status: http.StatusOK, body: "PUBLIC"},
							{path: "/admin/private.txt", status: http.StatusForbidden},
							{path: "/%2fadmin/private.txt"},
							{path: "/%2Fadmin/private.txt"},
							{path: "/%2fadmin%2fprivate.txt"},
							{path: "/%252fadmin/private.txt"},
							{path: "/%252Fadmin%252fprivate.txt"},
							{path: "//admin/private.txt"},
							{path: "/../outside.txt"},
							{path: "/%2f..%2f..%2foutside.txt"},
						} {
							t.Run(tc.path, func(t *testing.T) {
								rec := httptest.NewRecorder()
								e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, prefix+tc.path, nil))
								body := rec.Body.String()
								if tc.status != 0 && rec.Code != tc.status {
									t.Errorf("status=%d, want %d; body=%q", rec.Code, tc.status, body)
								}
								if tc.body != "" && body != tc.body {
									t.Errorf("body=%q, want %q", body, tc.body)
								}
								if strings.Contains(body, "PRIVATE_MARKER") || strings.Contains(body, "OUTSIDE_MARKER") {
									t.Errorf("protected content leaked: status=%d body=%q", rec.Code, body)
								}
							})
						}
						// Literal percent sequences remain usable with the default static-file decoding policy.
						if !unescapeFiles {
							for path, want := range map[string]string{
								"/100%25.txt":             "PERCENT",
								"/literal%252e%252e.txt":  "ENCODED_NAME",
								"/%252e%252e/literal.txt": "ENCODED_DIRECTORY",
							} {
								rec := httptest.NewRecorder()
								e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, prefix+path, nil))
								if rec.Code != http.StatusOK || rec.Body.String() != want {
									t.Errorf("GET %s: status=%d body=%q, want 200 and %q", path, rec.Code, rec.Body.String(), want)
								}
							}
						}
					})
				}
			}
		}
	}
}

func TestStaticEncodedEmptySegmentOverHTTP(t *testing.T) {
	e := echo.NewWithConfig(echo.Config{
		Router:     echo.NewRouter(echo.RouterConfig{UseEscapedPathForMatching: true}),
		Filesystem: fstest.MapFS{"public/admin/private.txt": {Data: []byte("PRIVATE_MARKER")}},
	})
	e.Static("/", "public")
	e.GET("/admin/*", func(c *echo.Context) error { return echo.ErrForbidden })
	server := httptest.NewServer(e)
	defer server.Close()
	response, err := http.Get(server.URL + "/%2fadmin/private.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound || strings.Contains(string(body), "PRIVATE_MARKER") {
		t.Errorf("status=%d body=%q, want 404 without protected content", response.StatusCode, body)
	}
}

func TestStaticRootGuardRejectsEncodedEmptySegment(t *testing.T) {
	for _, groupedMiddleware := range []bool{false, true} {
		for _, decodedMatching := range []bool{false, true} {
			for _, unescapeFiles := range []bool{false, true} {
				name := fmt.Sprintf("middleware=%t/decoded=%t/files=%t", groupedMiddleware, decodedMatching, unescapeFiles)
				t.Run(name, func(t *testing.T) {
					filesystem := fstest.MapFS{"public/index.html": {Data: []byte("PRIVATE_INDEX")}}
					e := echo.NewWithConfig(echo.Config{
						Router:     echo.NewRouter(echo.RouterConfig{UseEscapedPathForMatching: decodedMatching}),
						Filesystem: filesystem, EnablePathUnescapingStaticFiles: unescapeFiles,
					})
					prefix := ""
					guard := func(c *echo.Context) error { return echo.ErrForbidden }
					if groupedMiddleware {
						prefix = "/assets"
						group := e.Group(prefix)
						group.Use(middleware.StaticWithConfig(middleware.StaticConfig{
							Root: "public", Filesystem: filesystem, EnablePathUnescaping: unescapeFiles,
						}))
						group.GET("/*", func(c *echo.Context) error { return echo.ErrNotFound })
						group.GET("/", guard)
					} else {
						e.Static("/", "public")
						e.GET("/", guard)
					}
					for _, path := range []string{"/", "//", "/%2f", "/%252f"} {
						rec := httptest.NewRecorder()
						e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, prefix+path, nil))
						if path == "/" && rec.Code != http.StatusForbidden {
							t.Fatalf("root guard returned %d, want 403", rec.Code)
						}
						if rec.Code == http.StatusOK || strings.Contains(rec.Body.String(), "PRIVATE_INDEX") {
							t.Errorf("GET %s bypassed root guard: status=%d body=%q", path, rec.Code, rec.Body.String())
						}
					}
				})
			}
		}
	}
}

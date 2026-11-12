// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestAddTrailingSlashWithConfig(t *testing.T) {
	var testCases = []struct {
		whenURL        string
		whenMethod     string
		expectPath     string
		expectLocation []string
		expectStatus   int
	}{
		{
			whenURL:        "/add-slash",
			whenMethod:     http.MethodGet,
			expectPath:     "/add-slash",
			expectLocation: []string{`/add-slash/`},
		},
		{
			whenURL:        "/add-slash?key=value",
			whenMethod:     http.MethodGet,
			expectPath:     "/add-slash",
			expectLocation: []string{`/add-slash/?key=value`},
		},
		{
			whenURL:        "/",
			whenMethod:     http.MethodConnect,
			expectPath:     "/",
			expectLocation: nil,
			expectStatus:   http.StatusOK,
		},
		// cases for open redirect vulnerability
		{
			whenURL:        "http://localhost:1323/%5Cexample.com",
			expectPath:     `/\example.com`,
			expectLocation: []string{`/example.com/`},
		},
		{
			whenURL:        `http://localhost:1323/\example.com`,
			expectPath:     `/\example.com`,
			expectLocation: []string{`/example.com/`},
		},
		{
			whenURL:        `http://localhost:1323/\\%5C////%5C\\\example.com`,
			expectPath:     `/\\\////\\\\example.com`,
			expectLocation: []string{`/example.com/`},
		},
		{
			whenURL:        "http://localhost:1323//example.com",
			expectPath:     `//example.com`,
			expectLocation: []string{`/example.com/`},
		},
		{
			whenURL:        "http://localhost:1323/%5C%5C",
			expectPath:     `/\\`,
			expectLocation: []string{`/`},
		},
		// control characters that browsers strip from URLs must not hide a `//` prefix (GHSA-v753-g4cw-jm48)
		{
			whenURL:        "http://localhost:1323/%09/evil.example/x",
			expectPath:     "/\t/evil.example/x",
			expectLocation: []string{`/%09/evil.example/x/`},
		},
		{
			whenURL:        "http://localhost:1323/%09%5Cevil.example/x",
			expectPath:     "/\t\\evil.example/x",
			expectLocation: []string{`/%09%5Cevil.example/x/`},
		},
		{
			whenURL:        "http://localhost:1323/%09//evil.example/x",
			expectPath:     "/\t//evil.example/x",
			expectLocation: []string{`/%09//evil.example/x/`},
		},
		{
			whenURL:        "http://localhost:1323/%09%09/evil.example/x",
			expectPath:     "/\t\t/evil.example/x",
			expectLocation: []string{`/%09%09/evil.example/x/`},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			e := echo.New()

			mw := AddTrailingSlashWithConfig(AddTrailingSlashConfig{
				RedirectCode: http.StatusMovedPermanently,
			})
			h := mw(func(c *echo.Context) error {
				return nil
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			c := e.NewContext(req, rec)

			err := h(c)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectPath, req.URL.Path)
			assert.Equal(t, tc.expectLocation, rec.Header()[echo.HeaderLocation])
			if tc.expectStatus == 0 {
				assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			} else {
				assert.Equal(t, tc.expectStatus, rec.Code)
			}
		})
	}
}

func TestAddTrailingSlash(t *testing.T) {
	var testCases = []struct {
		whenURL        string
		whenMethod     string
		expectPath     string
		expectLocation []string
	}{
		{
			whenURL:    "/add-slash",
			whenMethod: http.MethodGet,
			expectPath: "/add-slash/",
		},
		{
			whenURL:    "/add-slash?key=value",
			whenMethod: http.MethodGet,
			expectPath: "/add-slash/",
		},
		{
			whenURL:        "/",
			whenMethod:     http.MethodConnect,
			expectPath:     "/",
			expectLocation: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			e := echo.New()

			h := AddTrailingSlash()(func(c *echo.Context) error {
				return nil
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			c := e.NewContext(req, rec)

			err := h(c)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectPath, req.URL.Path)
			assert.Equal(t, []string(nil), rec.Header()[echo.HeaderLocation])
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestRemoveTrailingSlashWithConfig(t *testing.T) {
	var testCases = []struct {
		whenURL        string
		whenMethod     string
		expectPath     string
		expectLocation []string
		expectStatus   int
	}{
		{
			whenURL:        "/remove-slash/",
			whenMethod:     http.MethodGet,
			expectPath:     "/remove-slash/",
			expectLocation: []string{`/remove-slash`},
		},
		{
			whenURL:        "/remove-slash/?key=value",
			whenMethod:     http.MethodGet,
			expectPath:     "/remove-slash/",
			expectLocation: []string{`/remove-slash?key=value`},
		},
		{
			whenURL:        "/",
			whenMethod:     http.MethodConnect,
			expectPath:     "/",
			expectLocation: nil,
			expectStatus:   http.StatusOK,
		},
		{
			whenURL:        "http://localhost",
			whenMethod:     http.MethodGet,
			expectPath:     "",
			expectLocation: nil,
			expectStatus:   http.StatusOK,
		},
		// cases for open redirect vulnerability
		{
			whenURL:        "http://localhost:1323/%5Cexample.com/",
			expectPath:     `/\example.com/`,
			expectLocation: []string{`/example.com`},
		},
		{
			whenURL:        `http://localhost:1323/\example.com/`,
			expectPath:     `/\example.com/`,
			expectLocation: []string{`/example.com`},
		},
		{
			whenURL:        `http://localhost:1323/\\%5C////%5C\\\example.com/`,
			expectPath:     `/\\\////\\\\example.com/`,
			expectLocation: []string{`/example.com`},
		},
		{
			whenURL:        "http://localhost:1323//example.com/",
			expectPath:     `//example.com/`,
			expectLocation: []string{`/example.com`},
		},
		{
			whenURL:        "http://localhost:1323/%5C%5C/",
			expectPath:     `/\\/`,
			expectLocation: []string{`/`},
		},
		// control characters that browsers strip from URLs must not hide a `//` prefix (GHSA-v753-g4cw-jm48)
		{
			whenURL:        "http://localhost:1323/%09/evil.example/x/",
			expectPath:     "/\t/evil.example/x/",
			expectLocation: []string{`/%09/evil.example/x`},
		},
		{
			whenURL:        "http://localhost:1323/%09%5Cevil.example/x/",
			expectPath:     "/\t\\evil.example/x/",
			expectLocation: []string{`/%09%5Cevil.example/x`},
		},
		{
			whenURL:        "http://localhost:1323/%09//evil.example/x/",
			expectPath:     "/\t//evil.example/x/",
			expectLocation: []string{`/%09//evil.example/x`},
		},
		{
			whenURL:        "http://localhost:1323/%09%09/evil.example/x/",
			expectPath:     "/\t\t/evil.example/x/",
			expectLocation: []string{`/%09%09/evil.example/x`},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			e := echo.New()

			mw := RemoveTrailingSlashWithConfig(RemoveTrailingSlashConfig{
				RedirectCode: http.StatusMovedPermanently,
			})
			h := mw(func(c *echo.Context) error {
				return nil
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			c := e.NewContext(req, rec)

			err := h(c)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectPath, req.URL.Path)
			assert.Equal(t, tc.expectLocation, rec.Header()[echo.HeaderLocation])
			if tc.expectStatus == 0 {
				assert.Equal(t, http.StatusMovedPermanently, rec.Code)
			} else {
				assert.Equal(t, tc.expectStatus, rec.Code)
			}
		})
	}
}

func TestRemoveTrailingSlash(t *testing.T) {
	var testCases = []struct {
		whenURL    string
		whenMethod string
		expectPath string
	}{
		{
			whenURL:    "/remove-slash/",
			whenMethod: http.MethodGet,
			expectPath: "/remove-slash",
		},
		{
			whenURL:    "/remove-slash/?key=value",
			whenMethod: http.MethodGet,
			expectPath: "/remove-slash",
		},
		{
			whenURL:    "/",
			whenMethod: http.MethodConnect,
			expectPath: "/",
		},
		{
			whenURL:    "http://localhost",
			whenMethod: http.MethodGet,
			expectPath: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			e := echo.New()

			h := RemoveTrailingSlash()(func(c *echo.Context) error {
				return nil
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			c := e.NewContext(req, rec)

			err := h(c)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectPath, req.URL.Path)
			assert.Equal(t, []string(nil), rec.Header()[echo.HeaderLocation])
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestSanitizeURI(t *testing.T) {
	var testCases = []struct {
		whenURI string
		expect  string
	}{
		{whenURI: "/path", expect: "/path"},
		{whenURI: "", expect: ""},
		{whenURI: "//example.com", expect: "/example.com"},
		{whenURI: `/\example.com`, expect: "/example.com"},
		{whenURI: "/\t/example.com", expect: "/%09/example.com"},
		{whenURI: "/\n/example.com", expect: "/%0A/example.com"},
		{whenURI: "/\r/example.com", expect: "/%0D/example.com"},
		{whenURI: "/\x7f/example.com", expect: "/%7F/example.com"},
		{whenURI: "\t//example.com", expect: "%09//example.com"},
		{whenURI: "/a\tb", expect: "/a%09b"},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURI, func(t *testing.T) {
			assert.Equal(t, tc.expect, sanitizeURI(tc.whenURI))
		})
	}
}

func TestTrailingSlashRedirectEscapedPath(t *testing.T) {
	for _, tc := range []struct{ path, add, remove string }{
		{"/files/a%2Fb", "/files/a%2Fb/", "/files/a%2Fb"},
		{"/files/a%3Fb?x=1", "/files/a%3Fb/?x=1", "/files/a%3Fb?x=1"},
		{"/files/a%20b", "/files/a%20b/", "/files/a%20b"},
		{"/files/a%23b", "/files/a%23b/", "/files/a%23b"},
		{"/%5Cexample.com", "/example.com/", "/example.com"},
		{"/%2f%2Fexample.com", "/example.com/", "/example.com"},
	} {
		for _, remove := range []bool{false, true} {
			name := "add"
			target, want := tc.path, tc.add
			mw := AddTrailingSlashWithConfig(AddTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently})
			if remove {
				name = "remove"
				target, want = tc.add, tc.remove
				// Keep the original escaping when constructing the remove request.
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)
				target = req.URL.EscapedPath() + "/"
				if req.URL.RawQuery != "" {
					target += "?" + req.URL.RawQuery
				}
				mw = RemoveTrailingSlashWithConfig(RemoveTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently})
			}
			t.Run(name+tc.path, func(t *testing.T) {
				e := echo.New()
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, target, nil)
				assert.NoError(t, mw(func(*echo.Context) error { return nil })(e.NewContext(req, rec)))
				assert.Equal(t, want, rec.Header().Get(echo.HeaderLocation))
			})
		}
	}
}

func TestRemoveTrailingSlashRedirectEncodedSuffix(t *testing.T) {
	for _, target := range []string{"/files/a%2Fb%2f", "/files/a%2Fb%2F"} {
		t.Run(target, func(t *testing.T) {
			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, target+"?x=1", nil)
			mw := RemoveTrailingSlashWithConfig(RemoveTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently})
			assert.NoError(t, mw(func(*echo.Context) error { return nil })(e.NewContext(req, rec)))
			assert.Equal(t, "/files/a%2Fb?x=1", rec.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestTrailingSlashRedirectPreservesRoute(t *testing.T) {
	for _, remove := range []bool{false, true} {
		name, target, route := "add", "/files/a%2Fb", "/files/:id/"
		mw := AddTrailingSlashWithConfig(AddTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently})
		if remove {
			name, target, route = "remove", "/files/a%2Fb/", "/files/:id"
			mw = RemoveTrailingSlashWithConfig(RemoveTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently})
		}
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			e.Pre(mw)
			e.GET(route, func(c *echo.Context) error { return c.String(http.StatusOK, c.Param("id")) })
			first := httptest.NewRecorder()
			e.ServeHTTP(first, httptest.NewRequest(http.MethodGet, target, nil))
			if !assert.Equal(t, http.StatusMovedPermanently, first.Code) {
				return
			}
			second := httptest.NewRecorder()
			e.ServeHTTP(second, httptest.NewRequest(http.MethodGet, first.Header().Get(echo.HeaderLocation), nil))
			assert.Equal(t, http.StatusOK, second.Code)
			assert.Equal(t, "a%2Fb", second.Body.String())
		})
	}
}

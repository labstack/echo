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
			// An escaped slash inside a path parameter is data, not a separator.
			// The redirect must keep %2F or following it matches a different route.
			whenURL:        "/files/a%2Fb",
			whenMethod:     http.MethodGet,
			expectPath:     "/files/a/b",
			expectLocation: []string{`/files/a%2Fb/`},
		},
		{
			// %3F must not be promoted into a query string by the redirect.
			whenURL:        "/files/a%3Fb",
			whenMethod:     http.MethodGet,
			expectPath:     "/files/a?b",
			expectLocation: []string{`/files/a%3Fb/`},
		},
		{
			// %20 must stay encoded; a raw space is not a valid Location.
			whenURL:        "/files/a%20b",
			whenMethod:     http.MethodGet,
			expectPath:     "/files/a b",
			expectLocation: []string{`/files/a%20b/`},
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
			// The escaped slash is part of the parameter value, so removing the
			// trailing slash must not decode it into a real separator.
			whenURL:        "/files/a%2Fb/",
			whenMethod:     http.MethodGet,
			expectPath:     "/files/a/b/",
			expectLocation: []string{`/files/a%2Fb`},
		},
		{
			whenURL:        "/files/a%3Fb/",
			whenMethod:     http.MethodGet,
			expectPath:     "/files/a?b/",
			expectLocation: []string{`/files/a%3Fb`},
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

// End-to-end: the redirect issued by AddTrailingSlash must be followable and land
// on the same route, rather than turning an escaped slash into a real separator.
func TestAddTrailingSlashRedirectIsFollowable(t *testing.T) {
	e := echo.New()
	e.Pre(AddTrailingSlashWithConfig(AddTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently}))
	e.GET("/files/:id/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "id="+c.Param("id"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/a%2Fb", nil)
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	loc := rec.Header().Get(echo.HeaderLocation)
	assert.Equal(t, "/files/a%2Fb/", loc)

	// follow it
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, loc, nil)
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "id=a%2Fb", rec2.Body.String())
}

func TestRemoveTrailingSlashRedirectIsFollowable(t *testing.T) {
	e := echo.New()
	e.Pre(RemoveTrailingSlashWithConfig(RemoveTrailingSlashConfig{RedirectCode: http.StatusMovedPermanently}))
	e.GET("/files/:id", func(c *echo.Context) error {
		return c.String(http.StatusOK, "id="+c.Param("id"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/a%2Fb/", nil)
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	loc := rec.Header().Get(echo.HeaderLocation)
	assert.Equal(t, "/files/a%2Fb", loc)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, loc, nil)
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "id=a%2Fb", rec2.Body.String())
}

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

type middlewareGenerator func() echo.MiddlewareFunc

func TestRedirectHTTPSRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "labstack.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "https://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "https://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "https://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "https://www.ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "labstack.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.labstack.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "https://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "https://ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.labstack.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "labstack.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(HTTPSNonWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "labstack.com",
			expectLocation:   "http://www.labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			expectLocation:   "http://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "http://www.ip/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "a.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://www.a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.ip",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(WWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	var testCases = []struct {
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.a.com",
			expectLocation:   "http://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "www.a.com",
			whenHeader:       map[string][]string{echo.HeaderXForwardedProto: {"https"}},
			expectLocation:   "https://a.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			whenHost:         "ip",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			res := redirectTest(NonWWWRedirect, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestNonWWWRedirectWithConfig(t *testing.T) {
	var testCases = []struct {
		name             string
		givenCode        int
		givenSkipFunc    func(c *echo.Context) bool
		whenHost         string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			name:             "usual redirect",
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name: "redirect is skipped",
			givenSkipFunc: func(c *echo.Context) bool {
				return true // skip always
			},
			whenHost:         "www.labstack.com",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "redirect with custom status code",
			givenCode:        http.StatusSeeOther,
			whenHost:         "www.labstack.com",
			expectLocation:   "http://labstack.com/",
			expectStatusCode: http.StatusSeeOther,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenHost, func(t *testing.T) {
			middleware := func() echo.MiddlewareFunc {
				return NonWWWRedirectWithConfig(RedirectConfig{
					Skipper: tc.givenSkipFunc,
					Code:    tc.givenCode,
				})
			}
			res := redirectTest(middleware, tc.whenHost, tc.whenHeader)

			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func redirectTest(fn middlewareGenerator, host string, header http.Header) *httptest.ResponseRecorder {
	e := echo.New()
	next := func(c *echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = host
	req.RemoteAddr = "10.0.0.1:1234" // request from a reverse proxy in a private network (trusted by default)
	if header != nil {
		req.Header = header
	}
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	fn()(next)(c)

	return res
}

func TestRedirectHTTPSRedirect_forwardedProtoTrust(t *testing.T) {
	var testCases = []struct {
		name             string
		whenRemoteAddr   string
		expectLocation   string
		expectStatusCode int
	}{
		{
			name:             "redirects when X-Forwarded-Proto comes from untrusted client",
			whenRemoteAddr:   "203.0.113.10:1234",
			expectLocation:   "https://labstack.com/",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name:             "does not redirect when X-Forwarded-Proto comes from trusted proxy",
			whenRemoteAddr:   "10.0.0.1:1234",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = "labstack.com"
			req.RemoteAddr = tc.whenRemoteAddr
			req.Header.Set(echo.HeaderXForwardedProto, "https")
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := HTTPSRedirect()(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectSecurityHardening(t *testing.T) {
	var testCases = []struct {
		name             string
		middleware       middlewareGenerator
		whenHost         string
		whenRequestURI   string
		whenHeader       http.Header
		expectLocation   string
		expectStatusCode int
	}{
		{
			name:             "NonWWWRedirect ignores bare www. host to prevent https:///attacker.com open redirect",
			middleware:       NonWWWRedirect,
			whenHost:         "www.",
			whenRequestURI:   "/attacker.com/login",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "HTTPSNonWWWRedirect ignores bare www. host to prevent https:///attacker.com open redirect",
			middleware:       HTTPSNonWWWRedirect,
			whenHost:         "www.",
			whenRequestURI:   "/attacker.com/login",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "HTTPSRedirect ignores empty host",
			middleware:       HTTPSRedirect,
			whenHost:         "",
			whenRequestURI:   "/attacker.com",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "HTTPSRedirect ignores host with authority or path delimiters",
			middleware:       HTTPSRedirect,
			whenHost:         "attacker.com@labstack.com",
			whenRequestURI:   "/",
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "NonWWWRedirect handles uppercase WWW. prefix",
			middleware:       NonWWWRedirect,
			whenHost:         "WWW.labstack.com",
			whenRequestURI:   "/docs",
			expectLocation:   "http://labstack.com/docs",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name:             "HTTPSWWWRedirect does not duplicate uppercase WWW. prefix",
			middleware:       HTTPSWWWRedirect,
			whenHost:         "WWW.labstack.com",
			whenRequestURI:   "/docs",
			expectLocation:   "https://www.labstack.com/docs",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name:             "HTTPSRedirect sanitizes double-slash and control chars in RequestURI",
			middleware:       HTTPSRedirect,
			whenHost:         "labstack.com",
			whenRequestURI:   "//attacker.com/phish",
			expectLocation:   "https://labstack.com/attacker.com/phish",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name:             "WWWRedirect normalizes ws scheme from X-Forwarded-Proto to http",
			middleware:       WWWRedirect,
			whenHost:         "labstack.com",
			whenRequestURI:   "/ws",
			whenHeader:       http.Header{echo.HeaderXForwardedProto: {"ws"}},
			expectLocation:   "http://www.labstack.com/ws",
			expectStatusCode: http.StatusMovedPermanently,
		},
		{
			name:             "HTTPSRedirect treats wss scheme from X-Forwarded-Proto as https",
			middleware:       HTTPSRedirect,
			whenHost:         "labstack.com",
			whenRequestURI:   "/ws",
			whenHeader:       http.Header{echo.HeaderXForwardedProto: {"wss"}},
			expectLocation:   "",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tc.whenHost
			req.RequestURI = tc.whenRequestURI
			req.RemoteAddr = "10.0.0.1:1234"
			if tc.whenHeader != nil {
				req.Header = tc.whenHeader
			}
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := tc.middleware()(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectStatusCode, res.Code)
			assert.Equal(t, tc.expectLocation, res.Header().Get(echo.HeaderLocation))
		})
	}
}


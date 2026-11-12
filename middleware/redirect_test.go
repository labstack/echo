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

func TestRedirectMalformedHost(t *testing.T) {
	for _, host := range []string{"", "www.", "www.:80", "WWW.", "attacker.com@example.com", "example.com/path", "example.com\\path", "example.com?x", "example.com#x", "example.com\t", "example.com%2fpath"} {
		for _, mw := range []struct {
			name string
			make func() echo.MiddlewareFunc
		}{{"https", HTTPSRedirect}, {"https-www", HTTPSWWWRedirect}, {"https-non-www", HTTPSNonWWWRedirect}, {"www", WWWRedirect}, {"non-www", NonWWWRedirect}} {
			t.Run(mw.name+"/"+host, func(t *testing.T) {
				e := echo.New()
				e.Pre(mw.make())
				e.GET("/*", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
				req := httptest.NewRequest(http.MethodGet, "/attacker.com/login", nil)
				req.Host = host
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				assert.Empty(t, rec.Header().Get(echo.HeaderLocation))
				assert.Equal(t, http.StatusOK, rec.Code)
			})
		}
	}
}

func TestRedirectBareWWWHostOverHTTP(t *testing.T) {
	e := echo.New()
	e.Pre(HTTPSNonWWWRedirect())
	e.GET("/*", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	server := httptest.NewServer(e)
	defer server.Close()
	req, err := http.NewRequest(http.MethodGet, server.URL+"/attacker.com/login", nil)
	if !assert.NoError(t, err) {
		return
	}
	req.Host = "www."
	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get(echo.HeaderLocation))
}

func TestRedirectRequestTarget(t *testing.T) {
	for _, tc := range []struct{ uri, want string }{
		{"", "https://example.com/files/a%2Fb?x=1"},
		{"http://example.com/files/a%2Fb?x=1", "https://example.com/files/a%2Fb?x=1"},
		{"//attacker.com/x", "https://example.com//attacker.com/x"},
		{"/x\t/y", "https://example.com/x%09/y"},
	} {
		t.Run(tc.uri, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "http://example.com/files/a%2Fb?x=1", nil)
			req.RequestURI = tc.uri
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			assert.NoError(t, HTTPSRedirect()(func(*echo.Context) error { return nil })(c))
			assert.Equal(t, tc.want, rec.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectValidAuthority(t *testing.T) {
	for _, host := range []string{"example.com", "example.com:8080", "[::1]", "[::1]:8080"} {
		t.Run(host, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/docs", nil)
			req.Host = host
			rec := httptest.NewRecorder()
			assert.NoError(t, HTTPSRedirect()(func(*echo.Context) error { return nil })(e.NewContext(req, rec)))
			assert.Equal(t, "https://"+host+"/docs", rec.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestRedirectWWWCaseAndWebSocketScheme(t *testing.T) {
	for _, tc := range []struct {
		name, host, proto, want string
		make                    func() echo.MiddlewareFunc
	}{
		{"strip uppercase", "WWW.example.com", "http", "http://example.com/docs", NonWWWRedirect},
		{"avoid duplicate", "WWW.example.com", "http", "https://www.example.com/docs", HTTPSWWWRedirect},
		{"already www", "WWW.example.com", "https", "", WWWRedirect},
		{"ws", "example.com", "ws", "http://www.example.com/docs", WWWRedirect},
		{"wss", "www.example.com", "wss", "https://example.com/docs", NonWWWRedirect},
		{"already secure", "example.com", "wss", "", HTTPSRedirect},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/docs", nil)
			req.Host = tc.host
			req.RemoteAddr = "127.0.0.1:1234"
			req.Header.Set(echo.HeaderXForwardedProto, tc.proto)
			rec := httptest.NewRecorder()
			assert.NoError(t, tc.make()(func(*echo.Context) error { return nil })(e.NewContext(req, rec)))
			assert.Equal(t, tc.want, rec.Header().Get(echo.HeaderLocation))
		})
	}
}

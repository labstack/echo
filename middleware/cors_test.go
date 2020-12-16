package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	e := echo.New()

	// Wildcard origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := CORS()(echo.NotFoundHandler)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Wildcard AllowedOrigin with no Origin header in request
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = CORS()(echo.NotFoundHandler)
	h(c)
	assert.NotContains(t, rec.Header(), echo.HeaderAccessControlAllowOrigin)

	// Allow origins
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})(echo.NotFoundHandler)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))

	// Preflight request
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cors := CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(echo.HeaderAccessControlMaxAge))

	// Preflight request with `AllowOrigins` *
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cors = CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(echo.HeaderAccessControlMaxAge))

	// Preflight request with Access-Control-Request-Headers
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderOrigin, "localhost")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, "Special-Request-Header")
	cors = CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"*"},
	})
	h = cors(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "Special-Request-Header", rec.Header().Get(echo.HeaderAccessControlAllowHeaders))
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderAccessControlAllowMethods))

	// Preflight request with `AllowOrigins` which allow all subdomains with *
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderOrigin, "http://aaa.example.com")
	cors = CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"http://*.example.com"},
	})
	h = cors(echo.NotFoundHandler)
	h(c)
	assert.Equal(t, "http://aaa.example.com", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	req.Header.Set(echo.HeaderOrigin, "http://bbb.example.com")
	h(c)
	assert.Equal(t, "http://bbb.example.com", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}

func Test_allowOriginScheme(t *testing.T) {
	tests := []struct {
		domain, pattern string
		expected        bool
	}{
		{
			domain:   "http://example.com",
			pattern:  "http://example.com",
			expected: true,
		},
		{
			domain:   "https://example.com",
			pattern:  "https://example.com",
			expected: true,
		},
		{
			domain:   "http://example.com",
			pattern:  "https://example.com",
			expected: false,
		},
		{
			domain:   "https://example.com",
			pattern:  "http://example.com",
			expected: false,
		},
	}

	e := echo.New()
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		req.Header.Set(echo.HeaderOrigin, tt.domain)
		cors := CORSWithConfig(CORSConfig{
			AllowOrigins: []string{tt.pattern},
		})
		h := cors(echo.NotFoundHandler)
		h(c)

		if tt.expected {
			assert.Equal(t, tt.domain, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
		} else {
			assert.NotContains(t, rec.Header(), echo.HeaderAccessControlAllowOrigin)
		}
	}
}

func Test_allowOriginSubdomain(t *testing.T) {
	tests := []struct {
		domain, pattern string
		expected        bool
	}{
		{
			domain:   "http://aaa.example.com",
			pattern:  "http://*.example.com",
			expected: true,
		},
		{
			domain:   "http://bbb.aaa.example.com",
			pattern:  "http://*.example.com",
			expected: true,
		},
		{
			domain:   "http://bbb.aaa.example.com",
			pattern:  "http://*.aaa.example.com",
			expected: true,
		},
		{
			domain:   "http://aaa.example.com:8080",
			pattern:  "http://*.example.com:8080",
			expected: true,
		},

		{
			domain:   "http://fuga.hoge.com",
			pattern:  "http://*.example.com",
			expected: false,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://*.aaa.example.com",
			expected: false,
		},
		{
			domain: `http://1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
		  .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
		  .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
		  .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.example.com`,
			pattern:  "http://*.example.com",
			expected: false,
		},
		{
			domain:   `http://1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.example.com`,
			pattern:  "http://*.example.com",
			expected: false,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://example.com",
			expected: false,
		},
		{
			domain:   "https://prod-preview--aaa.bbb.com",
			pattern:  "https://*--aaa.bbb.com",
			expected: true,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://*.example.com",
			expected: true,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://foo.[a-z]*.example.com",
			expected: false,
		},
	}

	e := echo.New()
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		req.Header.Set(echo.HeaderOrigin, tt.domain)
		cors := CORSWithConfig(CORSConfig{
			AllowOrigins: []string{tt.pattern},
		})
		h := cors(echo.NotFoundHandler)
		h(c)

		if tt.expected {
			assert.Equal(t, tt.domain, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
		} else {
			assert.NotContains(t, rec.Header(), echo.HeaderAccessControlAllowOrigin)
		}
	}
}

func TestCorsHeaders(t *testing.T) {
	tests := []struct {
		domain, allowedOrigin, method string
		expected                      bool
	}{
		{
			domain:        "", // Request does not have Origin header
			allowedOrigin: "*",
			method:        http.MethodGet,
			expected:      false,
		},
		{
			domain:        "http://example.com",
			allowedOrigin: "*",
			method:        http.MethodGet,
			expected:      true,
		},
		{
			domain:        "", // Request does not have Origin header
			allowedOrigin: "http://example.com",
			method:        http.MethodGet,
			expected:      false,
		},
		{
			domain:        "http://bar.com",
			allowedOrigin: "http://example.com",
			method:        http.MethodGet,
			expected:      false,
		},
		{
			domain:        "http://example.com",
			allowedOrigin: "http://example.com",
			method:        http.MethodGet,
			expected:      true,
		},
		{
			domain:        "", // Request does not have Origin header
			allowedOrigin: "*",
			method:        http.MethodOptions,
			expected:      false,
		},
		{
			domain:        "http://example.com",
			allowedOrigin: "*",
			method:        http.MethodOptions,
			expected:      true,
		},
		{
			domain:        "", // Request does not have Origin header
			allowedOrigin: "http://example.com",
			method:        http.MethodOptions,
			expected:      false,
		},
		{
			domain:        "http://bar.com",
			allowedOrigin: "http://example.com",
			method:        http.MethodGet,
			expected:      false,
		},
		{
			domain:        "http://example.com",
			allowedOrigin: "http://example.com",
			method:        http.MethodOptions,
			expected:      true,
		},
	}

	e := echo.New()
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if tt.domain != "" {
			req.Header.Set(echo.HeaderOrigin, tt.domain)
		}
		cors := CORSWithConfig(CORSConfig{
			AllowOrigins: []string{tt.allowedOrigin},
			//AllowCredentials: true,
			//MaxAge:           3600,
		})
		h := cors(echo.NotFoundHandler)
		h(c)

		assert.Equal(t, echo.HeaderOrigin, rec.Header().Get(echo.HeaderVary))

		expectedAllowOrigin := ""
		if tt.allowedOrigin == "*" {
			expectedAllowOrigin = "*"
		} else {
			expectedAllowOrigin = tt.domain
		}

		switch {
		case tt.expected && tt.method == http.MethodOptions:
			assert.Contains(t, rec.Header(), echo.HeaderAccessControlAllowMethods)
			assert.Equal(t, expectedAllowOrigin, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
			assert.Equal(t, 3, len(rec.Header()[echo.HeaderVary]))
		case tt.expected && tt.method == http.MethodGet:
			assert.Equal(t, expectedAllowOrigin, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
			assert.Equal(t, 1, len(rec.Header()[echo.HeaderVary])) // Vary: Origin
		default:
			assert.NotContains(t, rec.Header(), echo.HeaderAccessControlAllowOrigin)
			assert.Equal(t, 1, len(rec.Header()[echo.HeaderVary])) // Vary: Origin
		}

		if tt.method == http.MethodOptions {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
	}
}

func Test_allowOriginFunc(t *testing.T) {
	returnTrue := func(origin string) (bool, error) {
		return true, nil
	}
	returnFalse := func(origin string) (bool, error) {
		return false, nil
	}
	returnError := func(origin string) (bool, error) {
		return true, errors.New("this is a test error")
	}

	allowOriginFuncs := []func(origin string) (bool, error){
		returnTrue,
		returnFalse,
		returnError,
	}

	const origin = "http://example.com"

	e := echo.New()
	for _, allowOriginFunc := range allowOriginFuncs {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		req.Header.Set(echo.HeaderOrigin, origin)
		cors := CORSWithConfig(CORSConfig{
			AllowOriginFunc: allowOriginFunc,
		})
		h := cors(echo.NotFoundHandler)
		err := h(c)

		expected, expectedErr := allowOriginFunc(origin)
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Equal(t, "", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
			continue
		}

		if expected {
			assert.Equal(t, origin, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
		} else {
			assert.Equal(t, "", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
		}
	}
}

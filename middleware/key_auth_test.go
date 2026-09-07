// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func testKeyValidator(c *echo.Context, key string, source ExtractorSource) (bool, error) {
	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(key), []byte("valid-key")) == 1 {
		return true, nil
	}

	// Special case for testing error handling
	if key == "error-key" { // Error path doesn't need constant-time
		return false, errors.New("some user defined error")
	}

	return false, nil
}

func TestKeyAuth(t *testing.T) {
	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "test")
	}
	middlewareChain := KeyAuth(testKeyValidator)(handler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid-key")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := middlewareChain(c)

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestKeyAuthWithConfig(t *testing.T) {
	var testCases = []struct {
		name                string
		givenRequestFunc    func() *http.Request
		givenRequest        func(req *http.Request)
		whenConfig          func(conf *KeyAuthConfig)
		expectHandlerCalled bool
		expectError         string
	}{
		{
			name: "ok, defaults, key from header",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer valid-key")
			},
			expectHandlerCalled: true,
		},
		{
			name: "ok, custom skipper",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer error-key")
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.Skipper = func(context *echo.Context) bool {
					return true
				}
			},
			expectHandlerCalled: true,
		},
		{
			name: "nok, defaults, invalid key from header, Authorization: Bearer",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer invalid-key")
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=Unauthorized, err=code=401, message=invalid key",
		},
		{
			name: "nok, defaults, invalid scheme in header",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bear valid-key")
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=invalid value in request header",
		},
		{
			name:                "nok, defaults, missing header",
			givenRequest:        func(req *http.Request) {},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=missing value in request header",
		},
		{
			name: "ok, custom key lookup, header",
			givenRequest: func(req *http.Request) {
				req.Header.Set("API-Key", "valid-key")
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "header:API-Key"
			},
			expectHandlerCalled: true,
		},
		{
			name: "nok, custom key lookup, missing header",
			givenRequest: func(req *http.Request) {
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "header:API-Key"
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=missing value in request header",
		},
		{
			name: "ok, custom key lookup, query",
			givenRequest: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("key", "valid-key")
				req.URL.RawQuery = q.Encode()
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "query:key"
			},
			expectHandlerCalled: true,
		},
		{
			name: "nok, custom key lookup, missing query param",
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "query:key"
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=missing value in the query string",
		},
		{
			name: "ok, custom key lookup, form",
			givenRequestFunc: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("key=valid-key"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
				return req
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "form:key"
			},
			expectHandlerCalled: true,
		},
		{
			name: "nok, custom key lookup, missing key in form",
			givenRequestFunc: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("xxx=valid-key"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
				return req
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "form:key"
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=missing value in the form",
		},
		{
			name: "ok, custom key lookup, cookie",
			givenRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "key",
					Value: "valid-key",
				})
				q := req.URL.Query()
				q.Add("key", "valid-key")
				req.URL.RawQuery = q.Encode()
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "cookie:key"
			},
			expectHandlerCalled: true,
		},
		{
			name: "nok, custom key lookup, missing cookie param",
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "cookie:key"
			},
			expectHandlerCalled: false,
			expectError:         "code=401, message=missing key, err=missing value in cookies",
		},
		{
			name: "nok, custom errorHandler, error from extractor",
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "header:token"
				conf.ErrorHandler = func(c *echo.Context, err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "custom").Wrap(err)
				}
			},
			expectHandlerCalled: false,
			expectError:         "code=418, message=custom, err=missing value in request header",
		},
		{
			name: "nok, custom errorHandler, error from validator",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer error-key")
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.ErrorHandler = func(c *echo.Context, err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "custom").Wrap(err)
				}
			},
			expectHandlerCalled: false,
			expectError:         "code=418, message=custom, err=some user defined error",
		},
		{
			name: "nok, defaults, error from validator",
			givenRequest: func(req *http.Request) {
				req.Header.Set(echo.HeaderAuthorization, "Bearer error-key")
			},
			whenConfig:          func(conf *KeyAuthConfig) {},
			expectHandlerCalled: false,
			expectError:         "code=401, message=Unauthorized, err=some user defined error",
		},
		{
			name: "ok, custom validator checks source",
			givenRequest: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("key", "valid-key")
				req.URL.RawQuery = q.Encode()
			},
			whenConfig: func(conf *KeyAuthConfig) {
				conf.KeyLookup = "query:key"
				conf.Validator = func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
					if source == ExtractorSourceQuery {
						return true, nil
					}
					return false, errors.New("invalid source")
				}

			},
			expectHandlerCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			handler := func(c *echo.Context) error {
				handlerCalled = true
				return c.String(http.StatusOK, "test")
			}
			config := KeyAuthConfig{
				Validator: testKeyValidator,
			}
			if tc.whenConfig != nil {
				tc.whenConfig(&config)
			}
			middlewareChain := KeyAuthWithConfig(config)(handler)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.givenRequestFunc != nil {
				req = tc.givenRequestFunc()
			}
			if tc.givenRequest != nil {
				tc.givenRequest(req)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := middlewareChain(c)

			assert.Equal(t, tc.expectHandlerCalled, handlerCalled)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyAuthWithConfig_errors(t *testing.T) {
	var testCases = []struct {
		name        string
		whenConfig  KeyAuthConfig
		expectError string
	}{
		{
			name: "ok, no error",
			whenConfig: KeyAuthConfig{
				Validator: func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
					return false, nil
				},
			},
		},
		{
			name: "ok, missing validator func",
			whenConfig: KeyAuthConfig{
				Validator: nil,
			},
			expectError: "echo key-auth middleware requires a validator function",
		},
		{
			name: "ok, extractor source can not be split",
			whenConfig: KeyAuthConfig{
				KeyLookup: "nope",
				Validator: func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
					return false, nil
				},
			},
			expectError: "echo key-auth middleware could not create key extractor: extractor source for lookup could not be split into needed parts: nope",
		},
		{
			name: "ok, no extractors",
			whenConfig: KeyAuthConfig{
				KeyLookup: "nope:nope",
				Validator: func(c *echo.Context, key string, source ExtractorSource) (bool, error) {
					return false, nil
				},
			},
			expectError: "echo key-auth middleware could not create extractors from KeyLookup string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mw, err := tc.whenConfig.ToMiddleware()
			if tc.expectError != "" {
				assert.Nil(t, mw)
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NotNil(t, mw)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMustKeyAuthWithConfig_panic(t *testing.T) {
	assert.Panics(t, func() {
		KeyAuthWithConfig(KeyAuthConfig{})
	})
}

func TestKeyAuth_errorHandlerSwallowsError(t *testing.T) {
	handlerCalled := false
	var authValue string
	handler := func(c *echo.Context) error {
		handlerCalled = true
		authValue = c.Get("auth").(string)
		return c.String(http.StatusOK, "test")
	}
	middlewareChain := KeyAuthWithConfig(KeyAuthConfig{
		Validator: testKeyValidator,
		ErrorHandler: func(c *echo.Context, err error) error {
			// could check error to decide if we can swallow the error
			c.Set("auth", "public")
			return nil
		},
		ContinueOnIgnoredError: true,
	})(handler)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// no auth header this time
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := middlewareChain(c)

	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "public", authValue)
}

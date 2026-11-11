// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestMethodOverride(t *testing.T) {
	e := echo.New()
	m := MethodOverride()
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with http header
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	req.Header.Set(echo.HeaderXHTTPMethodOverride, http.MethodDelete)
	c := e.NewContext(req, rec)

	err := m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)

}

func TestMethodOverride_formParam(t *testing.T) {
	e := echo.New()
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with form parameter
	m, err := MethodOverrideConfig{Getter: MethodFromForm("_method")}.ToMiddleware()
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("_method="+http.MethodDelete)))
	rec := httptest.NewRecorder()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, rec)

	err = m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)
}

func TestMethodOverride_queryParam(t *testing.T) {
	e := echo.New()
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Override with query parameter
	m, err := MethodOverrideConfig{Getter: MethodFromQuery("_method")}.ToMiddleware()
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/?_method="+http.MethodDelete, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodDelete, req.Method)
}

func TestMethodOverride_ignoreGet(t *testing.T) {
	e := echo.New()
	m := MethodOverride()
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Ignore `GET`
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXHTTPMethodOverride, http.MethodDelete)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := m(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodGet, req.Method)
}

func TestMethodOverride_ignoresSafeTargetMethods(t *testing.T) {
	// Overriding POST to a safe method would let a cross-site form POST skip checks that only apply to
	// state-changing methods, such as the CSRF middleware (GHSA-r7w9-592q-9vg4).
	var testCases = []struct {
		whenMethod   string
		expectMethod string
	}{
		{whenMethod: http.MethodGet, expectMethod: http.MethodPost},
		{whenMethod: "get", expectMethod: http.MethodPost},
		{whenMethod: http.MethodHead, expectMethod: http.MethodPost},
		{whenMethod: http.MethodOptions, expectMethod: http.MethodPost},
		{whenMethod: http.MethodTrace, expectMethod: http.MethodPost},
		{whenMethod: http.MethodConnect, expectMethod: http.MethodPost},
		{whenMethod: http.MethodPut, expectMethod: http.MethodPut},
		{whenMethod: http.MethodPatch, expectMethod: http.MethodPatch},
		{whenMethod: http.MethodDelete, expectMethod: http.MethodDelete},
		{whenMethod: "PROPFIND", expectMethod: "PROPFIND"},
	}

	for _, tc := range testCases {
		t.Run(tc.whenMethod, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderXHTTPMethodOverride, tc.whenMethod)
			c := e.NewContext(req, httptest.NewRecorder())

			err := MethodOverride()(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectMethod, req.Method)
		})
	}
}

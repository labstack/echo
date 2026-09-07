// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContextResetClearsStore guards the clear(c.store) reuse: a pooled Context must not leak store
// values from a previous request into the next one, and Set must still work after a clear-based Reset.
func TestContextResetClearsStore(t *testing.T) {
	e := New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	c.Set("secret", "req1")
	assert.Equal(t, "req1", c.Get("secret"))

	c.Reset(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	assert.Nil(t, c.Get("secret"), "store must not leak across Reset")

	c.Set("k", "req2") // Set must still work after clear-based reset
	assert.Equal(t, "req2", c.Get("k"))
}

// TestContextJSONStatusAcrossReset guards the reused delayedStatusWriter (c.dsw): a second JSON
// response on a pooled+Reset Context must use the new status, not inherit the previous one.
func TestContextJSONStatusAcrossReset(t *testing.T) {
	e := New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	assert.NoError(t, c.JSON(http.StatusTeapot, map[string]int{"a": 1}))

	rec2 := httptest.NewRecorder()
	c.Reset(httptest.NewRequest(http.MethodGet, "/", nil), rec2)
	assert.NoError(t, c.JSON(http.StatusCreated, map[string]int{"b": 2}))
	assert.Equal(t, http.StatusCreated, rec2.Code)
	assert.JSONEq(t, `{"b":2}`, rec2.Body.String())
}

// TestNestedJSONUsesFreshDelayedWriter guards the nested c.JSON case: a serializer that calls c.JSON
// re-entrantly must not corrupt the outer delayedStatusWriter (regression test for c.dsw self-reference).
func TestNestedJSONUsesFreshDelayedWriter(t *testing.T) {
	e := New()
	e.JSONSerializer = nestedJSONSerializer{}
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	assert.NoError(t, c.JSON(http.StatusOK, map[string]int{"outer": 1}))
	assert.Equal(t, http.StatusOK, rec.Code)
}

type nestedJSONSerializer struct{}

func (nestedJSONSerializer) Serialize(c *Context, i any, indent string) error {
	if m, ok := i.(map[string]int); ok && m["outer"] == 1 {
		// re-enter c.JSON once before encoding the outer payload
		if err := c.JSON(http.StatusOK, map[string]int{"inner": 2}); err != nil {
			return err
		}
	}
	return (DefaultJSONSerializer{}).Serialize(c, i, indent)
}

func (nestedJSONSerializer) Deserialize(c *Context, i any) error {
	return (DefaultJSONSerializer{}).Deserialize(c, i)
}

// TestGlobalMiddlewareRunsOnNotFoundAndMethodNotAllowed pins the dispatch contract: global (Use) and
// pre (Pre) middleware must execute even when the router returns 404 / 405 / OPTIONS handlers.
func TestGlobalMiddlewareRunsOnNotFoundAndMethodNotAllowed(t *testing.T) {
	cases := []struct {
		name, method, path string
		code               int
	}{
		{"404", http.MethodGet, "/missing", http.StatusNotFound},
		{"405", http.MethodPost, "/", http.StatusMethodNotAllowed},
		{"OPTIONS", http.MethodOptions, "/", http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			var pre, use bool
			e.Pre(func(n HandlerFunc) HandlerFunc {
				return func(c *Context) error { pre = true; return n(c) }
			})
			e.Use(func(n HandlerFunc) HandlerFunc {
				return func(c *Context) error { use = true; return n(c) }
			})
			e.GET("/", func(c *Context) error { return c.String(http.StatusOK, "ok") })

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			assert.True(t, pre, "pre-middleware must run on %s", tc.name)
			assert.True(t, use, "global middleware must run on %s", tc.name)
			assert.Equal(t, tc.code, rec.Code)
		})
	}
}

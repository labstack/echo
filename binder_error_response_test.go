// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regression test for #2771: a BindingError returned from a handler must be
// serialized by DefaultHTTPErrorHandler into a structured response that retains
// the field name (and the binder message), not flattened to {"message":"Bad Request"}.
func TestBindingError_serializesToStructuredJSON(t *testing.T) {
	e := New()
	e.GET("/doc", func(c *Context) error {
		var docNum int
		return QueryParamsBinder(c).MustInt("docNum", &docNum).BindError()
	})

	req := httptest.NewRequest(http.MethodGet, "/doc?docNum=abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "docNum", body["field"], "binding error response must retain the field name")
	assert.Equal(t, "failed to bind field value to int", body["message"], "binding error response must retain the binder message")
}

// When the binding error carries no message, MarshalJSON falls back to the
// status text (mirroring DefaultHTTPErrorHandler's *HTTPError branch).
func TestBindingError_marshalJSON_emptyMessageFallsBackToStatusText(t *testing.T) {
	be := &BindingError{Field: "name", HTTPError: &HTTPError{Code: http.StatusBadRequest}}

	b, err := be.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"field":"name","message":"Bad Request"}`, string(b))
}

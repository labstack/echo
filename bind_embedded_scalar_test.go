// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type EmbeddedIdentifier int

func TestBindEmbeddedScalarPointer(t *testing.T) {
	identifier := EmbeddedIdentifier(7)
	target := struct {
		*EmbeddedIdentifier `query:"id"`
	}{EmbeddedIdentifier: &identifier}
	c := New().NewContext(httptest.NewRequest(http.MethodGet, "/?id=42", nil), httptest.NewRecorder())

	if assert.NoError(t, c.Bind(&target)) {
		assert.Equal(t, EmbeddedIdentifier(42), *target.EmbeddedIdentifier)
	}
}

func TestBindEmbeddedScalarValue(t *testing.T) {
	target := struct {
		EmbeddedIdentifier `query:"id"`
	}{}
	c := New().NewContext(httptest.NewRequest(http.MethodGet, "/?id=42", nil), httptest.NewRecorder())

	if assert.NoError(t, c.Bind(&target)) {
		assert.Equal(t, EmbeddedIdentifier(42), target.EmbeddedIdentifier)
	}
}

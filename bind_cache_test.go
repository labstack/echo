// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBindCachedMetaPreservesFieldNameError ensures the per-type bind metadata cache preserves the
// field-name prefix in conversion errors on BOTH the cold (first) and warm (cached) bind of a type.
// DTO is declared locally so its reflect.Type is independent of suite ordering, making the second
// bind a deterministic cache hit (the bindMetaFor Load branch).
func TestBindCachedMetaPreservesFieldNameError(t *testing.T) {
	type DTO struct {
		Number int `query:"number"`
	}
	bind := func() error {
		e := New()
		req := httptest.NewRequest(http.MethodGet, "/?number=10a", nil)
		var dto DTO
		return e.NewContext(req, httptest.NewRecorder()).Bind(&dto)
	}

	assert.ErrorContains(t, bind(), "number", "cold cache: error must carry field name")
	assert.ErrorContains(t, bind(), "number", "warm cache: error must still carry field name")
}

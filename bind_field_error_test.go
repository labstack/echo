// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regression test for #2629: when binding form data fails a type conversion, the
// returned error must identify which field failed (so applications can render a
// useful message), instead of a bare strconv error with no field context.
func TestBind_formConversionErrorIncludesFieldName(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("number=10a"))
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	c := e.NewContext(req, httptest.NewRecorder())

	type DTO struct {
		Number int `form:"number"`
	}
	var dto DTO
	err := c.Bind(&dto)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "number", "bind error must identify the failing field")
}

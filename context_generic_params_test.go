// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build go1.27

package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextParsePathParam(t *testing.T) {
	c := NewContext(nil, nil)
	c.SetPathValues(PathValues{
		{Name: "id", Value: "42"},
		{Name: "empty", Value: ""},
		{Name: "invalid", Value: "not-an-int"},
	})

	value, err := c.ParsePathParam[int]("id")
	assert.NoError(t, err)
	assert.Equal(t, 42, value)

	value, err = c.ParsePathParam[int]("missing")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Zero(t, value)

	value, err = c.ParsePathParam[int]("invalid")
	assert.ErrorContains(t, err, "message=path value")
	assert.Zero(t, value)

	value, err = c.ParsePathParamOr[int]("missing", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)

	value, err = c.ParsePathParamOr[int]("empty", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)
}

func TestContextParseQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=42&empty=&invalid=not-an-int&date=2026-08-24", nil)
	c := NewContext(req, nil)

	value, err := c.ParseQueryParam[int]("page")
	assert.NoError(t, err)
	assert.Equal(t, 42, value)

	value, err = c.ParseQueryParam[int]("missing")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Zero(t, value)

	value, err = c.ParseQueryParam[int]("invalid")
	assert.ErrorContains(t, err, "message=query param")
	assert.Zero(t, value)

	value, err = c.ParseQueryParamOr[int]("missing", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)

	value, err = c.ParseQueryParamOr[int]("empty", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)

	date, err := c.ParseQueryParam[time.Time]("date", TimeLayout(time.DateOnly))
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC), date)
}

func TestContextParseQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&id=2&id=3&invalid=1&invalid=not-an-int", nil)
	c := NewContext(req, nil)

	values, err := c.ParseQueryParams[int]("id")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, values)

	values, err = c.ParseQueryParams[int]("missing")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Nil(t, values)

	values, err = c.ParseQueryParams[int]("invalid")
	assert.ErrorContains(t, err, "message=query params")
	assert.Nil(t, values)

	values, err = c.ParseQueryParamsOr[int]("missing", []int{98, 99})
	assert.NoError(t, err)
	assert.Equal(t, []int{98, 99}, values)
}

func TestContextParseFormValue(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("count=42&empty=&invalid=not-an-int"),
	)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	c := NewContext(req, nil)

	value, err := c.ParseFormValue[int]("count")
	assert.NoError(t, err)
	assert.Equal(t, 42, value)

	value, err = c.ParseFormValue[int]("missing")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Zero(t, value)

	value, err = c.ParseFormValue[int]("invalid")
	assert.ErrorContains(t, err, "message=form value")
	assert.Zero(t, value)

	value, err = c.ParseFormValueOr[int]("missing", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)

	value, err = c.ParseFormValueOr[int]("empty", 99)
	assert.NoError(t, err)
	assert.Equal(t, 99, value)
}

func TestContextParseFormValues(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("id=1&id=2&id=3&invalid=1&invalid=not-an-int"),
	)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	c := NewContext(req, nil)

	values, err := c.ParseFormValues[int]("id")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, values)

	values, err = c.ParseFormValues[int]("missing")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Nil(t, values)

	values, err = c.ParseFormValues[int]("invalid")
	assert.ErrorContains(t, err, "message=form values")
	assert.Nil(t, values)

	values, err = c.ParseFormValuesOr[int]("missing", []int{98, 99})
	assert.NoError(t, err)
	assert.Equal(t, []int{98, 99}, values)
}

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

	var testCases = []struct {
		name            string
		parse           func() (any, error)
		want            any
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "value",
			parse: func() (any, error) { return c.ParsePathParam[int]("id") },
			want:  42,
		},
		{
			name:    "missing",
			parse:   func() (any, error) { return c.ParsePathParam[int]("missing") },
			want:    0,
			wantErr: ErrNonExistentKey,
		},
		{
			name:            "invalid",
			parse:           func() (any, error) { return c.ParsePathParam[int]("invalid") },
			want:            0,
			wantErrContains: "message=path value",
		},
		{
			name:  "missing with default",
			parse: func() (any, error) { return c.ParsePathParamOr[int]("missing", 99) },
			want:  99,
		},
		{
			name:  "empty with default",
			parse: func() (any, error) { return c.ParsePathParamOr[int]("empty", 99) },
			want:  99,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse()
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrContains != "":
				assert.ErrorContains(t, err, tc.wantErrContains)
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContextParseQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=42&empty=&invalid=not-an-int&date=2026-08-24", nil)
	c := NewContext(req, nil)

	var testCases = []struct {
		name            string
		parse           func() (any, error)
		want            any
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "value",
			parse: func() (any, error) { return c.ParseQueryParam[int]("page") },
			want:  42,
		},
		{
			name:    "missing",
			parse:   func() (any, error) { return c.ParseQueryParam[int]("missing") },
			want:    0,
			wantErr: ErrNonExistentKey,
		},
		{
			name:            "invalid",
			parse:           func() (any, error) { return c.ParseQueryParam[int]("invalid") },
			want:            0,
			wantErrContains: "message=query param",
		},
		{
			name:  "missing with default",
			parse: func() (any, error) { return c.ParseQueryParamOr[int]("missing", 99) },
			want:  99,
		},
		{
			name:  "empty with default",
			parse: func() (any, error) { return c.ParseQueryParamOr[int]("empty", 99) },
			want:  99,
		},
		{
			name:  "time layout option",
			parse: func() (any, error) { return c.ParseQueryParam[time.Time]("date", TimeLayout(time.DateOnly)) },
			want:  time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse()
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrContains != "":
				assert.ErrorContains(t, err, tc.wantErrContains)
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContextParseQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&id=2&id=3&invalid=1&invalid=not-an-int", nil)
	c := NewContext(req, nil)

	var testCases = []struct {
		name            string
		parse           func() (any, error)
		want            any
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "values",
			parse: func() (any, error) { return c.ParseQueryParams[int]("id") },
			want:  []int{1, 2, 3},
		},
		{
			name:    "missing",
			parse:   func() (any, error) { return c.ParseQueryParams[int]("missing") },
			want:    []int(nil),
			wantErr: ErrNonExistentKey,
		},
		{
			name:            "invalid",
			parse:           func() (any, error) { return c.ParseQueryParams[int]("invalid") },
			want:            []int(nil),
			wantErrContains: "message=query params",
		},
		{
			name:  "missing with default",
			parse: func() (any, error) { return c.ParseQueryParamsOr[int]("missing", []int{98, 99}) },
			want:  []int{98, 99},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse()
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrContains != "":
				assert.ErrorContains(t, err, tc.wantErrContains)
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContextParseFormValue(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("count=42&empty=&invalid=not-an-int"),
	)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	c := NewContext(req, nil)

	var testCases = []struct {
		name            string
		parse           func() (any, error)
		want            any
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "value",
			parse: func() (any, error) { return c.ParseFormValue[int]("count") },
			want:  42,
		},
		{
			name:    "missing",
			parse:   func() (any, error) { return c.ParseFormValue[int]("missing") },
			want:    0,
			wantErr: ErrNonExistentKey,
		},
		{
			name:            "invalid",
			parse:           func() (any, error) { return c.ParseFormValue[int]("invalid") },
			want:            0,
			wantErrContains: "message=form value",
		},
		{
			name:  "missing with default",
			parse: func() (any, error) { return c.ParseFormValueOr[int]("missing", 99) },
			want:  99,
		},
		{
			name:  "empty with default",
			parse: func() (any, error) { return c.ParseFormValueOr[int]("empty", 99) },
			want:  99,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse()
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrContains != "":
				assert.ErrorContains(t, err, tc.wantErrContains)
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContextParseFormValues(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("id=1&id=2&id=3&invalid=1&invalid=not-an-int"),
	)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	c := NewContext(req, nil)

	var testCases = []struct {
		name            string
		parse           func() (any, error)
		want            any
		wantErr         error
		wantErrContains string
	}{
		{
			name:  "values",
			parse: func() (any, error) { return c.ParseFormValues[int]("id") },
			want:  []int{1, 2, 3},
		},
		{
			name:    "missing",
			parse:   func() (any, error) { return c.ParseFormValues[int]("missing") },
			want:    []int(nil),
			wantErr: ErrNonExistentKey,
		},
		{
			name:            "invalid",
			parse:           func() (any, error) { return c.ParseFormValues[int]("invalid") },
			want:            []int(nil),
			wantErrContains: "message=form values",
		},
		{
			name:  "missing with default",
			parse: func() (any, error) { return c.ParseFormValuesOr[int]("missing", []int{98, 99}) },
			want:  []int{98, 99},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse()
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrContains != "":
				assert.ErrorContains(t, err, tc.wantErrContains)
			default:
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

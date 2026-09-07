// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_StatusCode(t *testing.T) {
	var err error = &HTTPError{Code: http.StatusBadRequest, Message: "my error message"}

	code := 0
	var sc HTTPStatusCoder
	if errors.As(err, &sc) {
		code = sc.StatusCode()
	}
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestHTTPError_Error(t *testing.T) {
	var testCases = []struct {
		name   string
		error  error
		expect string
	}{
		{
			name:   "ok, without message",
			error:  &HTTPError{Code: http.StatusBadRequest},
			expect: "code=400, message=Bad Request",
		},
		{
			name:   "ok, with message",
			error:  &HTTPError{Code: http.StatusBadRequest, Message: "my error message"},
			expect: "code=400, message=my error message",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.error.Error())
		})
	}
}

func TestHTTPError_WrapUnwrap(t *testing.T) {
	err := &HTTPError{Code: http.StatusBadRequest, Message: "bad"}
	wrapped := err.Wrap(errors.New("my_error")).(*HTTPError)

	err.Code = http.StatusOK
	err.Message = "changed"

	assert.Equal(t, http.StatusBadRequest, wrapped.Code)
	assert.Equal(t, "bad", wrapped.Message)

	assert.Equal(t, errors.New("my_error"), wrapped.Unwrap())
	assert.Equal(t, "code=400, message=bad, err=my_error", wrapped.Error())
}

func TestNewHTTPError(t *testing.T) {
	err := NewHTTPError(http.StatusBadRequest, "bad")
	err2 := &HTTPError{Code: http.StatusBadRequest, Message: "bad"}

	assert.Equal(t, err2, err)
}

func TestStatusCode(t *testing.T) {
	var testCases = []struct {
		name   string
		err    error
		expect int
	}{
		{
			name:   "ok, HTTPError",
			err:    &HTTPError{Code: http.StatusNotFound},
			expect: http.StatusNotFound,
		},
		{
			name:   "ok, sentinel error",
			err:    ErrNotFound,
			expect: http.StatusNotFound,
		},
		{
			name:   "ok, wrapped HTTPError",
			err:    fmt.Errorf("wrapped: %w", &HTTPError{Code: http.StatusTeapot}),
			expect: http.StatusTeapot,
		},
		{
			name:   "nok, normal error",
			err:    errors.New("error"),
			expect: 0,
		},
		{
			name:   "nok, nil",
			err:    nil,
			expect: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, StatusCode(tc.err))
		})
	}
}

func TestResolveResponseStatus(t *testing.T) {
	someErr := errors.New("some error")

	var testCases = []struct {
		name         string
		whenResp     http.ResponseWriter
		whenErr      error
		expectStatus int
		expectResp   bool
	}{
		{
			name:         "nil resp, nil err -> 200",
			whenResp:     nil,
			whenErr:      nil,
			expectStatus: http.StatusOK,
			expectResp:   false,
		},
		{
			name:         "resp suggested status used when no error",
			whenResp:     &Response{Status: http.StatusCreated},
			whenErr:      nil,
			expectStatus: http.StatusCreated,
			expectResp:   true,
		},
		{
			name:         "error overrides suggested status with StatusCode(err)",
			whenResp:     &Response{Status: http.StatusAccepted},
			whenErr:      ErrBadRequest,
			expectStatus: http.StatusBadRequest,
			expectResp:   true,
		},
		{
			name:         "error overrides suggested status with 500 when StatusCode(err)==0",
			whenResp:     &Response{Status: http.StatusAccepted},
			whenErr:      ErrInternalServerError,
			expectStatus: http.StatusInternalServerError,
			expectResp:   true,
		},
		{
			name:         "nil resp, error -> 500 when StatusCode(err)==0",
			whenResp:     nil,
			whenErr:      someErr,
			expectStatus: http.StatusInternalServerError,
			expectResp:   false,
		},
		{
			name:         "committed response wins over error",
			whenResp:     &Response{Committed: true, Status: http.StatusNoContent},
			whenErr:      someErr,
			expectStatus: http.StatusNoContent,
			expectResp:   true,
		},
		{
			name:         "committed response with status 0 falls back to 200 (defensive)",
			whenResp:     &Response{Committed: true, Status: 0},
			whenErr:      someErr,
			expectStatus: http.StatusOK,
			expectResp:   true,
		},
		{
			name:         "resp with status 0 and no error -> 200",
			whenResp:     &Response{Status: 0},
			whenErr:      nil,
			expectStatus: http.StatusOK,
			expectResp:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, status := ResolveResponseStatus(tc.whenResp, tc.whenErr)

			assert.Equal(t, tc.expectResp, resp != nil)
			assert.Equal(t, tc.expectStatus, status)
		})
	}
}

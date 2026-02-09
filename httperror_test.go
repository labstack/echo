// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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

func TestHTTPError_Is(t *testing.T) {
	var testCases = []struct {
		name   string
		err    *HTTPError
		target error
		expect bool
	}{
		{
			name:   "ok, same instance",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: &HTTPError{Code: http.StatusNotFound},
			expect: true,
		},
		{
			name:   "ok, different instance, same code",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: &HTTPError{Code: http.StatusNotFound, Message: "different"},
			expect: true,
		},
		{
			name:   "ok, target is sentinel error",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: ErrNotFound,
			expect: true,
		},
		{
			name:   "nok, different code",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: &HTTPError{Code: http.StatusInternalServerError},
			expect: false,
		},
		{
			name:   "nok, target is sentinel error with different code",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: ErrInternalServerError,
			expect: false,
		},
		{
			name:   "nok, target is different error type",
			err:    &HTTPError{Code: http.StatusNotFound},
			target: errors.New("some error"),
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, errors.Is(tc.err, tc.target))
		})
	}
}

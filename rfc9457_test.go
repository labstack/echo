// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customProblemErrorer struct {
	pe *ProblemError
}

func (ce *customProblemErrorer) Error() string {
	return "custom problem errorer"
}

func (ce *customProblemErrorer) ProblemError() *ProblemError {
	return ce.pe
}

func TestProblemDetailsHTTPErrorHandler(t *testing.T) {
	var testCases = []struct {
		whenError        error
		name             string
		whenMethod       string
		expectBody       string
		expectStatus     int
		givenExposeError bool
	}{
		{
			name:             "ok, expose error = true, HTTPError, no wrapped err",
			givenExposeError: true,
			whenError:        &HTTPError{Code: http.StatusTeapot, Message: "my_error"},
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + wrapped error",
			givenExposeError: true,
			whenError:        HTTPError{Code: http.StatusTeapot, Message: "my_error"}.Wrap(errors.New("internal_error")),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"my_error: internal_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError + wrapped HTTPError",
			givenExposeError: true,
			whenError:        HTTPError{Code: http.StatusTeapot, Message: "my_error"}.Wrap(&HTTPError{Code: http.StatusTeapot, Message: "early_error"}),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"my_error: code=418, message=early_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError + wrapped error",
			whenError:    HTTPError{Code: http.StatusTeapot, Message: "my_error"}.Wrap(errors.New("internal_error")),
			expectStatus: http.StatusTeapot,
			expectBody:   `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"my_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError",
			whenError:    &HTTPError{Code: http.StatusTeapot, Message: "my_error"},
			expectStatus: http.StatusTeapot,
			expectBody:   `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"my_error"}` + "\n",
		},
		{
			name:             "ok, expose error = true, HTTPError, no message, wrapped error",
			givenExposeError: true,
			whenError:        HTTPError{Code: http.StatusTeapot, Message: ""}.Wrap(errors.New("internal_error")),
			expectStatus:     http.StatusTeapot,
			expectBody:       `{"type":"about:blank","title":"I'm a teapot","status":418,"detail":"internal_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, HTTPError, no message",
			whenError:    &HTTPError{Code: http.StatusTeapot, Message: ""},
			expectStatus: http.StatusTeapot,
			expectBody:   `{"type":"about:blank","title":"I'm a teapot","status":418}` + "\n",
		},
		{
			name:             "ok, expose error = true, plain error",
			givenExposeError: true,
			whenError:        fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus:     http.StatusInternalServerError,
			expectBody:       `{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"my errors wraps: internal_error"}` + "\n",
		},
		{
			name:         "ok, expose error = false, plain error",
			whenError:    fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus: http.StatusInternalServerError,
			expectBody:   `{"type":"about:blank","title":"Internal Server Error","status":500}` + "\n",
		},
		{
			name:             "ok, http.HEAD, expose error = true, plain error",
			givenExposeError: true,
			whenMethod:       http.MethodHead,
			whenError:        fmt.Errorf("my errors wraps: %w", errors.New("internal_error")),
			expectStatus:     http.StatusInternalServerError,
			expectBody:       ``,
		},
		{
			name:         "ok, error is *ProblemError, used as-is",
			whenMethod:   http.MethodGet,
			whenError:    &ProblemError{Type: "https://example.com/probs/out-of-credit", Title: "You do not have enough credit.", Status: http.StatusForbidden, Detail: "Your current balance is 30, but that costs 50.", Instance: "/account/12345/msgs/abc"},
			expectStatus: http.StatusForbidden,
			expectBody:   `{"type":"https://example.com/probs/out-of-credit","title":"You do not have enough credit.","status":403,"detail":"Your current balance is 30, but that costs 50.","instance":"/account/12345/msgs/abc"}` + "\n",
		},
		{
			name:         "ok, error is *ProblemError with zero fields, defaults are filled",
			whenMethod:   http.MethodGet,
			whenError:    &ProblemError{},
			expectStatus: http.StatusInternalServerError,
			expectBody:   `{"type":"about:blank","title":"Internal Server Error","status":500}` + "\n",
		},
		{
			name:         "ok, custom error implements ProblemErrorer",
			whenMethod:   http.MethodGet,
			whenError:    &customProblemErrorer{pe: &ProblemError{Status: http.StatusConflict, Detail: "already exists"}},
			expectStatus: http.StatusConflict,
			expectBody:   `{"type":"about:blank","title":"Conflict","status":409,"detail":"already exists"}` + "\n",
		},
		{
			name:         "ok, custom error implements ProblemErrorer, returns nil",
			whenMethod:   http.MethodGet,
			whenError:    &customProblemErrorer{pe: nil},
			expectStatus: http.StatusInternalServerError,
			expectBody:   `{"type":"about:blank","title":"Internal Server Error","status":500}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Logger = slog.New(slog.DiscardHandler)
			e.Any("/path", func(c *Context) error {
				return tc.whenError
			})

			e.HTTPErrorHandler = ProblemDetailsHTTPErrorHandler(tc.givenExposeError)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			req := httptest.NewRequest(method, "/path", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.Equal(t, tc.expectBody, rec.Body.String())
			assert.Equal(t, MIMEApplicationProblemJSON, rec.Header().Get(HeaderContentType))
		})
	}
}

func TestProblemDetailsHTTPErrorHandler_CommittedResponse(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)

	c.orgResponse.Committed = true
	errHandler := ProblemDetailsHTTPErrorHandler(false)

	errHandler(c, errors.New("my_error"))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", resp.Header().Get(HeaderContentType))
}

func TestProblemError_Error(t *testing.T) {
	pe := &ProblemError{Status: http.StatusTeapot, Title: "I'm a teapot"}
	assert.Equal(t, "code=418, message=I'm a teapot", pe.Error())

	pe.Detail = "brewing"
	assert.Equal(t, "code=418, message=I'm a teapot: brewing", pe.Error())
}

func TestProblemError_StatusCode(t *testing.T) {
	pe := &ProblemError{Status: http.StatusTeapot}
	assert.Equal(t, http.StatusTeapot, pe.StatusCode())
}

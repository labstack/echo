// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"fmt"
	"net/http"
)

// ProblemError represents a "problem detail" as defined in RFC 9457 (Problem Details for
// HTTP APIs). https://www.rfc-editor.org/rfc/rfc9457
type ProblemError struct {
	// Type is a URI reference that identifies the problem type. Defaults to "about:blank"
	// when empty, which means the problem is the HTTP status code itself.
	Type string `json:"type"`
	// Title is a short, human-readable summary of the problem type. Defaults to the status
	// text of Status when empty.
	Title string `json:"title"`
	// Status is the HTTP status code for this occurrence of the problem.
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence of the problem.
	Instance string `json:"instance,omitempty"`
}

// Error makes ProblemError compatible with the `error` interface.
func (pe *ProblemError) Error() string {
	msg := pe.Title
	if pe.Detail != "" {
		msg = fmt.Sprintf("%v: %v", pe.Title, pe.Detail)
	}
	return fmt.Sprintf("code=%d, message=%v", pe.Status, msg)
}

// StatusCode returns status code for HTTP response, implementing HTTPStatusCoder interface.
func (pe *ProblemError) StatusCode() int {
	return pe.Status
}

// ProblemErrorer is the interface that custom error types can implement so they can be
// converted into a *ProblemError by ProblemDetailsHTTPErrorHandler.
type ProblemErrorer interface {
	ProblemError() *ProblemError
}

// ProblemDetailsHTTPErrorHandler creates a new HTTP error handler that converts every error
// into a RFC 9457 (Problem Details for HTTP APIs) response and sends it with
// `application/problem+json` content type. `exposeError` parameter decides if the returned
// problem detail will contain the underlying error message for errors that are not *HTTPError
// or do not implement ProblemErrorer.
//
// Precedence used to build the response for a given error:
//  1. If err (or any error wrapped by it) is a *ProblemError, it is used as-is.
//  2. Else if err (or any error wrapped by it) implements ProblemErrorer, ProblemError() is used.
//  3. Else a *ProblemError is built from err's HTTPStatusCoder status code (defaulting to 500).
//     For *HTTPError, Detail is set to its Message, further extended with the wrapped error's
//     message when exposeError is true. For any other error, Detail is only populated (with
//     err.Error()) when exposeError is true.
//
// Any zero-value Type, Title or Status field is defaulted to "about:blank", the status text
// of Status, and 500 respectively.
//
// Note: ProblemDetailsHTTPErrorHandler does not log errors. Use middleware for it if errors
// need to be logged (separately).
func ProblemDetailsHTTPErrorHandler(exposeError bool) HTTPErrorHandler {
	return func(c *Context, err error) {
		if r, _ := UnwrapResponse(c.response); r != nil && r.Committed {
			return
		}

		var pe *ProblemError
		var pder ProblemErrorer
		switch {
		case errors.As(err, &pe):
		case errors.As(err, &pder):
			if pe = pder.ProblemError(); pe == nil {
				pe = &ProblemError{}
			}
		default:
			pe = &ProblemError{}

			var sc HTTPStatusCoder
			if errors.As(err, &sc) {
				pe.Status = sc.StatusCode()
			}

			var he *HTTPError
			if errors.As(err, &he) {
				pe.Detail = he.Message
				if exposeError {
					if wrapped := he.Unwrap(); wrapped != nil {
						if pe.Detail == "" {
							pe.Detail = wrapped.Error()
						} else {
							pe.Detail = fmt.Sprintf("%s: %s", pe.Detail, wrapped.Error())
						}
					}
				}
			} else if exposeError {
				pe.Detail = err.Error()
			}
		}

		if pe.Status == 0 {
			pe.Status = http.StatusInternalServerError
		}
		if pe.Type == "" {
			pe.Type = "about:blank"
		}
		if pe.Title == "" {
			pe.Title = http.StatusText(pe.Status)
		}

		c.Response().Header().Set(HeaderContentType, MIMEApplicationProblemJSON)

		var cErr error
		if c.Request().Method == http.MethodHead { // Issue #608
			cErr = c.NoContent(pe.Status)
		} else {
			cErr = c.JSON(pe.Status, pe)
		}
		if cErr != nil {
			c.Logger().Error("echo RFC 9457 error handler failed to send error to client", "error", cErr) // truly rare case. ala client already disconnected
		}
	}
}

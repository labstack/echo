// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPStatusCoder is interface that errors can implement to produce status code for HTTP response
type HTTPStatusCoder interface {
	StatusCode() int
}

// Following errors can produce HTTP status code by implementing HTTPStatusCoder interface
var (
	ErrBadRequest                  = &httpError{http.StatusBadRequest}            // 400
	ErrUnauthorized                = &httpError{http.StatusUnauthorized}          // 401
	ErrForbidden                   = &httpError{http.StatusForbidden}             // 403
	ErrNotFound                    = &httpError{http.StatusNotFound}              // 404
	ErrMethodNotAllowed            = &httpError{http.StatusMethodNotAllowed}      // 405
	ErrRequestTimeout              = &httpError{http.StatusRequestTimeout}        // 408
	ErrStatusRequestEntityTooLarge = &httpError{http.StatusRequestEntityTooLarge} // 413
	ErrUnsupportedMediaType        = &httpError{http.StatusUnsupportedMediaType}  // 415
	ErrTooManyRequests             = &httpError{http.StatusTooManyRequests}       // 429
	ErrInternalServerError         = &httpError{http.StatusInternalServerError}   // 500
	ErrBadGateway                  = &httpError{http.StatusBadGateway}            // 502
	ErrServiceUnavailable          = &httpError{http.StatusServiceUnavailable}    // 503
)

// Following errors fall into 500 (InternalServerError) category
var (
	ErrValidatorNotRegistered = errors.New("validator not registered")
	ErrRendererNotRegistered  = errors.New("renderer not registered")
	ErrInvalidRedirectCode    = errors.New("invalid redirect status code")
	ErrCookieNotFound         = errors.New("cookie not found")
	ErrInvalidCertOrKeyType   = errors.New("invalid cert or key type, must be string or []byte")
	ErrInvalidListenerNetwork = errors.New("invalid listener network")
)

// NewHTTPError creates new instance of HTTPError
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

// HTTPError represents an error that occurred while handling a request.
type HTTPError struct {
	// Code is status code for HTTP response
	Code    int    `json:"-"`
	Message string `json:"message"`
	err     error
}

// StatusCode returns status code for HTTP response
func (he *HTTPError) StatusCode() int {
	return he.Code
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	msg := he.Message
	if msg == "" {
		msg = http.StatusText(he.Code)
	}
	if he.err == nil {
		return fmt.Sprintf("code=%d, message=%v", he.Code, msg)
	}
	return fmt.Sprintf("code=%d, message=%v, err=%v", he.Code, msg, he.err.Error())
}

// Wrap eturns new HTTPError with given errors wrapped inside
func (he HTTPError) Wrap(err error) error {
	return &HTTPError{
		Code:    he.Code,
		Message: he.Message,
		err:     err,
	}
}

func (he *HTTPError) Unwrap() error {
	return he.err
}

type httpError struct {
	code int
}

func (he httpError) StatusCode() int {
	return he.code
}

func (he httpError) Error() string {
	return http.StatusText(he.code) // does not include status code
}

func (he httpError) Wrap(err error) error {
	return &HTTPError{
		Code:    he.code,
		Message: http.StatusText(he.code),
		err:     err,
	}
}

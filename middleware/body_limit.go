// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"io"

	"github.com/labstack/echo/v5"
)

// BodyLimitConfig defines the config for BodyLimitWithConfig middleware.
type BodyLimitConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// LimitBytes is maximum allowed size in bytes for a request body
	LimitBytes int64
}

// limitedReader returns Echo's status-coded 413 error. Unlike
// http.MaxBytesReader, it does not tell net/http to close the connection
// after an over-limit read.
type limitedReader struct {
	BodyLimitConfig
	reader io.ReadCloser
	read   int64
	err    error
}

// BodyLimit returns a BodyLimit middleware.
//
// BodyLimit middleware sets the maximum allowed size for a request body, if the size exceeds the configured limit, it
// sends "413 - Request Entity Too Large" response. The BodyLimit is determined based on both `Content-Length` request
// header and actual content read, which makes it super secure.
func BodyLimit(limitBytes int64) echo.MiddlewareFunc {
	return BodyLimitWithConfig(BodyLimitConfig{LimitBytes: limitBytes})
}

// BodyLimitWithConfig returns a BodyLimitWithConfig middleware. Middleware sets the maximum allowed size in bytes for
// a request body, if the  size exceeds the configured limit, it sends "413 - Request Entity Too Large" response.
// The BodyLimitWithConfig is determined based on both `Content-Length` request header and actual content read, which
// makes it super secure.
func BodyLimitWithConfig(config BodyLimitConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts BodyLimitConfig to middleware or returns an error for invalid configuration
func (config BodyLimitConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.LimitBytes < 0 {
		return nil, errors.New("body limit must be non-negative")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			req := c.Request()

			// Based on content length
			if req.ContentLength > config.LimitBytes {
				return echo.ErrStatusRequestEntityTooLarge
			}

			// Keep the wrapper attached to the request for its entire lifetime.
			// Outer middleware may still use req.Body after next returns.
			req.Body = &limitedReader{BodyLimitConfig: config, reader: req.Body}

			return next(c)
		}
	}, nil
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(b) == 0 {
		return 0, nil
	}
	remaining := r.LimitBytes - r.read
	// If the caller asked for more bytes than are still allowed, cap the
	// buffer one byte past the limit. That single extra byte is enough to
	// tell whether the underlying reader holds more data than allowed,
	// without ever reading more of it than necessary.
	if int64(len(b))-1 > remaining {
		b = b[:remaining+1]
	}
	n, err = r.reader.Read(b)

	if int64(n) <= remaining {
		r.read += int64(n)
		return n, err
	}

	// The underlying reader offered more data than the limit allows. Only
	// hand out the allowed portion and make the error sticky, so callers
	// that process the n>0 bytes before handling the error (as io.Reader
	// documents) cannot read any further data on subsequent calls.
	n = int(remaining)
	r.read = r.LimitBytes
	r.err = echo.ErrStatusRequestEntityTooLarge
	return n, r.err
}

func (r *limitedReader) Close() error {
	return r.reader.Close()
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
)

// BodyLimitConfig defines the config for BodyLimitWithConfig middleware.
type BodyLimitConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// LimitBytes is maximum allowed size in bytes for a request body
	LimitBytes int64
}

type limitedReader struct {
	BodyLimitConfig
	reader io.ReadCloser
	read   int64
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
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	pool := sync.Pool{
		New: func() any {
			return &limitedReader{BodyLimitConfig: config}
		},
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

			// Based on content read
			r, ok := pool.Get().(*limitedReader)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid pool object")
			}
			r.Reset(req.Body)
			defer pool.Put(r)
			req.Body = r

			return next(c)
		}
	}, nil
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	// Once the limit is known to be exceeded, stay refused. io.Reader's
	// contract invites callers to process the n>0 bytes of a failed read and
	// carry on, so a reader that keeps serving data after the first refusal
	// hands out the whole body to anyone following that advice.
	if r.read > r.LimitBytes {
		return 0, echo.ErrStatusRequestEntityTooLarge
	}

	// Never read further than one byte past the limit: that byte is what
	// proves the body is too large, and anything beyond it is data the caller
	// asked for but is not allowed to have.
	if max := r.LimitBytes - r.read + 1; int64(len(b)) > max {
		b = b[:max]
	}

	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.LimitBytes {
		// Hand back only what fits. The byte past the limit was read to prove
		// the body is too large, not to be delivered, and a caller that
		// processes n>0 before the error must not receive it.
		if over := int(r.read - r.LimitBytes); over <= n {
			n -= over
		}
		return n, echo.ErrStatusRequestEntityTooLarge
	}
	return
}

func (r *limitedReader) Close() error {
	return r.reader.Close()
}

func (r *limitedReader) Reset(reader io.ReadCloser) {
	r.reader = reader
	r.read = 0
}

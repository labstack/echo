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
	// Once the limit has been exceeded, make it sticky — always return 0 bytes and the error.
	// This prevents callers (like encoding/json.Decoder, which processes n>0 before checking err)
	// from reading arbitrarily far past the limit by repeatedly calling Read.
	if r.err != nil {
		return 0, r.err
	}

	// Bound the read size so a single Read call can never return enough bytes to both
	// cross the limit and complete a value (e.g. a JSON object).
	// We allow one byte past the limit so the caller gets the error alongside the data
	// that caused the overflow, which matches standard io.Reader semantics.
	remaining := r.LimitBytes - r.read
	if remaining < int64(len(b))-1 {
		b = b[:remaining+1]
	}

	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.LimitBytes {
		r.err = echo.ErrStatusRequestEntityTooLarge
		return n, r.err
	}
	return
}

func (r *limitedReader) Close() error {
	return r.reader.Close()
}

func (r *limitedReader) Reset(reader io.ReadCloser) {
	r.reader = reader
	r.read = 0
	r.err = nil
}

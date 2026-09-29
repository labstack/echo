// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/bytes"
)

// BodyLimitConfig defines the config for BodyLimit middleware.
type BodyLimitConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Maximum allowed size for a request body, it can be specified
	// as `4x` or `4xB`, where x is one of the multiple from K, M, G, T or P.
	Limit string `yaml:"limit"`
	limit int64
}

type limitedReader struct {
	BodyLimitConfig
	reader io.ReadCloser
	read   int64
	err    error
}

// DefaultBodyLimitConfig is the default BodyLimit middleware config.
var DefaultBodyLimitConfig = BodyLimitConfig{
	Skipper: DefaultSkipper,
}

// BodyLimit returns a BodyLimit middleware.
//
// BodyLimit middleware sets the maximum allowed size for a request body, if the
// size exceeds the configured limit, it sends "413 - Request Entity Too Large"
// response. The BodyLimit is determined based on both `Content-Length` request
// header and actual content read, which makes it super secure.
// Limit can be specified as `4x` or `4xB`, where x is one of the multiple from K, M,
// G, T or P.
func BodyLimit(limit string) echo.MiddlewareFunc {
	c := DefaultBodyLimitConfig
	c.Limit = limit
	return BodyLimitWithConfig(c)
}

// BodyLimitWithConfig returns a BodyLimit middleware with config.
// See: `BodyLimit()`.
func BodyLimitWithConfig(config BodyLimitConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultBodyLimitConfig.Skipper
	}

	limit, err := bytes.Parse(config.Limit)
	if err != nil || limit < 0 {
		panic(fmt.Errorf("echo: invalid body-limit=%s", config.Limit))
	}
	config.limit = limit

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()

			// Based on content length
			if req.ContentLength > config.limit {
				return echo.ErrStatusRequestEntityTooLarge
			}

			// Based on content read
			req.Body = &limitedReader{BodyLimitConfig: config, reader: req.Body}

			return next(c)
		}
	}
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	// A zero limit has historically disabled the read limit in v4.
	if r.limit == 0 {
		return r.reader.Read(b)
	}
	if r.err != nil {
		return 0, r.err
	}
	if len(b) == 0 {
		return 0, nil
	}

	remaining := r.limit - r.read
	// Read at most one byte beyond the limit to distinguish an exact-size body
	// from an oversized one without exposing the extra byte to the caller.
	// Unlike http.MaxBytesReader, this reader does not signal net/http to close
	// the connection after the limit is exceeded.
	if int64(len(b)) > remaining {
		b = b[:remaining+1]
	}
	n, err = r.reader.Read(b)
	if int64(n) > remaining {
		r.read = r.limit
		r.err = echo.ErrStatusRequestEntityTooLarge
		return int(remaining), r.err
	}
	r.read += int64(n)
	return n, err
}

func (r *limitedReader) Close() error {
	return r.reader.Close()
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
)

// DecompressConfig defines the config for Decompress middleware.
type DecompressConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// GzipDecompressPool defines an interface to provide the sync.Pool used to create/store Gzip readers
	GzipDecompressPool Decompressor

	// MaxDecompressedSize limits the maximum size of decompressed request body in bytes.
	// If the decompressed body exceeds this limit, the middleware returns HTTP 413 error.
	// This prevents zip bomb attacks where small compressed payloads decompress to huge sizes.
	// Default: 100 * MB (104,857,600 bytes)
	// Set to -1 to disable limits (not recommended in production).
	MaxDecompressedSize int64
}

// GZIPEncoding content-encoding header if set to "gzip", decompress body contents.
const GZIPEncoding string = "gzip"

// Decompressor is used to get the sync.Pool used by the middleware to get Gzip readers
type Decompressor interface {
	gzipDecompressPool() sync.Pool
}

// DefaultGzipDecompressPool is the default implementation of Decompressor interface
type DefaultGzipDecompressPool struct {
}

func (d *DefaultGzipDecompressPool) gzipDecompressPool() sync.Pool {
	return sync.Pool{New: func() any { return new(gzip.Reader) }}
}

// Decompress decompresses request body based if content encoding type is set to "gzip" with default config
//
// SECURITY: By default, this limits decompressed data to 100MB to prevent zip bomb attacks.
// To customize the limit, use DecompressWithConfig. To disable limits (not recommended in production),
// set MaxDecompressedSize to -1.
func Decompress() echo.MiddlewareFunc {
	return DecompressWithConfig(DecompressConfig{})
}

// DecompressWithConfig returns a decompress middleware with config or panics on invalid configuration.
//
// SECURITY: If MaxDecompressedSize is not set (zero value), it defaults to 100MB to prevent
// DoS attacks via zip bombs. Set to -1 to explicitly disable limits if needed for your use case.
func DecompressWithConfig(config DecompressConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts DecompressConfig to middleware or returns an error for invalid configuration
func (config DecompressConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.GzipDecompressPool == nil {
		config.GzipDecompressPool = &DefaultGzipDecompressPool{}
	}
	// Apply secure default for decompression limit
	if config.MaxDecompressedSize == 0 {
		config.MaxDecompressedSize = 100 * MB
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		pool := config.GzipDecompressPool.gzipDecompressPool()

		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if c.Request().Header.Get(echo.HeaderContentEncoding) != GZIPEncoding {
				return next(c)
			}

			i := pool.Get()
			gr, ok := i.(*gzip.Reader)
			if !ok || gr == nil {
				if err, isErr := i.(error); isErr {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
				}
				return echo.NewHTTPError(http.StatusInternalServerError, "unexpected type from gzip decompression pool")
			}
			defer pool.Put(gr)

			b := c.Request().Body
			defer b.Close()

			if err := gr.Reset(b); err != nil {
				if err == io.EOF { //ignore if body is empty
					return next(c)
				}
				return err
			}

			// only Close gzip reader if it was set to a proper gzip source otherwise it will panic on close.
			defer gr.Close()

			// Apply decompression size limit to prevent zip bombs
			if config.MaxDecompressedSize > 0 {
				c.Request().Body = &limitedGzipReader{
					Reader:    gr,
					remaining: config.MaxDecompressedSize,
					limit:     config.MaxDecompressedSize,
				}
			} else {
				// -1 means explicitly unlimited (not recommended)
				c.Request().Body = gr
			}

			return next(c)
		}
	}, nil
}

// limitedGzipReader wraps a gzip reader with size limiting to prevent zip bombs
type limitedGzipReader struct {
	*gzip.Reader
	remaining int64
	limit     int64
}

func (r *limitedGzipReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		// Limit exceeded - return 413 error
		return 0, echo.ErrStatusRequestEntityTooLarge
	}

	// Limit the read to remaining bytes
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}

	n, err = r.Reader.Read(p)
	r.remaining -= int64(n)

	return n, err
}

func (r *limitedGzipReader) Close() error {
	return r.Reader.Close()
}

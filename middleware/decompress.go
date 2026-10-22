// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// DecompressConfig defines the config for Decompress middleware.
type DecompressConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// GzipDecompressPool defines an interface to provide the sync.Pool used to create/store Gzip readers
	GzipDecompressPool Decompressor

	// MaxDecompressedSize limits the decompressed body in bytes. Reading beyond
	// the limit returns HTTP 413. Defaults to 100 MiB; -1 disables the limit.
	MaxDecompressedSize int64
}

// GZIPEncoding content-encoding header if set to "gzip", decompress body contents.
const GZIPEncoding string = "gzip"

// Decompressor is used to get the sync.Pool used by the middleware to get Gzip readers
type Decompressor interface {
	gzipDecompressPool() sync.Pool
}

// DefaultDecompressConfig defines the config for decompress middleware
var DefaultDecompressConfig = DecompressConfig{
	Skipper:            DefaultSkipper,
	GzipDecompressPool: &DefaultGzipDecompressPool{},
}

// DefaultGzipDecompressPool is the default implementation of Decompressor interface
type DefaultGzipDecompressPool struct {
}

func (d *DefaultGzipDecompressPool) gzipDecompressPool() sync.Pool {
	return sync.Pool{New: func() interface{} { return new(gzip.Reader) }}
}

// Decompress decompresses request body based if content encoding type is set to "gzip" with default config
func Decompress() echo.MiddlewareFunc {
	return DecompressWithConfig(DefaultDecompressConfig)
}

// DecompressWithConfig decompresses request body based if content encoding type is set to "gzip" with config
func DecompressWithConfig(config DecompressConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultGzipConfig.Skipper
	}
	if config.GzipDecompressPool == nil {
		config.GzipDecompressPool = DefaultDecompressConfig.GzipDecompressPool
	}
	if config.MaxDecompressedSize == 0 {
		config.MaxDecompressedSize = 100 * 1024 * 1024
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		pool := config.GzipDecompressPool.gzipDecompressPool()

		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if c.Request().Header.Get(echo.HeaderContentEncoding) != GZIPEncoding {
				return next(c)
			}

			i := pool.Get()
			gr, ok := i.(*gzip.Reader)
			if !ok || gr == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
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

			if config.MaxDecompressedSize > 0 {
				c.Request().Body = &limitedGzipReader{Reader: gr, remaining: config.MaxDecompressedSize}
			} else {
				c.Request().Body = gr
			}
			c.Request().ContentLength = -1

			return next(c)
		}
	}
}

type limitedGzipReader struct {
	*gzip.Reader
	remaining int64
}

func (r *limitedGzipReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.remaining < 0 {
		return 0, echo.ErrStatusRequestEntityTooLarge
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.Reader.Read(probe[:])
		if n > 0 {
			r.remaining = -1
			return 0, echo.ErrStatusRequestEntityTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.Reader.Read(p)
	r.remaining -= int64(n)
	return n, err
}

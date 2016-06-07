package middleware

import (
	"fmt"
	"io"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/bytes"
)

type (
	// BodyLimitConfig defines the config for body limit middleware.
	BodyLimitConfig struct {
		// Maximum allowed size for a request body, it can be specified
		// as `4x` or `4xB`, where x is one of the multiple from K, M, G, T or P.
		Limit string `json:"limit"`
		limit int64
	}

	limitedReader struct {
		BodyLimitConfig
		reader  io.Reader
		read    int64
		context echo.Context
	}
)

// BodyLimit returns a body limit middleware.
//
// BodyLimit middleware sets the maximum allowed size for a request body, if the
// size exceeds the configured limit, it sends "413 - Request Entity Too Large"
// response. The body limit is determined based on both `Content-Length` request
// header and actual content read, which makes it super secure.
// Limit can be specified as `4x` or `4xB`, where x is one of the multiple from K, M,
// G, T or P.
func BodyLimit(limit string) echo.MiddlewareFunc {
	return BodyLimitWithConfig(BodyLimitConfig{Limit: limit})
}

// BodyLimitWithConfig returns a body limit middleware from config.
// See: `BodyLimit()`.
func BodyLimitWithConfig(config BodyLimitConfig) echo.MiddlewareFunc {
	limit, err := bytes.Parse(config.Limit)
	if err != nil {
		panic(fmt.Errorf("invalid body-limit=%s", config.Limit))
	}
	config.limit = int64(limit)
	pool := limitedReaderPool(config)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Based on content length
			if req.ContentLength() > config.limit {
				return echo.ErrStatusRequestEntityTooLarge
			}

			// Based on content read
			r := pool.Get().(*limitedReader)
			r.Reset(req.Body(), c)
			defer pool.Put(r)
			req.SetBody(r)

			return next(c)
		}
	}
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.limit {
		return n, echo.ErrStatusRequestEntityTooLarge
	}
	return
}

func (r *limitedReader) Reset(reader io.Reader, context echo.Context) {
	r.reader = reader
	r.context = context
}

func limitedReaderPool(c BodyLimitConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return &limitedReader{BodyLimitConfig: c}
		},
	}
}

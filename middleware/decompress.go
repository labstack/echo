package middleware

import (
	"compress/gzip"
	"errors"
	"github.com/andybalholm/brotli"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

type (
	// DecompressConfig defines the config for Decompress middleware.
	DecompressConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
	}
)

// GZIPEncoding content-encoding header if set to "gzip", decompress body contents.
const GZIPEncoding string = "gzip"

// BrotliEncoding content-encoding header if set to "br", decompress body contents.
const BrotliEncoding string = "br"

var (
	//DefaultDecompressConfig defines the config for decompress middleware
	DefaultDecompressConfig = DecompressConfig{
		Skipper: DefaultSkipper,
	}

	unsupportedDecompressEncodingErr = errors.New("unsupported content encoding")
)

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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			encode := c.Request().Header.Get(echo.HeaderContentEncoding)
			if encode != GZIPEncoding && encode != BrotliEncoding {
				return next(c)
			}

			b := c.Request().Body
			defer b.Close()

			cr, err := acquireCompressReader(encode, b)
			if err == io.EOF { //ignore if body is empty
				return next(c)
			}
			if err != nil {
				return err
			}
			if cr == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, unsupportedDecompressEncodingErr)
			}
			defer releaseCompressReader(encode, cr)

			c.Request().Body = cr

			return next(c)
		}
	}
}

type compressReader interface {
	io.ReadCloser
	Reset(io.Reader) error
}

type brotliReaderWrapper struct {
	reader *brotli.Reader
}

func (b *brotliReaderWrapper) Read(p []byte) (n int, err error) {
	return b.reader.Read(p)
}

func (b *brotliReaderWrapper) Close() error {
	// do nothing
	return nil
}

func (b *brotliReaderWrapper) Reset(reader io.Reader) error {
	return b.reader.Reset(reader)
}

var (
	gzipReaderPool   = sync.Pool{New: func() interface{} { return new(gzip.Reader) }}
	brotliReaderPool = sync.Pool{New: func() interface{} { return &brotliReaderWrapper{reader: new(brotli.Reader)} }}
)

func acquireCompressReader(encode string, source io.Reader) (compressReader, error) {
	switch encode {
	case BrotliEncoding:
		v := brotliReaderPool.Get()
		r := v.(*brotliReaderWrapper)
		err := r.Reset(source)
		return r, err
	case GZIPEncoding:
		v := gzipReaderPool.Get()
		r := v.(*gzip.Reader)
		err := r.Reset(source)
		return r, err
	}
	return nil, unsupportedDecompressEncodingErr
}

func releaseCompressReader(encode string, r io.ReadCloser) {
	// only Close reader if it was set to a proper source otherwise it will panic on close.
	r.Close()
	switch encode {
	case BrotliEncoding:
		brotliWriterPool.Put(r)
	case GZIPEncoding:
		gzipWriterPool.Put(r)
	}
}

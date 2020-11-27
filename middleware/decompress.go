package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
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

//GZIPEncoding content-encoding header if set to "gzip", decompress body contents.
const GZIPEncoding string = "gzip"

var (
	//DefaultDecompressConfig defines the config for decompress middleware
	DefaultDecompressConfig = DecompressConfig{Skipper: DefaultSkipper}
)

//Decompress decompresses request body based if content encoding type is set to "gzip" with default config
func Decompress() echo.MiddlewareFunc {
	return DecompressWithConfig(DefaultDecompressConfig)
}

//DecompressWithConfig decompresses request body based if content encoding type is set to "gzip" with config
func DecompressWithConfig(config DecompressConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		pool := gzipDecompressPool()
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			switch c.Request().Header.Get(echo.HeaderContentEncoding) {
			case GZIPEncoding:
				b := c.Request().Body

				i := pool.Get()
				gr, ok := i.(*gzip.Reader)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, i.(error).Error())
				}

				if err := gr.Reset(b); err != nil {
					pool.Put(gr)
					if err == io.EOF { //ignore if body is empty
						return next(c)
					}
					return err
				}
				var buf bytes.Buffer
				io.Copy(&buf, gr)

				gr.Close()
				pool.Put(gr)

				b.Close() // http.Request.Body is closed by the Server, but because we are replacing it, it must be closed here

				r := ioutil.NopCloser(&buf)
				c.Request().Body = r
			}
			return next(c)
		}
	}
}

func gzipDecompressPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			// create with an empty reader (but with GZIP header)
			w, err := gzip.NewWriterLevel(ioutil.Discard, gzip.BestSpeed)
			if err != nil {
				return err
			}

			b := new(bytes.Buffer)
			w.Reset(b)
			w.Flush()
			w.Close()

			r, err := gzip.NewReader(bytes.NewReader(b.Bytes()))
			if err != nil {
				return err
			}
			return r
		},
	}
}

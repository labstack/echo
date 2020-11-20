package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
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
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			switch c.Request().Header.Get(echo.HeaderContentEncoding) {
			case GZIPEncoding:
				gr, err := gzip.NewReader(c.Request().Body)
				if err != nil {
					if err == io.EOF { //ignore if body is empty
						return next(c)
					}
					return err
				}
				defer gr.Close()
				var buf bytes.Buffer
				io.Copy(&buf, gr)
				r := ioutil.NopCloser(&buf)
				defer r.Close()
				c.Request().Body = r
			}
			return next(c)
		}
	}
}

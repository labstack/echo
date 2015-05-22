package middleware

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/labstack/echo"
	"net/http"
)

type (
	gzipWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	scheme := "gzip"

	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if strings.Contains(c.Request().Header.Get(echo.AcceptEncoding), scheme) {
				w := gzip.NewWriter(c.Response().Writer())
				defer w.Close()
				gw := gzipWriter{Writer: w, ResponseWriter: c.Response().Writer()}
				c.Response().Header().Set(echo.ContentEncoding, scheme)
				c.Response().SetWriter(gw)
			}
			return h(c)
		}
	}
}

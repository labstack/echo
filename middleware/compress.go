package middleware

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/labstack/echo"
)

type (
	gzipWriter struct {
		io.Writer
		*echo.Response
	}
)

func (g gzipWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// Gzip compresses HTTP response using gzip compression scheme.
func Gzip() echo.MiddlewareFunc {
	scheme := "gzip"

	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) *echo.HTTPError {
			if !strings.Contains(c.Request.Header.Get(echo.AcceptEncoding), scheme) {
				return nil
			}

			w := gzip.NewWriter(c.Response.Writer)
			defer w.Close()
			gw := gzipWriter{Writer: w, Response: c.Response}
			c.Response.Header().Set(echo.ContentEncoding, scheme)
			c.Response = &echo.Response{Writer: gw}
			if he := h(c); he != nil {
				c.Error(he)
			}
			return nil
		}
	}
}

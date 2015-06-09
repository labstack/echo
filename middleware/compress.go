package middleware

import (
	"compress/gzip"
	"strings"

	"net/http"

	"bufio"
	"net"

	"github.com/labstack/echo"
	"io"
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

func (w gzipWriter) Flush() {
	w.Writer.(*gzip.Writer).Flush()
}

func (w gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
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
			if err := h(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}

package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo"
)

type (
	gzipWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

func (w gzipWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.ContentType) == "" {
		w.Header().Set(echo.ContentType, http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func (w gzipWriter) Flush() error {
	return w.Writer.(*gzip.Writer).Flush()
}

func (w gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(ioutil.Discard)
	},
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() echo.MiddlewareFunc {
	scheme := "gzip"

	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Add(echo.Vary, echo.AcceptEncoding)
			if strings.Contains(c.Request().Header.Get(echo.AcceptEncoding), scheme) {
				w := writerPool.Get().(*gzip.Writer)
				w.Reset(c.Response().Writer())
				defer func() {
					w.Close()
					writerPool.Put(w)
				}()
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

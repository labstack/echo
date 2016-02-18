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
	"github.com/labstack/echo/engine"
)

type (
	Gzip struct {
		level    int
		priority int
	}

	gzipWriter struct {
		io.Writer
		engine.Response
	}
)

func NewGzip() *Gzip {
	return &Gzip{}
}

func (g *Gzip) SetLevel(l int) {
	g.level = l
}

func (g *Gzip) SetPriority(p int) {
	g.priority = p
}

func (g *Gzip) Priority() int {
	return g.priority
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func (*Gzip) Handle(h echo.Handler) echo.Handler {
	scheme := "gzip"
	return echo.HandlerFunc(func(c echo.Context) error {
		c.Response().Header().Add(echo.Vary, echo.AcceptEncoding)
		if strings.Contains(c.Request().Header().Get(echo.AcceptEncoding), scheme) {
			w := writerPool.Get().(*gzip.Writer)
			w.Reset(c.Response().Writer())
			defer func() {
				w.Close()
				writerPool.Put(w)
			}()
			gw := gzipWriter{Writer: w, Response: c.Response()}
			c.Response().Header().Set(echo.ContentEncoding, scheme)
			c.Response().SetWriter(gw)
		}
		if err := h.Handle(c); err != nil {
			c.Error(err)
		}
		return nil
	})
}

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
	return w.Response.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.Response.(http.CloseNotifier).CloseNotify()
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(ioutil.Discard)
	},
}

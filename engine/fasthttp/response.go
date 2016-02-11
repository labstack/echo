// +build !appengine

package fasthttp

import (
	"io"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/logger"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
)

type (
	Response struct {
		context   *fasthttp.RequestCtx
		header    engine.Header
		status    int
		size      int64
		committed bool
		writer    io.Writer
		logger    logger.Logger
	}
)

func NewResponse(c *fasthttp.RequestCtx) *Response {
	return &Response{
		context: c,
		header:  &ResponseHeader{c.Response.Header},
		writer:  c,
		logger:  log.New("test"),
	}
}

func (r *Response) Object() interface{} {
	return r.context
}

func (r *Response) Header() engine.Header {
	return r.header
}

func (r *Response) WriteHeader(code int) {
	if r.committed {
		r.logger.Warn("response already committed")
		return
	}
	r.status = code
	r.context.SetStatusCode(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (int, error) {
	return r.context.Write(b)
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int64 {
	return r.size
}

func (r *Response) Committed() bool {
	return r.committed
}

func (r *Response) SetWriter(w io.Writer) {
	r.writer = w
}

func (r *Response) Writer() io.Writer {
	return r.writer
}

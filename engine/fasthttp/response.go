package fasthttp

import (
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Response struct {
		context   *fasthttp.RequestCtx
		header    engine.Header
		status    int
		size      int64
		committed bool
	}
)

func (r *Response) Header() engine.Header {
	return r.header
}

func (r *Response) WriteHeader(code int) {
	r.context.SetStatusCode(code)
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

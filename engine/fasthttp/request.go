package fasthttp

import "io"

import (
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Request struct {
		request *fasthttp.RequestCtx
		url     engine.URL
		header  engine.Header
	}
)

func (r *Request) Object() interface{} {
	return r.request
}

func (r *Request) URI() string {
	return string(r.request.RequestURI())
}

func (r *Request) URL() engine.URL {
	return r.url
}

func (r *Request) Header() engine.Header {
	return r.header
}

func (r *Request) RemoteAddress() string {
	return r.request.RemoteAddr().String()
}

func (r *Request) Method() string {
	return string(r.request.Method())
}

func (r *Request) Body() io.ReadCloser {
	// return r.context.PostBody()
	return nil
}

func (r *Request) FormValue(name string) string {
	return ""
}

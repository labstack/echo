package fasthttp

import (
	"bytes"
	"io"
	"io/ioutil"
)

import (
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Request struct {
		context *fasthttp.RequestCtx
		url     engine.URL
		header  engine.Header
	}
)

func NewRequest(c *fasthttp.RequestCtx) *Request {
	return &Request{
		context: c,
		url:     &URL{url: c.URI()},
		header:  &RequestHeader{c.Request.Header},
	}
}

func (r *Request) Host() string {
	return string(r.context.Host())
}

func (r *Request) URI() string {
	return string(r.context.RequestURI())
}

func (r *Request) URL() engine.URL {
	return r.url
}

func (r *Request) Header() engine.Header {
	return r.header
}

func (r *Request) RemoteAddress() string {
	return r.context.RemoteAddr().String()
}

func (r *Request) Method() string {
	return string(r.context.Method())
}

func (r *Request) Body() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(r.context.PostBody()))
}

func (r *Request) FormValue(name string) string {
	return ""
}

func (r *Request) Object() interface{} {
	return r.context
}

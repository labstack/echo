// +build !appengine

package fasthttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
)

import (
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Request struct {
		*fasthttp.RequestCtx
		url    engine.URL
		header engine.Header
	}
)

func NewRequest(c *fasthttp.RequestCtx) *Request {
	return &Request{
		RequestCtx: c,
		url:        &URL{URI: c.URI()},
		header:     &RequestHeader{c.Request.Header},
	}
}

func (r *Request) TLS() bool {
	return r.IsTLS()
}

func (r *Request) Scheme() string {
	return string(r.RequestCtx.URI().Scheme())
}

func (r *Request) Host() string {
	return string(r.RequestCtx.Host())
}

func (r *Request) URI() string {
	return string(r.RequestURI())
}

func (r *Request) URL() engine.URL {
	return r.url
}

func (r *Request) Header() engine.Header {
	return r.header
}

func (r *Request) RemoteAddress() string {
	return r.RemoteAddr().String()
}

func (r *Request) Method() string {
	return string(r.RequestCtx.Method())
}

func (r *Request) Body() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(r.PostBody()))
}

func (r *Request) FormValue(name string) string {
	return string(r.RequestCtx.FormValue(name))
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.RequestCtx.FormFile(name)
}

func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.RequestCtx.MultipartForm()
}

func (r *Request) reset(c *fasthttp.RequestCtx, h engine.Header, u engine.URL) {
	r.RequestCtx = c
	r.header = h
	r.url = u
}

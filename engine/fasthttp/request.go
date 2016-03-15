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
	// Request implements `engine.Request`.
	Request struct {
		*fasthttp.RequestCtx
		url    engine.URL
		header engine.Header
	}
)

// TLS implements `Request#TLS` method.
func (r *Request) TLS() bool {
	return r.IsTLS()
}

// Scheme implements `Request#Scheme` method.
func (r *Request) Scheme() string {
	return string(r.RequestCtx.URI().Scheme())
}

// Host implements `Request#Host` method.
func (r *Request) Host() string {
	return string(r.RequestCtx.Host())
}

// URL implements `Request#URL` method.
func (r *Request) URL() engine.URL {
	return r.url
}

// Header implements `Request#Header` method.
func (r *Request) Header() engine.Header {
	return r.header
}

// UserAgent implements `Request#UserAgent` method.
func (r *Request) UserAgent() string {
	return string(r.RequestCtx.UserAgent())
}

// RemoteAddress implements `Request#RemoteAddress` method.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr().String()
}

// Method implements `Request#Method` method.
func (r *Request) Method() string {
	return string(r.RequestCtx.Method())
}

// URI implements `Request#URI` method.
func (r *Request) URI() string {
	return string(r.RequestURI())
}

// Body implements `Request#Body` method.
func (r *Request) Body() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(r.PostBody()))
}

// FormValue implements `Request#FormValue` method.
func (r *Request) FormValue(name string) string {
	return string(r.RequestCtx.FormValue(name))
}

// FormFile implements `Request#FormFile` method.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.RequestCtx.FormFile(name)
}

// MultipartForm implements `Request#MultipartForm` method.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.RequestCtx.MultipartForm()
}

func (r *Request) reset(c *fasthttp.RequestCtx, h engine.Header, u engine.URL) {
	r.RequestCtx = c
	r.header = h
	r.url = u
}

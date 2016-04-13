// +build !appengine

package fasthttp

import (
	"bytes"
	"io"
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

// IsTLS implements `engine.Request#TLS` function.
func (r *Request) IsTLS() bool {
	return r.IsTLS()
}

// Scheme implements `engine.Request#Scheme` function.
func (r *Request) Scheme() string {
	return string(r.RequestCtx.URI().Scheme())
}

// Host implements `engine.Request#Host` function.
func (r *Request) Host() string {
	return string(r.RequestCtx.Host())
}

// URL implements `engine.Request#URL` function.
func (r *Request) URL() engine.URL {
	return r.url
}

// Header implements `engine.Request#Header` function.
func (r *Request) Header() engine.Header {
	return r.header
}

// ContentLength implements `engine.Request#ContentLength` function.
func (r *Request) ContentLength() int {
	return r.Request.Header.ContentLength()
}

// UserAgent implements `engine.Request#UserAgent` function.
func (r *Request) UserAgent() string {
	return string(r.RequestCtx.UserAgent())
}

// RemoteAddress implements `engine.Request#RemoteAddress` function.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr().String()
}

// Method implements `engine.Request#Method` function.
func (r *Request) Method() string {
	return string(r.RequestCtx.Method())
}

// SetMethod implements `engine.Request#SetMethod` function.
func (r *Request) SetMethod(method string) {
	r.Request.Header.SetMethod(method)
}

// URI implements `engine.Request#URI` function.
func (r *Request) URI() string {
	return string(r.RequestURI())
}

// SetURI implements `engine.Request#SetURI` function.
func (r *Request) SetURI(uri string) {
	r.Request.Header.SetRequestURI(uri)
}

// Body implements `engine.Request#Body` function.
func (r *Request) Body() io.Reader {
	return bytes.NewBuffer(r.PostBody())
}

// FormValue implements `engine.Request#FormValue` function.
func (r *Request) FormValue(name string) string {
	return string(r.RequestCtx.FormValue(name))
}

// FormParams implements `engine.Request#FormParams` function.
func (r *Request) FormParams() (params map[string][]string) {
	params = make(map[string][]string)
	r.PostArgs().VisitAll(func(k, v []byte) {
		// TODO: Filling with only first value
		params[string(k)] = []string{string(v)}
	})
	return
}

// FormFile implements `engine.Request#FormFile` function.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.RequestCtx.FormFile(name)
}

// MultipartForm implements `engine.Request#MultipartForm` function.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.RequestCtx.MultipartForm()
}

func (r *Request) reset(c *fasthttp.RequestCtx, h engine.Header, u engine.URL) {
	r.RequestCtx = c
	r.header = h
	r.url = u
}

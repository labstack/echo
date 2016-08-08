// +build !appengine

package fasthttp

import (
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
	"github.com/valyala/fasthttp"
)

type (
	// Response implements `engine.Response`.
	Response struct {
		*fasthttp.RequestCtx
		header    engine.Header
		status    int
		size      int64
		committed bool
		logger    log.Logger
	}
)

// NewResponse returns `Response` instance.
func NewResponse(c *fasthttp.RequestCtx, l log.Logger) *Response {
	return &Response{
		RequestCtx: c,
		header:     &ResponseHeader{ResponseHeader: &c.Response.Header},
		logger:     l,
	}
}

// Header implements `engine.Response#Header` function.
func (r *Response) Header() engine.Header {
	return r.header
}

// WriteHeader implements `engine.Response#WriteHeader` function.
func (r *Response) WriteHeader(code int) {
	if r.committed {
		r.logger.Warn("response already committed")
		return
	}
	r.status = code
	r.SetStatusCode(code)
	r.committed = true
}

// Write implements `engine.Response#Write` function.
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.committed {
		r.WriteHeader(http.StatusOK)
	}
	r.Response.AppendBody(b)
	r.size += int64(len(b))
	return
}

// WriteString implements `engine.Response#WriteString` function.
func (r *Response) WriteString(s string) (n int, err error) {
	if !r.committed {
		r.WriteHeader(http.StatusOK)
	}
	r.Response.AppendBodyString(s)
	r.size += int64(len(s))
	return
}

// SetCookie implements `engine.Response#SetCookie` function.
func (r *Response) SetCookie(c engine.Cookie) {
	cookie := new(fasthttp.Cookie)
	cookie.SetKey(c.Name())
	cookie.SetValue(c.Value())
	cookie.SetPath(c.Path())
	cookie.SetDomain(c.Domain())
	cookie.SetExpire(c.Expires())
	cookie.SetSecure(c.Secure())
	cookie.SetHTTPOnly(c.HTTPOnly())
	r.Response.Header.SetCookie(cookie)
}

// Status implements `engine.Response#Status` function.
func (r *Response) Status() int {
	return r.status
}

// Size implements `engine.Response#Size` function.
func (r *Response) Size() int64 {
	return r.size
}

// Committed implements `engine.Response#Committed` function.
func (r *Response) Committed() bool {
	return r.committed
}

func (r *Response) reset(c *fasthttp.RequestCtx, h engine.Header) {
	r.RequestCtx = c
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
}

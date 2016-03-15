// +build !appengine

package fasthttp

import (
	"io"
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
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
		writer    io.Writer
		logger    *log.Logger
	}
)

// Header implements `engine.Response#Header` method.
func (r *Response) Header() engine.Header {
	return r.header
}

// WriteHeader implements `engine.Response#WriteHeader` method.
func (r *Response) WriteHeader(code int) {
	if r.committed {
		r.logger.Warn("response already committed")
		return
	}
	r.status = code
	r.SetStatusCode(code)
	r.committed = true
}

// Write implements `engine.Response#Write` method.
func (r *Response) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

// Status implements `engine.Response#Status` method.
func (r *Response) Status() int {
	return r.status
}

// Size implements `engine.Response#Size` method.
func (r *Response) Size() int64 {
	return r.size
}

// Committed implements `engine.Response#Committed` method.
func (r *Response) Committed() bool {
	return r.committed
}

// Writer implements `engine.Response#Writer` method.
func (r *Response) Writer() io.Writer {
	return r.writer
}

// SetWriter implements `engine.Response#SetWriter` method.
func (r *Response) SetWriter(w io.Writer) {
	r.writer = w
}

func (r *Response) reset(c *fasthttp.RequestCtx, h engine.Header) {
	r.RequestCtx = c
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = c
}

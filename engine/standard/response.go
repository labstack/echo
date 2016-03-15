package standard

import (
	"io"
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	// Response implements `engine.Response`.
	Response struct {
		http.ResponseWriter
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
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

// Write implements `engine.Response#Write` method.
func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.writer.Write(b)
	r.size += int64(n)
	return
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

func (r *Response) reset(w http.ResponseWriter, h engine.Header) {
	r.ResponseWriter = w
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = w
}

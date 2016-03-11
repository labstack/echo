package standard

import (
	"io"
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
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

func NewResponse(w http.ResponseWriter, l *log.Logger) *Response {
	return &Response{
		ResponseWriter: w,
		header:         &Header{w.Header()},
		writer:         w,
		logger:         l,
	}
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
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.writer.Write(b)
	r.size += int64(n)
	return
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

func (r *Response) reset(w http.ResponseWriter, h engine.Header) {
	r.ResponseWriter = w
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = w
}

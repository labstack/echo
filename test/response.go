package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	Response struct {
		response  http.ResponseWriter
		header    engine.Header
		status    int
		size      int64
		committed bool
		writer    io.Writer
		logger    *log.Logger
	}

	ResponseRecorder struct {
		engine.Response
		Body *bytes.Buffer
	}
)

func NewResponseRecorder() *ResponseRecorder {
	rec := httptest.NewRecorder()
	return &ResponseRecorder{
		Response: &Response{
			response: rec,
			header:   &Header{rec.Header()},
			writer:   rec,
			logger:   log.New("test"),
		},
		Body: rec.Body,
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
	r.response.WriteHeader(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	if !r.committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.writer.Write(b)
	r.size += int64(n)
	return
}

// SetCookie implements `engine.Response#SetCookie` function.
func (r *Response) SetCookie(c engine.Cookie) {
	http.SetCookie(r.response, &http.Cookie{
		Name:     c.Name(),
		Value:    c.Value(),
		Path:     c.Path(),
		Domain:   c.Domain(),
		Expires:  c.Expires(),
		Secure:   c.Secure(),
		HttpOnly: c.HTTPOnly(),
	})
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
	r.response = w
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = w
}

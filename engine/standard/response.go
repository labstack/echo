package standard

import (
	"bufio"
	"io"
	"net"
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
)

type (
	// Response implements `engine.Response`.
	Response struct {
		http.ResponseWriter
		adapter   *responseAdapter
		header    engine.Header
		status    int
		size      int64
		committed bool
		writer    io.Writer
		logger    log.Logger
	}

	responseAdapter struct {
		*Response
	}
)

// NewResponse returns `Response` instance.
func NewResponse(w http.ResponseWriter, l log.Logger) (r *Response) {
	r = &Response{
		ResponseWriter: w,
		header:         &Header{Header: w.Header()},
		writer:         w,
		logger:         l,
	}
	r.adapter = &responseAdapter{Response: r}
	return
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
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

// Write implements `engine.Response#Write` function.
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
	http.SetCookie(r.ResponseWriter, &http.Cookie{
		Name:     c.Name(),
		Value:    c.Value(),
		Path:     c.Path(),
		Domain:   c.Domain(),
		Expires:  c.Expires(),
		Secure:   c.Secure(),
		HttpOnly: c.HTTPOnly(),
	})
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

// Writer implements `engine.Response#Writer` function.
func (r *Response) Writer() io.Writer {
	return r.writer
}

// SetWriter implements `engine.Response#SetWriter` function.
func (r *Response) SetWriter(w io.Writer) {
	r.writer = w
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See https://golang.org/pkg/net/http/#Flusher
func (r *Response) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See https://golang.org/pkg/net/http/#Hijacker
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
// This mechanism can be used to cancel long operations on the server if the
// client has disconnected before the response is ready.
// See https://golang.org/pkg/net/http/#CloseNotifier
func (r *Response) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *Response) reset(w http.ResponseWriter, a *responseAdapter, h engine.Header) {
	r.ResponseWriter = w
	r.adapter = a
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = w
}

func (r *responseAdapter) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *responseAdapter) reset(res *Response) {
	r.Response = res
}

package standard

import (
	"bufio"
	"io"
	"net"
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

	responseAdapter struct {
		http.ResponseWriter
		writer   io.Writer
		response *Response
		code     int
	}
)

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
	n, err = r.writer.Write(b)
	r.size += int64(n)
	return
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

func (r *Response) reset(w http.ResponseWriter, h engine.Header) {
	r.ResponseWriter = w
	r.header = h
	r.status = http.StatusOK
	r.size = 0
	r.committed = false
	r.writer = w
}

func (r *responseAdapter) Write(b []byte) (n int, err error) {
	return r.writer.Write(b)
}

func (r *responseAdapter) WriteHeader(code int) {
	r.response.WriteHeader(code)
}

func (r *responseAdapter) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *responseAdapter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

func (r *responseAdapter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

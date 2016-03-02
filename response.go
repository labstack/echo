package echo

import (
	"bufio"
	"net"
	"net/http"
)

// Response wraps an http.ResponseWriter and implements its interface to be used
// by an HTTP handler to construct an HTTP response.
// See [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter)
type Response struct {
	writer    http.ResponseWriter
	status    int
	size      int64
	committed bool
	echo      *Echo
}

// NewResponse creates a new instance of Response.
func NewResponse(w http.ResponseWriter, e *Echo) *Response {
	return &Response{writer: w, echo: e}
}

// SetWriter sets the http.ResponseWriter instance for this Response.
func (r *Response) SetWriter(w http.ResponseWriter) {
	r.writer = w
}

// Writer returns the http.ResponseWriter instance for this Response.
func (r *Response) Writer() http.ResponseWriter {
	return r.writer
}

// Header returns the header map for the writer that will be sent by
// WriteHeader. Changing the header after a call to WriteHeader (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to WriteHeader (see example)
// To suppress implicit response headers, set their value to nil.
// Example [ResponseWriter.Trailers](https://golang.org/pkg/net/http/#example_ResponseWriter_trailers)
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly
// used to send error codes.
func (r *Response) WriteHeader(code int) {
	if r.committed {
		r.echo.Logger().Warn("response already committed")
		return
	}
	r.status = code
	r.writer.WriteHeader(code)
	r.committed = true
}

// Write wraps and implements the http.Response.Write specification.
// Additionally, Write will increment the size of the current response.
// See [http.Response.Write](https://golang.org/pkg/net/http/#Response.Write)
func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.writer.Write(b)
	r.size += int64(n)
	return n, err
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.writer.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.writer.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
// This mechanism can be used to cancel long operations on the server if the
// client has disconnected before the response is ready.
// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r *Response) CloseNotify() <-chan bool {
	return r.writer.(http.CloseNotifier).CloseNotify()
}

// Status returns the HTTP status code of the response.
func (r *Response) Status() int {
	return r.status
}

// Size returns the current size, in bytes, of the response.
func (r *Response) Size() int64 {
	return r.size
}

// Committed asserts whether or not the response has been committed to.
func (r *Response) Committed() bool {
	return r.committed
}

func (r *Response) reset(w http.ResponseWriter, e *Echo) {
	r.writer = w
	r.size = 0
	r.status = http.StatusOK
	r.committed = false
	r.echo = e
}

package echo

import (
	"bufio"
	"net"
	"net/http"
)

type (
	// Response wraps an http.ResponseWriter and implements its interface to be used
	// by an HTTP handler to construct an HTTP response.
	// See: https://golang.org/pkg/net/http/#ResponseWriter
	Response struct {
		Writer    http.ResponseWriter
		Status    int
		Size      int64
		Committed bool
		echo      *Echo
	}
)

// NewResponse creates a new instance of Response.
func NewResponse(w http.ResponseWriter, e *Echo) (r *Response) {
	return &Response{Writer: w, echo: e}
}

// Header returns the header map for the writer that will be sent by
// WriteHeader. Changing the header after a call to WriteHeader (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to WriteHeader (see example)
// To suppress implicit response headers, set their value to nil.
// Example: https://golang.org/pkg/net/http/#example_ResponseWriter_trailers
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly
// used to send error codes.
func (r *Response) WriteHeader(code int) {
	if r.Committed {
		r.echo.Logger.Warn("response already committed")
		return
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.Writer.Write(b)
	r.Size += int64(n)
	return
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
// This mechanism can be used to cancel long operations on the server if the
// client has disconnected before the response is ready.
// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r *Response) CloseNotify() <-chan bool {
	return r.Writer.(http.CloseNotifier).CloseNotify()
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Writer = w
	r.Size = 0
	r.Status = http.StatusOK
	r.Committed = false
}

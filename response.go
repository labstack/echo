package echo

import (
	"bufio"
	"log"
	"net"
	"net/http"

	"github.com/labstack/gommon/color"
)

type (
	Response struct {
		Writer    http.ResponseWriter
		status    int
		size      int64
		committed bool
	}
)

func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

func (r *Response) WriteHeader(code int) {
	if r.committed {
		// TODO: Warning
		log.Printf("echo => %s", color.Yellow("response already committed"))
		return
	}
	r.status = code
	r.Writer.WriteHeader(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.Writer.Write(b)
	r.size += int64(n)
	return n, err
}

// Flush wraps response writer's Flush function.
func (r *Response) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// Hijack wraps response writer's Hijack function.
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

// CloseNotify wraps response writer's CloseNotify function.
func (r *Response) CloseNotify() <-chan bool {
	return r.Writer.(http.CloseNotifier).CloseNotify()
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int64 {
	return r.size
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Writer = w
	r.status = http.StatusOK
	r.committed = false
}

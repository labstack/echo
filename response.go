package bolt

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type (
	// ResponseWriter interface {
	// }

	response struct {
		http.ResponseWriter
		status int
		size   int
	}
)

func NewResponse(rw http.ResponseWriter) *response {
	return &response{
		ResponseWriter: rw,
		status:         http.StatusOK,
	}
}

func (r *response) WriteHeader(c int) {
	r.status = c
	r.ResponseWriter.WriteHeader(c)
}

func (r *response) Write(b []byte) (n int, err error) {
	n, err = r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func (r *response) CloseNotify() <-chan bool {
	cn, ok := r.ResponseWriter.(http.CloseNotifier)
	if !ok {
		return nil
	}
	return cn.CloseNotify()
}

func (r *response) Flusher() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker interface not supported")
	}
	return h.Hijack()
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Size() int {
	return r.size
}

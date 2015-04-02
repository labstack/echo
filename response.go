package echo

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
)

type (
	response struct {
		http.ResponseWriter
		status    int
		size      int
		committed bool
	}
)

func (r *response) WriteHeader(n int) {
	// TODO: fix when halted.
	if r.committed {
		// TODO: Warning
		log.Println("echo: response already committed")
		return
	}
	r.status = n
	r.ResponseWriter.WriteHeader(n)
	r.committed = true
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
		return nil, nil, errors.New("echo: hijacker interface not supported")
	}
	return h.Hijack()
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Size() int {
	return r.size
}

func (r *response) reset(rw http.ResponseWriter) {
	r.ResponseWriter = rw
	r.committed = false
}

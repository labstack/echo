package echo

import (
	"log"
	"net/http"
)

type (
	response struct {
		writer    http.ResponseWriter
		status    int
		size      int
		committed bool
	}
)

func (r *response) Header() http.Header {
	return r.writer.Header()
}

func (r *response) WriteHeader(n int) {
	if r.committed {
		// TODO: Warning
		log.Println("echo: response already committed")
		return
	}
	r.status = n
	r.writer.WriteHeader(n)
	r.committed = true
}

func (r *response) Write(b []byte) (n int, err error) {
	n, err = r.writer.Write(b)
	r.size += n
	return n, err
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Size() int {
	return r.size
}

func (r *response) reset(w http.ResponseWriter) {
	r.writer = w
	r.committed = false
}

package echo

import (
	"log"
	"net/http"

	"github.com/labstack/gommon/color"
)

type (
	Response struct {
		Writer    http.ResponseWriter
		status    int
		size      int
		committed bool
	}
)

func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

func (r *Response) WriteHeader(n int) {
	if r.committed {
		// TODO: Warning
		log.Printf("echo: %s", color.Yellow("response already committed"))
		return
	}
	r.status = n
	r.Writer.WriteHeader(n)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.Writer.Write(b)
	r.size += n
	return n, err
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int {
	return r.size
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Writer = w
	r.committed = false
}

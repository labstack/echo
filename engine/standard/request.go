package standard

import (
	"io"
	"net/http"

	"github.com/labstack/echo/engine"
)

type (
	Request struct {
		*http.Request
		url    engine.URL
		header engine.Header
	}
)

func NewRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
		url:     &URL{URL: r.URL},
		header:  &Header{r.Header},
	}
}

func (r *Request) TLS() bool {
	return r.Request.TLS != nil
}

func (r *Request) Scheme() string {
	if r.TLS() {
		return "https"
	}
	return "http"
}

func (r *Request) Host() string {
	return r.Request.Host
}

func (r *Request) URL() engine.URL {
	return r.url
}

func (r *Request) Header() engine.Header {
	return r.header
}

// func Proto() string {
// 	return r.request.Proto()
// }
//
// func ProtoMajor() int {
// 	return r.request.ProtoMajor()
// }
//
// func ProtoMinor() int {
// 	return r.request.ProtoMinor()
// }

func (r *Request) RemoteAddress() string {
	return r.RemoteAddr
}

func (r *Request) Method() string {
	return r.Request.Method
}

func (r *Request) URI() string {
	return r.RequestURI
}

func (r *Request) Body() io.ReadCloser {
	return r.Request.Body
}

func (r *Request) FormValue(name string) string {
	return r.Request.FormValue(name)
}

func (r *Request) reset(req *http.Request, h engine.Header, u engine.URL) {
	r.Request = req
	r.header = h
	r.url = u
}

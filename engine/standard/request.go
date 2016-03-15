package standard

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/engine"
)

type (
	// Request implements `engine.Request`.
	Request struct {
		*http.Request
		url    engine.URL
		header engine.Header
	}
)

// TLS implements `engine.Request#TLS` method.
func (r *Request) TLS() bool {
	return r.Request.TLS != nil
}

// Scheme implements `engine.Request#Scheme` method.
func (r *Request) Scheme() string {
	if r.TLS() {
		return "https"
	}
	return "http"
}

// Host implements `engine.Request#Host` method.
func (r *Request) Host() string {
	return r.Request.Host
}

// URL implements `engine.Request#URL` method.
func (r *Request) URL() engine.URL {
	return r.url
}

// Header implements `engine.Request#URL` method.
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

// UserAgent implements `engine.Request#UserAgent` method.
func (r *Request) UserAgent() string {
	return r.Request.UserAgent()
}

// RemoteAddress implements `engine.Request#RemoteAddress` method.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr
}

// Method implements `engine.Request#Method` method.
func (r *Request) Method() string {
	return r.Request.Method
}

// SetMethod implements `engine.Request#SetMethod` method.
func (r *Request) SetMethod(method string) {
	r.Request.Method = method
}

// URI implements `engine.Request#URI` method.
func (r *Request) URI() string {
	return r.RequestURI
}

// Body implements `engine.Request#Body` method.
func (r *Request) Body() io.ReadCloser {
	return r.Request.Body
}

// FormValue implements `engine.Request#FormValue` method.
func (r *Request) FormValue(name string) string {
	return r.Request.FormValue(name)
}

// FormFile implements `engine.Request#FormFile` method.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := r.Request.FormFile(name)
	return fh, err
}

// MultipartForm implements `engine.Request#MultipartForm` method.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	r.Request.ParseMultipartForm(32 << 20) // 32 MB
	return r.Request.MultipartForm, nil
}

func (r *Request) reset(req *http.Request, h engine.Header, u engine.URL) {
	r.Request = req
	r.header = h
	r.url = u
}

package standard

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	// Request implements `engine.Request`.
	Request struct {
		*http.Request
		url    engine.URL
		header engine.Header
		logger *log.Logger
	}
)

// IsTLS implements `engine.Request#TLS` function.
func (r *Request) IsTLS() bool {
	return r.Request.TLS != nil
}

// Scheme implements `engine.Request#Scheme` function.
func (r *Request) Scheme() string {
	if r.IsTLS() {
		return "https"
	}
	return "http"
}

// Host implements `engine.Request#Host` function.
func (r *Request) Host() string {
	return r.Request.Host
}

// URL implements `engine.Request#URL` function.
func (r *Request) URL() engine.URL {
	return r.url
}

// Header implements `engine.Request#URL` function.
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

// ContentLength implements `engine.Request#ContentLength` function.
func (r *Request) ContentLength() int {
	return int(r.Request.ContentLength)
}

// UserAgent implements `engine.Request#UserAgent` function.
func (r *Request) UserAgent() string {
	return r.Request.UserAgent()
}

// RemoteAddress implements `engine.Request#RemoteAddress` function.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr
}

// Method implements `engine.Request#Method` function.
func (r *Request) Method() string {
	return r.Request.Method
}

// SetMethod implements `engine.Request#SetMethod` function.
func (r *Request) SetMethod(method string) {
	r.Request.Method = method
}

// URI implements `engine.Request#URI` function.
func (r *Request) URI() string {
	return r.RequestURI
}

// Body implements `engine.Request#Body` function.
func (r *Request) Body() io.Reader {
	return r.Request.Body
}

// FormValue implements `engine.Request#FormValue` function.
func (r *Request) FormValue(name string) string {
	return r.Request.FormValue(name)
}

// FormParams implements `engine.Request#FormParams` function.
func (r *Request) FormParams() map[string][]string {
	if err := r.ParseForm(); err != nil {
		r.logger.Error(err)
	}
	return map[string][]string(r.Request.PostForm)
}

// FormFile implements `engine.Request#FormFile` function.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := r.Request.FormFile(name)
	return fh, err
}

// MultipartForm implements `engine.Request#MultipartForm` function.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	err := r.ParseMultipartForm(32 << 20) // 32 MB
	return r.Request.MultipartForm, err
}

func (r *Request) reset(rq *http.Request, h engine.Header, u engine.URL) {
	r.Request = rq
	r.header = h
	r.url = u
}

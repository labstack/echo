package test

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/engine"
)

type (
	Request struct {
		request *http.Request
		url     engine.URL
		header  engine.Header
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
)

func NewRequest(method, url string, body io.Reader) engine.Request {
	r, _ := http.NewRequest(method, url, body)
	return &Request{
		request: r,
		url:     &URL{url: r.URL},
		header:  &Header{r.Header},
	}
}

func (r *Request) IsTLS() bool {
	return r.request.TLS != nil
}

func (r *Request) Scheme() string {
	if r.IsTLS() {
		return "https"
	}
	return "http"
}

func (r *Request) Host() string {
	return r.request.Host
}

func (r *Request) SetHost(host string) {
	r.request.Host = host
}

func (r *Request) URL() engine.URL {
	return r.url
}

func (r *Request) Header() engine.Header {
	return r.header
}

func (r *Request) Referer() string {
	return r.request.Referer()
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

func (r *Request) ContentLength() int64 {
	return r.request.ContentLength
}

func (r *Request) UserAgent() string {
	return r.request.UserAgent()
}

func (r *Request) RemoteAddress() string {
	return r.request.RemoteAddr
}

func (r *Request) RealIP() string {
	ra := r.RemoteAddress()
	if ip := r.Header().Get("X-Forwarded-For"); ip != "" {
		ra = ip
	} else if ip := r.Header().Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

func (r *Request) Method() string {
	return r.request.Method
}

func (r *Request) SetMethod(method string) {
	r.request.Method = method
}

func (r *Request) URI() string {
	return r.request.RequestURI
}

func (r *Request) SetURI(uri string) {
	r.request.RequestURI = uri
}

func (r *Request) Body() io.Reader {
	return r.request.Body
}

func (r *Request) SetBody(reader io.Reader) {
	r.request.Body = ioutil.NopCloser(reader)
}

func (r *Request) FormValue(name string) string {
	return r.request.FormValue(name)
}

func (r *Request) FormParams() map[string][]string {
	if strings.HasPrefix(r.header.Get("Content-Type"), "multipart/form-data") {
		if err := r.request.ParseMultipartForm(defaultMemory); err != nil {
			panic(err)
		}
	} else {
		if err := r.request.ParseForm(); err != nil {
			panic(err)
		}
	}
	return map[string][]string(r.request.Form)
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := r.request.FormFile(name)
	return fh, err
}

func (r *Request) MultipartForm() (*multipart.Form, error) {
	err := r.request.ParseMultipartForm(defaultMemory)
	return r.request.MultipartForm, err
}

func (r *Request) Cookie(name string) (engine.Cookie, error) {
	c, err := r.request.Cookie(name)
	if err != nil {
		return nil, errors.New("cookie not found")
	}
	return &Cookie{c}, nil
}

// Cookies implements `engine.Request#Cookies` function.
func (r *Request) Cookies() []engine.Cookie {
	cs := r.request.Cookies()
	cookies := make([]engine.Cookie, len(cs))
	for i, c := range cs {
		cookies[i] = &Cookie{c}
	}
	return cookies
}

func (r *Request) reset(req *http.Request, h engine.Header, u engine.URL) {
	r.request = req
	r.header = h
	r.url = u
}

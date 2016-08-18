package standard

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
)

type (
	// Request implements `engine.Request`.
	Request struct {
		*http.Request
		header engine.Header
		url    engine.URL
		logger log.Logger
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
)

// NewRequest returns `Request` instance.
func NewRequest(r *http.Request, l log.Logger) *Request {
	return &Request{
		Request: r,
		url:     &URL{URL: r.URL},
		header:  &Header{Header: r.Header},
		logger:  l,
	}
}

// IsTLS implements `engine.Request#TLS` function.
func (r *Request) IsTLS() bool {
	return r.Request.TLS != nil
}

// Scheme implements `engine.Request#Scheme` function.
func (r *Request) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
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

// Referer implements `engine.Request#Referer` function.
func (r *Request) Referer() string {
	return r.Request.Referer()
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
func (r *Request) ContentLength() int64 {
	return r.Request.ContentLength
}

// UserAgent implements `engine.Request#UserAgent` function.
func (r *Request) UserAgent() string {
	return r.Request.UserAgent()
}

// RemoteAddress implements `engine.Request#RemoteAddress` function.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr
}

// RealIP implements `engine.Request#RealIP` function.
func (r *Request) RealIP() string {
	ra := r.RemoteAddress()
	if ip := r.Header().Get(echo.HeaderXForwardedFor); ip != "" {
		ra = ip
	} else if ip := r.Header().Get(echo.HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
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

// SetURI implements `engine.Request#SetURI` function.
func (r *Request) SetURI(uri string) {
	r.RequestURI = uri
}

// Body implements `engine.Request#Body` function.
func (r *Request) Body() io.Reader {
	return r.Request.Body
}

// SetBody implements `engine.Request#SetBody` function.
func (r *Request) SetBody(reader io.Reader) {
	r.Request.Body = ioutil.NopCloser(reader)
}

// FormValue implements `engine.Request#FormValue` function.
func (r *Request) FormValue(name string) string {
	return r.Request.FormValue(name)
}

// FormParams implements `engine.Request#FormParams` function.
func (r *Request) FormParams() map[string][]string {
	if strings.HasPrefix(r.header.Get(echo.HeaderContentType), echo.MIMEMultipartForm) {
		if err := r.ParseMultipartForm(defaultMemory); err != nil {
			panic(fmt.Sprintf("echo: %v", err))
		}
	} else {
		if err := r.ParseForm(); err != nil {
			panic(fmt.Sprintf("echo: %v", err))
		}
	}
	return map[string][]string(r.Request.Form)
}

// FormFile implements `engine.Request#FormFile` function.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := r.Request.FormFile(name)
	return fh, err
}

// MultipartForm implements `engine.Request#MultipartForm` function.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	err := r.ParseMultipartForm(defaultMemory)
	return r.Request.MultipartForm, err
}

// Cookie implements `engine.Request#Cookie` function.
func (r *Request) Cookie(name string) (engine.Cookie, error) {
	c, err := r.Request.Cookie(name)
	if err != nil {
		return nil, echo.ErrCookieNotFound
	}
	return &Cookie{c}, nil
}

// Cookies implements `engine.Request#Cookies` function.
func (r *Request) Cookies() []engine.Cookie {
	cs := r.Request.Cookies()
	cookies := make([]engine.Cookie, len(cs))
	for i, c := range cs {
		cookies[i] = &Cookie{c}
	}
	return cookies
}

func (r *Request) reset(req *http.Request, h engine.Header, u engine.URL) {
	r.Request = req
	r.header = h
	r.url = u
}

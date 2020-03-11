// +build !appengine

package fasthttp

import (
	"bytes"
	"io"
	"mime/multipart"
	"net"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
	"github.com/valyala/fasthttp"
)

type (
	// Request implements `engine.Request`.
	Request struct {
		*fasthttp.RequestCtx
		header engine.Header
		url    engine.URL
		logger log.Logger
	}
)

// NewRequest returns `Request` instance.
func NewRequest(c *fasthttp.RequestCtx, l log.Logger) *Request {
	return &Request{
		RequestCtx: c,
		url:        &URL{URI: c.URI()},
		header:     &RequestHeader{RequestHeader: &c.Request.Header},
		logger:     l,
	}
}

// IsTLS implements `engine.Request#TLS` function.
func (r *Request) IsTLS() bool {
	return r.RequestCtx.IsTLS()
}

// Scheme implements `engine.Request#Scheme` function.
func (r *Request) Scheme() string {
	return string(r.RequestCtx.URI().Scheme())
}

// Host implements `engine.Request#Host` function.
func (r *Request) Host() string {
	return string(r.RequestCtx.Host())
}

// SetHost implements `engine.Request#SetHost` function.
func (r *Request) SetHost(host string) {
	r.RequestCtx.Request.SetHost(host)
}

// URL implements `engine.Request#URL` function.
func (r *Request) URL() engine.URL {
	return r.url
}

// Header implements `engine.Request#Header` function.
func (r *Request) Header() engine.Header {
	return r.header
}

// Referer implements `engine.Request#Referer` function.
func (r *Request) Referer() string {
	return string(r.Request.Header.Referer())
}

// ContentLength implements `engine.Request#ContentLength` function.
func (r *Request) ContentLength() int64 {
	return int64(r.Request.Header.ContentLength())
}

// UserAgent implements `engine.Request#UserAgent` function.
func (r *Request) UserAgent() string {
	return string(r.RequestCtx.UserAgent())
}

// RemoteAddress implements `engine.Request#RemoteAddress` function.
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr().String()
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
	return string(r.RequestCtx.Method())
}

// SetMethod implements `engine.Request#SetMethod` function.
func (r *Request) SetMethod(method string) {
	r.Request.Header.SetMethodBytes([]byte(method))
}

// URI implements `engine.Request#URI` function.
func (r *Request) URI() string {
	return string(r.RequestURI())
}

// SetURI implements `engine.Request#SetURI` function.
func (r *Request) SetURI(uri string) {
	r.Request.Header.SetRequestURI(uri)
}

// Body implements `engine.Request#Body` function.
func (r *Request) Body() io.Reader {
	return bytes.NewBuffer(r.Request.Body())
}

// SetBody implements `engine.Request#SetBody` function.
func (r *Request) SetBody(reader io.Reader) {
	r.Request.SetBodyStream(reader, 0)
}

// FormValue implements `engine.Request#FormValue` function.
func (r *Request) FormValue(name string) string {
	return string(r.RequestCtx.FormValue(name))
}

// FormParams implements `engine.Request#FormParams` function.
func (r *Request) FormParams() (params map[string][]string) {
	params = make(map[string][]string)
	mf, err := r.RequestCtx.MultipartForm()

	if err == fasthttp.ErrNoMultipartForm {
		r.PostArgs().VisitAll(func(k, v []byte) {
			key := string(k)
			if _, ok := params[key]; ok {
				params[key] = append(params[key], string(v))
			} else {
				params[string(k)] = []string{string(v)}
			}
		})
	} else if err == nil {
		for k, v := range mf.Value {
			if len(v) > 0 {
				params[k] = v
			}
		}
	}

	return
}

// FormFile implements `engine.Request#FormFile` function.
func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.RequestCtx.FormFile(name)
}

// MultipartForm implements `engine.Request#MultipartForm` function.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.RequestCtx.MultipartForm()
}

// Cookie implements `engine.Request#Cookie` function.
func (r *Request) Cookie(name string) (engine.Cookie, error) {
	c := new(fasthttp.Cookie)
	b := r.Request.Header.Cookie(name)
	if b == nil {
		return nil, echo.ErrCookieNotFound
	}
	c.SetKey(name)
	c.SetValueBytes(b)
	return &Cookie{c}, nil
}

// Cookies implements `engine.Request#Cookies` function.
func (r *Request) Cookies() []engine.Cookie {
	cookies := []engine.Cookie{}
	r.Request.Header.VisitAllCookie(func(name, value []byte) {
		c := new(fasthttp.Cookie)
		c.SetKeyBytes(name)
		c.SetValueBytes(value)
		cookies = append(cookies, &Cookie{c})
	})
	return cookies
}

func (r *Request) reset(c *fasthttp.RequestCtx, h engine.Header, u engine.URL) {
	r.RequestCtx = c
	r.header = h
	r.url = u
}

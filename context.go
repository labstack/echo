package echo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		// Request returns `*http.Request`.
		Request() *http.Request

		// SetRequest sets `*http.Request`.
		SetRequest(*http.Request)

		// Request returns `*Response`.
		Response() *Response

		// IsTLS returns true if HTTP connection is TLS otherwise false.
		IsTLS() bool

		// Scheme returns the HTTP protocol scheme, `http` or `https`.
		Scheme() string

		// RealIP returns the client's network address based on `X-Forwarded-For`
		// or `X-Real-IP` request header.
		RealIP() string

		// Path returns the registered path for the handler.
		Path() string

		// SetPath sets the registered path for the handler.
		SetPath(string)

		// Param returns path parameter by name.
		Param(string) string

		// ParamNames returns path parameter names.
		ParamNames() []string

		// SetParamNames sets path parameter names.
		SetParamNames(...string)

		// ParamValues returns path parameter values.
		ParamValues() []string

		// SetParamValues sets path parameter values.
		SetParamValues(...string)

		// QueryParam returns the query param for the provided name.
		QueryParam(string) string

		// QueryParams returns the query parameters as `url.Values`.
		QueryParams() url.Values

		// QueryString returns the URL query string.
		QueryString() string

		// FormValue returns the form field value for the provided name.
		FormValue(string) string

		// FormParams returns the form parameters as `url.Values`.
		FormParams() (url.Values, error)

		// FormFile returns the multipart form file for the provided name.
		FormFile(string) (*multipart.FileHeader, error)

		// MultipartForm returns the multipart form.
		MultipartForm() (*multipart.Form, error)

		// Cookie returns the named cookie provided in the request.
		Cookie(string) (*http.Cookie, error)

		// SetCookie adds a `Set-Cookie` header in HTTP response.
		SetCookie(*http.Cookie)

		// Cookies returns the HTTP cookies sent with the request.
		Cookies() []*http.Cookie

		// Get retrieves data from the context.
		Get(string) interface{}

		// Set saves data in the context.
		Set(string, interface{})

		// Bind binds the request body into provided type `i`. The default binder
		// does it based on Content-Type header.
		Bind(interface{}) error

		// Render renders a template with data and sends a text/html response with status
		// code. Templates can be registered using `Echo.Renderer`.
		Render(int, string, interface{}) error

		// HTML sends an HTTP response with status code.
		HTML(int, string) error

		// HTMLBlob sends an HTTP blob response with status code.
		HTMLBlob(int, []byte) error

		// String sends a string response with status code.
		String(int, string) error

		// JSON sends a JSON response with status code.
		JSON(int, interface{}) error

		// JSONBlob sends a JSON blob response with status code.
		JSONBlob(int, []byte) error

		// JSONP sends a JSONP response with status code. It uses `callback` to construct
		// the JSONP payload.
		JSONP(int, string, interface{}) error

		// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
		// to construct the JSONP payload.
		JSONPBlob(int, string, []byte) error

		// XML sends an XML response with status code.
		XML(int, interface{}) error

		// XMLBlob sends a XML blob response with status code.
		XMLBlob(int, []byte) error

		// Blob sends a blob response with status code and content type.
		Blob(int, string, []byte) error

		// Stream sends a streaming response with status code and content type.
		Stream(int, string, io.Reader) error

		// File sends a response with the content of the file.
		File(string) error

		// Attachment sends a response as attachment, prompting client to save the
		// file.
		Attachment(string, string) error

		// Inline sends a response as inline, opening the file in the browser.
		Inline(string, string) error

		// NoContent sends a response with no body and a status code.
		NoContent(int) error

		// Redirect redirects the request with status code.
		Redirect(int, string) error

		// Error invokes the registered HTTP error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// SetHandler sets the matched handler by router.
		SetHandler(HandlerFunc)

		// Logger returns the `Logger` instance.
		Logger() Logger

		// Echo returns the `Echo` instance.
		Echo() *Echo

		// Reset resets the context after request completes. It must be called along
		// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
		// See `Echo#ServeHTTP()`
		Reset(*http.Request, http.ResponseWriter)
	}

	context struct {
		request  *http.Request
		response *Response
		path     string
		pnames   []string
		pvalues  []string
		query    url.Values
		handler  HandlerFunc
		store    Map
		echo     *Echo
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
)

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) Response() *Response {
	return c.response
}

func (c *context) IsTLS() bool {
	return c.request.TLS != nil
}

func (c *context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	return "http"
}

func (c *context) RealIP() string {
	ra := c.request.RemoteAddr
	if ip := c.request.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = ip
	} else if ip := c.request.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

func (c *context) Path() string {
	return c.path
}

func (c *context) SetPath(p string) {
	c.path = p
}

func (c *context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}

			// Param name with aliases
			for _, p := range strings.Split(n, ",") {
				if p == name {
					return c.pvalues[i]
				}
			}
		}
	}
	return ""
}

func (c *context) ParamNames() []string {
	return c.pnames
}

func (c *context) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *context) ParamValues() []string {
	return c.pvalues
}

func (c *context) SetParamValues(values ...string) {
	c.pvalues = values
}

func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

func (c *context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

func (c *context) Get(key string) interface{} {
	return c.store[key]
}

func (c *context) Bind(i interface{}) error {
	return c.echo.Binder.Bind(i, c)
}

func (c *context) Render(code int, name string, data interface{}) (err error) {
	if c.echo.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.echo.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	c.response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(buf.Bytes())
	return
}

func (c *context) HTML(code int, html string) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, []byte(html))
}

func (c *context) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *context) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *context) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if c.echo.Debug {
		b, err = json.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

func (c *context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, b)
}

func (c *context) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.response.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.response.Write(b); err != nil {
		return
	}
	_, err = c.response.Write([]byte(");"))
	return
}

func (c *context) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if c.echo.Debug {
		b, err = xml.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

func (c *context) XMLBlob(code int, b []byte) (err error) {
	c.response.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

func (c *context) Blob(code int, contentType string, b []byte) (err error) {
	c.response.Header().Set(HeaderContentType, contentType)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

func (c *context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.response.Header().Set(HeaderContentType, contentType)
	c.response.WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

func (c *context) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return nil
}

func (c *context) Attachment(file, name string) (err error) {
	return c.contentDisposition(file, name, "attachment")
}

func (c *context) Inline(file, name string) (err error) {
	return c.contentDisposition(file, name, "inline")
}

func (c *context) contentDisposition(file, name, dispositionType string) (err error) {
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", dispositionType, name))
	c.File(file)
	return
}

func (c *context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(HeaderLocation, url)
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Error(err error) {
	c.echo.HTTPErrorHandler(err, c)
}

func (c *context) Echo() *Echo {
	return c.echo
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

func (c *context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *context) Logger() Logger {
	return c.echo.Logger
}

func (c *context) Reset(r *http.Request, w http.ResponseWriter) {
	c.query = nil
	c.store = nil
	c.request = r
	c.response.reset(w)
	c.handler = NotFoundHandler
}

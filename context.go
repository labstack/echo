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
		SetRequest(r *http.Request)

		// Response returns `*Response`.
		Response() *Response

		// IsTLS returns true if HTTP connection is TLS otherwise false.
		IsTLS() bool

		// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
		IsWebSocket() bool

		// Scheme returns the HTTP protocol scheme, `http` or `https`.
		Scheme() string

		// RealIP returns the client's network address based on `X-Forwarded-For`
		// or `X-Real-IP` request header.
		RealIP() string

		// Path returns the registered path for the handler.
		Path() string

		// SetPath sets the registered path for the handler.
		SetPath(p string)

		// Param returns path parameter by name.
		Param(name string) string

		// ParamNames returns path parameter names.
		ParamNames() []string

		// SetParamNames sets path parameter names.
		SetParamNames(names ...string)

		// ParamValues returns path parameter values.
		ParamValues() []string

		// SetParamValues sets path parameter values.
		SetParamValues(values ...string)

		// QueryParam returns the query param for the provided name.
		QueryParam(name string) string

		// QueryParams returns the query parameters as `url.Values`.
		QueryParams() url.Values

		// QueryString returns the URL query string.
		QueryString() string

		// FormValue returns the form field value for the provided name.
		FormValue(name string) string

		// FormParams returns the form parameters as `url.Values`.
		FormParams() (url.Values, error)

		// FormFile returns the multipart form file for the provided name.
		FormFile(name string) (*multipart.FileHeader, error)

		// MultipartForm returns the multipart form.
		MultipartForm() (*multipart.Form, error)

		// Cookie returns the named cookie provided in the request.
		Cookie(name string) (*http.Cookie, error)

		// SetCookie adds a `Set-Cookie` header in HTTP response.
		SetCookie(cookie *http.Cookie)

		// Cookies returns the HTTP cookies sent with the request.
		Cookies() []*http.Cookie

		// Get retrieves data from the context.
		Get(key string) interface{}

		// Set saves data in the context.
		Set(key string, val interface{})

		// Bind binds the request body into provided type `i`. The default binder
		// does it based on Content-Type header.
		Bind(i interface{}) error

		// Validate validates provided `i`. It is usually called after `Context#Bind()`.
		// Validator must be registered using `Echo#Validator`.
		Validate(i interface{}) error

		// Render renders a template with data and sends a text/html response with status
		// code. Renderer must be registered using `Echo.Renderer`.
		Render(code int, name string, data interface{}) error

		// HTML sends an HTTP response with status code.
		HTML(code int, html string) error

		// HTMLBlob sends an HTTP blob response with status code.
		HTMLBlob(code int, b []byte) error

		// String sends a string response with status code.
		String(code int, s string) error

		// JSON sends a JSON response with status code.
		JSON(code int, i interface{}) error

		// JSONPretty sends a pretty-print JSON with status code.
		JSONPretty(code int, i interface{}, indent string) error

		// JSONBlob sends a JSON blob response with status code.
		JSONBlob(code int, b []byte) error

		// JSONP sends a JSONP response with status code. It uses `callback` to construct
		// the JSONP payload.
		JSONP(code int, callback string, i interface{}) error

		// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
		// to construct the JSONP payload.
		JSONPBlob(code int, callback string, b []byte) error

		// XML sends an XML response with status code.
		XML(code int, i interface{}) error

		// XMLPretty sends a pretty-print XML with status code.
		XMLPretty(code int, i interface{}, indent string) error

		// XMLBlob sends an XML blob response with status code.
		XMLBlob(code int, b []byte) error

		// Blob sends a blob response with status code and content type.
		Blob(code int, contentType string, b []byte) error

		// Stream sends a streaming response with status code and content type.
		Stream(code int, contentType string, r io.Reader) error

		// File sends a response with the content of the file.
		File(file string) error

		// Attachment sends a response as attachment, prompting client to save the
		// file.
		Attachment(file string, name string) error

		// Inline sends a response as inline, opening the file in the browser.
		Inline(file string, name string) error

		// NoContent sends a response with no body and a status code.
		NoContent(code int) error

		// Redirect redirects the request to a provided URL with status code.
		Redirect(code int, url string) error

		// Error invokes the registered HTTP error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// SetHandler sets the matched handler by router.
		SetHandler(h HandlerFunc)

		// Logger returns the `Logger` instance.
		Logger() Logger

		// Echo returns the `Echo` instance.
		Echo() *Echo

		// Reset resets the context after request completes. It must be called along
		// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
		// See `Echo#ServeHTTP()`
		Reset(r *http.Request, w http.ResponseWriter)
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

func (c *context) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

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

func (c *context) IsWebSocket() bool {
	upgrade := c.request.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func (c *context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := c.request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (c *context) RealIP() string {
	if ip := c.request.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.request.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.request.RemoteAddr)
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
	return c.pvalues[:len(c.pnames)]
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

func (c *context) Get(key string) interface{} {
	return c.store[key]
}

func (c *context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

func (c *context) Bind(i interface{}) error {
	return c.echo.Binder.Bind(i, c)
}

func (c *context) Validate(i interface{}) error {
	if c.echo.Validator == nil {
		return ErrValidatorNotRegistered
	}
	return c.echo.Validator.Validate(i)
}

func (c *context) Render(code int, name string, data interface{}) (err error) {
	if c.echo.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.echo.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func (c *context) HTML(code int, html string) (err error) {
	return c.HTMLBlob(code, []byte(html))
}

func (c *context) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *context) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *context) JSON(code int, i interface{}) (err error) {
	_, pretty := c.QueryParams()["pretty"]
	if c.echo.Debug || pretty {
		return c.JSONPretty(code, i, "  ")
	}
	b, err := json.Marshal(i)
	if err != nil {
		return
	}
	return c.JSONBlob(code, b)
}

func (c *context) JSONPretty(code int, i interface{}, indent string) (err error) {
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return c.JSONBlob(code, b)
}

func (c *context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return
	}
	return c.JSONPBlob(code, callback, b)
}

func (c *context) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
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
	_, pretty := c.QueryParams()["pretty"]
	if c.echo.Debug || pretty {
		return c.XMLPretty(code, i, "  ")
	}
	b, err := xml.Marshal(i)
	if err != nil {
		return
	}
	return c.XMLBlob(code, b)
}

func (c *context) XMLPretty(code int, i interface{}, indent string) (err error) {
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return c.XMLBlob(code, b)
}

func (c *context) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

func (c *context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

func (c *context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

func (c *context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func (c *context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *context) contentDisposition(file, name, dispositionType string) error {
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
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
	c.request = r
	c.response.reset(w)
	c.query = nil
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil
	// NOTE: Don't reset because it has to have length c.echo.maxParam at all times
	// c.pvalues = nil
}

package echo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"bytes"

	"github.com/labstack/echo/context"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		context.Context

		// StdContext returns `context.Context`.
		StdContext() context.Context

		// SetStdContext sets `context.Context`.
		SetStdContext(context.Context)

		// Request returns `*http.Request`.
		Request() *http.Request

		// Request returns `*Response`.
		Response() *Response

		// WebSocket returns `*websocket.Conn`.
		WebSocket() *websocket.Conn

		// SetWebSocket sets `*websocket.Conn`.
		SetWebSocket(*websocket.Conn)

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

		// P returns path parameter by index.
		P(int) string

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
		// code. Templates can be registered using `Echo.SetRenderer()`.
		Render(int, string, interface{}) error

		// HTML sends an HTTP response with status code.
		HTML(int, string) error

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

		// Attachment sends a response from `io.ReaderSeeker` as attachment, prompting
		// client to save the file.
		Attachment(io.ReadSeeker, string) error

		// Inline sends a response from `io.ReaderSeeker` as inline, opening
		// the file in the browser.
		Inline(io.ReadSeeker, string) error

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

		// ServeContent sends static content from `io.Reader` and handles caching
		// via `If-Modified-Since` request header. It automatically sets `Content-Type`
		// and `Last-Modified` response headers.
		ServeContent(io.ReadSeeker, string, time.Time) error

		// Reset resets the context after request completes. It must be called along
		// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
		// See `Echo#ServeHTTP()`
		Reset(*http.Request, http.ResponseWriter)
	}

	echoContext struct {
		context   context.Context
		request   *http.Request
		response  *Response
		webSocket *websocket.Conn
		path      string
		pnames    []string
		pvalues   []string
		query     url.Values
		handler   HandlerFunc
		echo      *Echo
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
)

func (c *echoContext) StdContext() context.Context {
	return c.context
}

func (c *echoContext) SetStdContext(ctx context.Context) {
	c.context = ctx
}

func (c *echoContext) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

func (c *echoContext) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *echoContext) Err() error {
	return c.context.Err()
}

func (c *echoContext) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

func (c *echoContext) Request() *http.Request {
	return c.request
}

func (c *echoContext) Response() *Response {
	return c.response
}

func (c *echoContext) WebSocket() *websocket.Conn {
	return c.webSocket
}

func (c *echoContext) SetWebSocket(ws *websocket.Conn) {
	c.webSocket = ws
}

func (c *echoContext) IsTLS() bool {
	return c.request.TLS != nil
}

func (c *echoContext) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	return "http"
}

func (c *echoContext) RealIP() string {
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

func (c *echoContext) Path() string {
	return c.path
}

func (c *echoContext) SetPath(p string) {
	c.path = p
}

func (c *echoContext) P(i int) (value string) {
	l := len(c.pnames)
	if i < l {
		value = c.pvalues[i]
	}
	return
}

func (c *echoContext) Param(name string) (value string) {
	l := len(c.pnames)
	for i, n := range c.pnames {
		if n == name && i < l {
			value = c.pvalues[i]
			break
		}
	}
	return
}

func (c *echoContext) ParamNames() []string {
	return c.pnames
}

func (c *echoContext) SetParamNames(names ...string) {
	c.pnames = names
}

func (c *echoContext) ParamValues() []string {
	return c.pvalues
}

func (c *echoContext) SetParamValues(values ...string) {
	c.pvalues = values
}

func (c *echoContext) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *echoContext) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *echoContext) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *echoContext) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *echoContext) FormParams() (url.Values, error) {
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

func (c *echoContext) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

func (c *echoContext) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

func (c *echoContext) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *echoContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *echoContext) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *echoContext) Set(key string, val interface{}) {
	c.context = context.WithValue(c.context, key, val)
}

func (c *echoContext) Get(key string) interface{} {
	return c.context.Value(key)
}

func (c *echoContext) Bind(i interface{}) error {
	return c.echo.Binder.Bind(i, c)
}

func (c *echoContext) Render(code int, name string, data interface{}) (err error) {
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

func (c *echoContext) HTML(code int, html string) (err error) {
	c.response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(html))
	return
}

func (c *echoContext) String(code int, s string) (err error) {
	c.response.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(s))
	return
}

func (c *echoContext) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if c.echo.Debug {
		b, err = json.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

func (c *echoContext) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (c *echoContext) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, b)
}

func (c *echoContext) JSONPBlob(code int, callback string, b []byte) (err error) {
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

func (c *echoContext) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if c.echo.Debug {
		b, err = xml.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

func (c *echoContext) XMLBlob(code int, b []byte) (err error) {
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	return c.Blob(code, MIMEApplicationXMLCharsetUTF8, b)
}

func (c *echoContext) Blob(code int, contentType string, b []byte) (err error) {
	c.response.Header().Set(HeaderContentType, contentType)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

func (c *echoContext) Stream(code int, contentType string, r io.Reader) (err error) {
	c.response.Header().Set(HeaderContentType, contentType)
	c.response.WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

func (c *echoContext) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	return c.ServeContent(f, fi.Name(), fi.ModTime())
}

func (c *echoContext) Attachment(r io.ReadSeeker, name string) (err error) {
	return c.contentDisposition(r, name, "attachment")
}

func (c *echoContext) Inline(r io.ReadSeeker, name string) (err error) {
	return c.contentDisposition(r, name, "inline")
}

func (c *echoContext) contentDisposition(r io.ReadSeeker, name, dispositionType string) (err error) {
	c.response.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", dispositionType, name))
	c.response.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.response, r)
	return
}

func (c *echoContext) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *echoContext) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(HeaderLocation, url)
	c.response.WriteHeader(code)
	return nil
}

func (c *echoContext) Error(err error) {
	c.echo.HTTPErrorHandler(err, c)
}

func (c *echoContext) Echo() *Echo {
	return c.echo
}

func (c *echoContext) Handler() HandlerFunc {
	return c.handler
}

func (c *echoContext) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *echoContext) Logger() Logger {
	return c.echo.Logger
}

func (c *echoContext) ServeContent(content io.ReadSeeker, name string, modtime time.Time) error {
	req := c.Request()
	res := c.Response()

	if t, err := time.Parse(http.TimeFormat, req.Header.Get(HeaderIfModifiedSince)); err == nil && modtime.Before(t.Add(1*time.Second)) {
		res.Header().Del(HeaderContentType)
		res.Header().Del(HeaderContentLength)
		return c.NoContent(http.StatusNotModified)
	}

	res.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	res.Header().Set(HeaderLastModified, modtime.UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	_, err := io.Copy(res, content)
	return err
}

// ContentTypeByExtension returns the MIME type associated with the file based on
// its extension. It returns `application/octet-stream` incase MIME type is not
// found.
func ContentTypeByExtension(name string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
		t = MIMEOctetStream
	}
	return
}

func (c *echoContext) Reset(r *http.Request, w http.ResponseWriter) {
	// c.query = nil
	c.context = context.Background()
	c.request = r
	c.response.reset(w)
	c.handler = NotFoundHandler
}

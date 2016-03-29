package echo

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"

	"net/url"

	"bytes"

	netContext "golang.org/x/net/context"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		netContext.Context

		// NetContext returns `http://blog.golang.org/context.Context` interface.
		NetContext() netContext.Context

		// SetNetContext sets `http://blog.golang.org/context.Context` interface.
		SetNetContext(netContext.Context)

		// Request returns `engine.Request` interface.
		Request() engine.Request

		// Request returns `engine.Response` interface.
		Response() engine.Response

		// Path returns the registered path for the handler.
		Path() string

		// P returns path parameter by index.
		P(int) string

		// Param returns path parameter by name.
		Param(string) string

		// ParamNames returns path parameter names.
		ParamNames() []string

		// QueryParam returns the query param for the provided name. It is an alias
		// for `engine.URL#QueryParam()`.
		QueryParam(string) string

		// QueryParam returns the query parameters as map. It is an alias for `engine.URL#QueryParams()`.
		QueryParams() map[string][]string

		// FormValue returns the form field value for the provided name. It is an
		// alias for `engine.Request#FormValue()`.
		FormValue(string) string

		// FormParams returns the form parameters as map. It is an alias for `engine.Request#FormParams()`.
		FormParams() map[string][]string

		// FormFile returns the multipart form file for the provided name. It is an
		// alias for `engine.Request#FormFile()`.
		FormFile(string) (*multipart.FileHeader, error)

		// MultipartForm returns the multipart form. It is an alias for `engine.Request#MultipartForm()`.
		MultipartForm() (*multipart.Form, error)

		// Get retrieves data from the context.
		Get(string) interface{}

		// Set saves data in the context.
		Set(string, interface{})

		// Bind binds the request body into provided type `i`. The default binder does
		// it based on Content-Type header.
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

		// XML sends an XML response with status code.
		XML(int, interface{}) error

		// XMLBlob sends a XML blob response with status code.
		XMLBlob(int, []byte) error

		// File sends a response with the content of the file.
		File(string) error

		// Attachment sends a response from `io.ReaderSeeker` as attachment, prompting
		// client to save the file.
		Attachment(io.ReadSeeker, string) error

		// NoContent sends a response with no body and a status code.
		NoContent(int) error

		// Redirect redirects the request with status code.
		Redirect(int, string) error

		// Error invokes the registered HTTP error handler. Generally used by middleware.
		Error(err error)

		// Handler implements `Handler` interface.
		Handle(Context) error

		// Logger returns the `Logger` instance.
		Logger() *log.Logger

		// Echo returns the `Echo` instance.
		Echo() *Echo

		// ServeContent sends static content from `io.Reader` and handles caching
		// via `If-Modified-Since` request header. It automatically sets `Content-Type`
		// and `Last-Modified` response headers.
		ServeContent(io.ReadSeeker, string, time.Time) error

		// Object returns the `context` instance.
		Object() *context

		// Reset resets the context after request completes. It must be called along
		// with `Echo#GetContext()` and `Echo#PutContext()`. See `Echo#ServeHTTP()`
		Reset(engine.Request, engine.Response)
	}

	context struct {
		netContext netContext.Context
		request    engine.Request
		response   engine.Response
		path       string
		pnames     []string
		pvalues    []string
		query      url.Values
		store      store
		handler    Handler
		echo       *Echo
	}

	store map[string]interface{}
)

const (
	indexPage = "index.html"
)

// NewContext creates a Context object.
func NewContext(rq engine.Request, rs engine.Response, e *Echo) Context {
	return &context{
		request:  rq,
		response: rs,
		echo:     e,
		pvalues:  make([]string, *e.maxParam),
		store:    make(store),
		handler:  notFoundHandler,
	}
}

func (c *context) NetContext() netContext.Context {
	return c.netContext
}

func (c *context) SetNetContext(ctx netContext.Context) {
	c.netContext = ctx
}

func (c *context) Deadline() (deadline time.Time, ok bool) {
	return c.netContext.Deadline()
}

func (c *context) Done() <-chan struct{} {
	return c.netContext.Done()
}

func (c *context) Err() error {
	return c.netContext.Err()
}

func (c *context) Value(key interface{}) interface{} {
	return c.netContext.Value(key)
}

func (c *context) Handle(ctx Context) error {
	return c.handler.Handle(ctx)
}

func (c *context) Request() engine.Request {
	return c.request
}

func (c *context) Response() engine.Response {
	return c.response
}

func (c *context) Path() string {
	return c.path
}

func (c *context) P(i int) (value string) {
	l := len(c.pnames)
	if i < l {
		value = c.pvalues[i]
	}
	return
}

func (c *context) Param(name string) (value string) {
	l := len(c.pnames)
	for i, n := range c.pnames {
		if n == name && i < l {
			value = c.pvalues[i]
			break
		}
	}
	return
}

func (c *context) ParamNames() []string {
	return c.pnames
}

func (c *context) QueryParam(name string) string {
	return c.request.URL().QueryParam(name)
}

func (c *context) QueryParams() map[string][]string {
	return c.request.URL().QueryParams()
}

func (c *context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *context) FormParams() map[string][]string {
	return c.request.FormParams()
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.request.FormFile(name)
}

func (c *context) MultipartForm() (*multipart.Form, error) {
	return c.request.MultipartForm()
}

func (c *context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = val
}

func (c *context) Get(key string) interface{} {
	return c.store[key]
}

func (c *context) Bind(i interface{}) error {
	return c.echo.binder.Bind(i, c)
}

func (c *context) Render(code int, name string, data interface{}) (err error) {
	if c.echo.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.echo.renderer.Render(buf, name, data, c); err != nil {
		return
	}
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(buf.Bytes())
	return
}

func (c *context) HTML(code int, html string) (err error) {
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(html))
	return
}

func (c *context) String(code int, s string) (err error) {
	c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(s))
	return
}

func (c *context) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if c.echo.Debug() {
		b, err = json.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

func (c *context) JSONBlob(code int, b []byte) (err error) {
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

func (c *context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
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
	if c.echo.Debug() {
		b, err = xml.MarshalIndent(i, "", "  ")
	}
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

func (c *context) XMLBlob(code int, b []byte) (err error) {
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
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
		file = path.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		fi, _ = f.Stat()
	}
	return c.ServeContent(f, fi.Name(), fi.ModTime())
}

func (c *context) Attachment(r io.ReadSeeker, name string) (err error) {
	c.response.Header().Set(ContentType, ContentTypeByExtension(name))
	c.response.Header().Set(ContentDisposition, "attachment; filename="+name)
	c.response.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.response, r)
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
	c.response.Header().Set(Location, url)
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

func (c *context) Echo() *Echo {
	return c.echo
}

func (c *context) Logger() *log.Logger {
	return c.echo.logger
}

func (c *context) Object() *context {
	return c
}

func (c *context) ServeContent(content io.ReadSeeker, name string, modtime time.Time) error {
	rq := c.Request()
	rs := c.Response()

	if t, err := time.Parse(http.TimeFormat, rq.Header().Get(IfModifiedSince)); err == nil && modtime.Before(t.Add(1*time.Second)) {
		rs.Header().Del(ContentType)
		rs.Header().Del(ContentLength)
		return c.NoContent(http.StatusNotModified)
	}

	rs.Header().Set(ContentType, ContentTypeByExtension(name))
	rs.Header().Set(LastModified, modtime.UTC().Format(http.TimeFormat))
	rs.WriteHeader(http.StatusOK)
	_, err := io.Copy(rs, content)
	return err
}

// ContentTypeByExtension returns the MIME type associated with the file based on
// its extension. It returns `application/octet-stream` incase MIME type is not
// found.
func ContentTypeByExtension(name string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
		t = OctetStream
	}
	return
}

func (c *context) Reset(rq engine.Request, rs engine.Response) {
	c.netContext = nil
	c.request = rq
	c.response = rs
	c.query = nil
	c.store = nil
	c.handler = notFoundHandler
}

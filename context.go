package echo

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"

	"net/url"

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

		// QueryParams returns the query parameters as map. It is an alias for `engine.URL#QueryParams()`.
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

		// Del data from the context
		Del(string)

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

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// Logger returns the `Logger` instance.
		Logger() *log.Logger

		// Echo returns the `Echo` instance.
		Echo() *Echo

		// ServeContent sends static content from `io.Reader` and handles caching
		// via `If-Modified-Since` request header. It automatically sets `Content-Type`
		// and `Last-Modified` response headers.
		ServeContent(io.ReadSeeker, string, time.Time) error

		// Reset resets the context after request completes. It must be called along
		// with `Echo#GetContext()` and `Echo#PutContext()`. See `Echo#ServeHTTP()`
		Reset(engine.Request, engine.Response)
	}

	context struct {
		ResponseWriter
		RequestReader
		NetContextEmbedder
		MapStore

		path    string
		pnames  []string
		pvalues []string
		query   url.Values
		handler HandlerFunc
		echo    *Echo
	}
)

const (
	indexPage = "index.html"
)

// NewContext creates a Context object.
func NewContext(rq engine.Request, rs engine.Response, e *Echo) Context {
	return &context{
		RequestReader:  RequestReader{rq},
		ResponseWriter: ResponseWriter{rs},

		echo:    e,
		pvalues: make([]string, *e.maxParam),
		handler: notFoundHandler,
	}
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

// Set optimizes by only creating the MapStore if this method is called.
func (c *context) Set(key string, value interface{}) {
	if c.MapStore == nil {
		c.MapStore = make(MapStore)
	}
	c.MapStore.Set(key, value)
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
	res := c.Response()
	res.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	res.WriteHeader(code)
	_, err = res.Write(buf.Bytes())
	return err
}

func (c *context) File(file string) error {
	return c.ResponseWriter.File(c.Request(), file)
}

func (c *context) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

func (c *context) Echo() *Echo {
	return c.echo
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

func (c *context) Logger() *log.Logger {
	return c.echo.logger
}

func (c *context) ServeContent(content io.ReadSeeker, name string, modtime time.Time) error {
	rq := c.Request()
	rs := c.Response()

	return ServeContent(rq, rs, content, name, modtime)
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

func (c *context) Reset(rq engine.Request, rs engine.Response) {
	c.ResponseWriter.Res = rs
	c.RequestReader.Req = rq
	c.NetContextEmbedder.Ctx = nil
	// TODO(aarondl): Is this supposed to be nil?
	// Constructor makes one on New(), seems like memory optimization
	// to not create one unless it's used.
	c.MapStore = nil

	c.query = nil
	c.handler = notFoundHandler
}

package echo

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime"
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
	// Context represents context for the current request. It holds request and
	// response objects, path parameters, data and registered handler.
	Context interface {
		netContext.Context
		NetContext() netContext.Context
		SetNetContext(netContext.Context)
		Request() engine.Request
		Response() engine.Response
		Path() string
		P(int) string
		Param(string) string
		ParamNames() []string
		Query(string) string
		Form(string) string
		Get(string) interface{}
		Set(string, interface{})
		Bind(interface{}) error
		Render(int, string, interface{}) error
		HTML(int, string) error
		String(int, string) error
		JSON(int, interface{}) error
		JSONBlob(int, []byte) error
		JSONP(int, string, interface{}) error
		XML(int, interface{}) error
		XMLBlob(int, []byte) error
		File(string) error
		Attachment(io.Reader, string) error
		NoContent(int) error
		Redirect(int, string) error
		Error(err error)
		Handle(Context) error
		Logger() *log.Logger
		Echo() *Echo
		Object() *context
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
func NewContext(req engine.Request, res engine.Response, e *Echo) Context {
	return &context{
		request:  req,
		response: res,
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

// Request returns *http.Request.
func (c *context) Request() engine.Request {
	return c.request
}

// Response returns `engine.Response`.
func (c *context) Response() engine.Response {
	return c.response
}

// Path returns the registered path for the handler.
func (c *context) Path() string {
	return c.path
}

// P returns path parameter by index.
func (c *context) P(i int) (value string) {
	l := len(c.pnames)
	if i < l {
		value = c.pvalues[i]
	}
	return
}

// Param returns path parameter by name.
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

// ParamNames returns path parameter names.
func (c *context) ParamNames() []string {
	return c.pnames
}

// Query returns query parameter by name.
func (c *context) Query(name string) string {
	return c.request.URL().QueryValue(name)
}

// Form returns form parameter by name.
func (c *context) Form(name string) string {
	return c.request.FormValue(name)
}

// Set saves data in the context.
func (c *context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = val
}

// Get retrieves data from the context.
func (c *context) Get(key string) interface{} {
	return c.store[key]
}

// Bind binds the request body into specified type `i`. The default binder does
// it based on Content-Type header.
func (c *context) Bind(i interface{}) error {
	return c.echo.binder.Bind(i, c)
}

// Render renders a template with data and sends a text/html response with status
// code. Templates can be registered using `Echo.SetRenderer()`.
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

// HTML sends an HTTP response with status code.
func (c *context) HTML(code int, html string) (err error) {
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(html))
	return
}

// String sends a string response with status code.
func (c *context) String(code int, s string) (err error) {
	c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write([]byte(s))
	return
}

// JSON sends a JSON response with status code.
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

// JSONBlob sends a JSON blob response with status code.
func (c *context) JSONBlob(code int, b []byte) (err error) {
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
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

// XML sends an XML response with status code.
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

// XMLBlob sends a XML blob response with status code.
func (c *context) XMLBlob(code int, b []byte) (err error) {
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

// File sends a response with the content of the file.
func (c *context) File(file string) error {
	root, file := filepath.Split(file)
	fs := http.Dir(root)
	f, err := fs.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = path.Join(file, "index.html")
		f, err = fs.Open(file)
		if err != nil {
			return ErrNotFound
		}
		fi, _ = f.Stat()
	}

	return ServeContent(c.Request(), c.Response(), f, fi)
}

// Attachment sends a response from `io.Reader` as attachment, prompting client
// to save the file.
func (c *context) Attachment(r io.Reader, name string) (err error) {
	c.response.Header().Set(ContentType, detectContentType(name))
	c.response.Header().Set(ContentDisposition, "attachment; filename="+name)
	c.response.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.response, r)
	return
}

// NoContent sends a response with no body and a status code.
func (c *context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request with status code.
func (c *context) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(Location, url)
	c.response.WriteHeader(code)
	return nil
}

// Error invokes the registered HTTP error handler. Generally used by middleware.
func (c *context) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

// Echo returns the `Echo` instance.
func (c *context) Echo() *Echo {
	return c.echo
}

// Logger returns the `Logger` instance.
func (c *context) Logger() *log.Logger {
	return c.echo.logger
}

// Object returns the `context` object.
func (c *context) Object() *context {
	return c
}

func ServeContent(req engine.Request, res engine.Response, f http.File, fi os.FileInfo) error {
	res.Header().Set(ContentType, detectContentType(fi.Name()))
	res.Header().Set(LastModified, fi.ModTime().UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	_, err := io.Copy(res, f)
	return err
}

func detectContentType(name string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
		t = OctetStream
	}
	return
}

func (c *context) reset(req engine.Request, res engine.Response) {
	c.request = req
	c.response = res
	c.query = nil
	c.store = nil
	c.handler = notFoundHandler
}

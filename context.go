package echo

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"path/filepath"

	"net/url"

	"bytes"

	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

type (
	// Context represents context for the current request. It holds request and
	// response objects, path parameters, data and registered handler.
	Context struct {
		context.Context
		request  *http.Request
		response *Response
		socket   *websocket.Conn
		path     string
		pnames   []string
		pvalues  []string
		query    url.Values
		store    store
		echo     *Echo
	}
	store map[string]interface{}
)

// NewContext creates a Context object.
func NewContext(req *http.Request, res *Response, e *Echo) *Context {
	return &Context{
		request:  req,
		response: res,
		echo:     e,
		pvalues:  make([]string, *e.maxParam),
		store:    make(store),
	}
}

// Request returns *http.Request.
func (c *Context) Request() *http.Request {
	return c.request
}

// Response returns *Response.
func (c *Context) Response() *Response {
	return c.response
}

// Socket returns *websocket.Conn.
func (c *Context) Socket() *websocket.Conn {
	return c.socket
}

// Path returns the registered path for the handler.
func (c *Context) Path() string {
	return c.path
}

// P returns path parameter by index.
func (c *Context) P(i int) (value string) {
	l := len(c.pnames)
	if i < l {
		value = c.pvalues[i]
	}
	return
}

// Param returns path parameter by name.
func (c *Context) Param(name string) (value string) {
	l := len(c.pnames)
	for i, n := range c.pnames {
		if n == name && i < l {
			value = c.pvalues[i]
			break
		}
	}
	return
}

// Query returns query parameter by name.
func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

// Form returns form parameter by name.
func (c *Context) Form(name string) string {
	return c.request.FormValue(name)
}

// Get retrieves data from the context.
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *Context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = val
}

// Bind binds the request body into specified type `i`. The default binder does
// it based on Content-Type header.
func (c *Context) Bind(i interface{}) error {
	return c.echo.binder.Bind(c.request, i)
}

// Render renders a template with data and sends a text/html response with status
// code. Templates can be registered using `Echo.SetRenderer()`.
func (c *Context) Render(code int, name string, data interface{}) (err error) {
	if c.echo.renderer == nil {
		return RendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.echo.renderer.Render(buf, name, data); err != nil {
		return
	}
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write(buf.Bytes())
	return
}

// HTML sends an HTTP response with status code.
func (c *Context) HTML(code int, html string) (err error) {
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(html))
	return
}

// String sends a string response with status code.
func (c *Context) String(code int, s string) (err error) {
	c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(s))
	return
}

// JSON sends a JSON response with status code.
func (c *Context) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write(b)
	return
}

// JSONIndent sends a JSON response with status code, but it applies prefix and indent to format the output.
func (c *Context) JSONIndent(code int, i interface{}, prefix string, indent string) (err error) {
	b, err := json.MarshalIndent(i, prefix, indent)
	if err != nil {
		return err
	}
	c.json(code, b)
	return
}

func (c *Context) json(code int, b []byte) {
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write(b)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *Context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(callback + "("))
	c.response.Write(b)
	c.response.Write([]byte(");"))
	return
}

// XML sends an XML response with status code.
func (c *Context) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(xml.Header))
	c.response.Write(b)
	return
}

// XMLIndent sends an XML response with status code, but it applies prefix and indent to format the output.
func (c *Context) XMLIndent(code int, i interface{}, prefix string, indent string) (err error) {
	b, err := xml.MarshalIndent(i, prefix, indent)
	if err != nil {
		return err
	}
	c.xml(code, b)
	return
}

func (c *Context) xml(code int, b []byte) {
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(xml.Header))
	c.response.Write(b)
}

// File sends a response with the content of the file. If `attachment` is set
// to true, the client is prompted to save the file with provided `name`,
// name can be empty, in that case name of the file is used.
func (c *Context) File(path, name string, attachment bool) (err error) {
	dir, file := filepath.Split(path)
	if attachment {
		c.response.Header().Set(ContentDisposition, "attachment; filename="+name)
	}
	if err = c.echo.serveFile(dir, file, c); err != nil {
		c.response.Header().Del(ContentDisposition)
	}
	return
}

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request using http.Redirect with status code.
func (c *Context) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return InvalidRedirectCode
	}
	http.Redirect(c.response, c.request, url, code)
	return nil
}

// Error invokes the registered HTTP error handler. Generally used by middleware.
func (c *Context) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

// Echo returns the `Echo` instance.
func (c *Context) Echo() *Echo {
	return c.echo
}

func (c *Context) reset(r *http.Request, w http.ResponseWriter, e *Echo) {
	c.request = r
	c.response.reset(w, e)
	c.query = nil
	c.store = nil
	c.echo = e
}

package echo

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"fmt"

	"net/url"

	"os"

	"golang.org/x/net/websocket"
)

type (
	// Context represents context for the current request. It holds request and
	// response objects, path parameters, data and registered handler.
	Context struct {
		request  *http.Request
		response *Response
		socket   *websocket.Conn
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
func (c *Context) Render(code int, name string, data interface{}) error {
	if c.echo.renderer == nil {
		return RendererNotRegistered
	}
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	return c.echo.renderer.Render(c.response, name, data)
}

// HTML formats according to a format specifier and sends HTML response with
// status code.
func (c *Context) HTML(code int, format string, a ...interface{}) (err error) {
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = fmt.Fprintf(c.response, format, a...)
	return
}

// String formats according to a format specifier and sends text response with status
// code.
func (c *Context) String(code int, format string, a ...interface{}) (err error) {
	c.response.Header().Set(ContentType, TextPlain)
	c.response.WriteHeader(code)
	_, err = fmt.Fprintf(c.response, format, a...)
	return
}

// JSON sends a JSON response with status code.
func (c *Context) JSON(code int, i interface{}) error {
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(i)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *Context) JSONP(code int, callback string, i interface{}) (err error) {
	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(callback + "("))
	if err = json.NewEncoder(c.response).Encode(i); err == nil {
		c.response.Write([]byte(");"))
	}
	return
}

// XML sends an XML response with status code.
func (c *Context) XML(code int, i interface{}) error {
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(xml.Header))
	return xml.NewEncoder(c.response).Encode(i)
}

// File sends a file as attachment.
func (c *Context) File(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	fi, _ := file.Stat()
	c.response.Header().Set(ContentDisposition, "attachment; filename="+fi.Name())
	_, err = io.Copy(c.response, file)
	return err
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

func (c *Context) reset(r *http.Request, w http.ResponseWriter, e *Echo) {
	c.request = r
	c.response.reset(w)
	c.query = nil
	c.store = nil
	c.echo = e
}

package echo

import (
	"encoding/json"
	"net/http"

	"fmt"

	"golang.org/x/net/websocket"
	"net/url"
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
	c.store[key] = val
}

// Bind binds the request body into specified type v. Default binder does it
// based on Content-Type header.
func (c *Context) Bind(i interface{}) error {
	return c.echo.binder(c.request, i)
}

// Render invokes the registered HTML template renderer and sends a text/html
// response with status code.
func (c *Context) Render(code int, name string, data interface{}) error {
	if c.echo.renderer == nil {
		return RendererNotRegistered
	}
	c.response.Header().Set(ContentType, TextHTML)
	c.response.WriteHeader(code)
	return c.echo.renderer.Render(c.response, name, data)
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(code int, i interface{}) error {
	c.response.Header().Set(ContentType, ApplicationJSON)
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(i)
}

// String formats according to a format specifier and sends text/plain response
// with status code.
func (c *Context) String(code int, format string, a ...interface{}) error {
	c.response.Header().Set(ContentType, TextPlain)
	c.response.WriteHeader(code)
	_, err := fmt.Fprintf(c.response, format, a...)
	return err
}

// HTML formats according to a format specifier and sends text/html response with
// status code.
func (c *Context) HTML(code int, format string, a ...interface{}) error {
	c.response.Header().Set(ContentType, TextHTML)
	c.response.WriteHeader(code)
	_, err := fmt.Fprintf(c.response, format, a...)
	return err
}

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request using http.Redirect with status code.
func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.response, c.request, url, code)
}

// Error invokes the registered HTTP error handler. Usually used by middleware.
func (c *Context) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

func (c *Context) reset(r *http.Request, w http.ResponseWriter, e *Echo) {
	c.request = r
	c.response.reset(w)
	c.echo = e
}

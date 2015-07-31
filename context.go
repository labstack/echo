package echo

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"fmt"

	"net/url"

	"golang.org/x/net/websocket"
)

type (
	// Context represents context for the current request. It holds request and
	// response objects, path parameters, data and registered handler.
	Context interface {
		Bind(interface{}) error
		Error(err error)
		Form(string) string
		Get(string) interface{}
		JSON(int, interface{}) error
		NoContent(int) error
		P(int) string
		Param(string) string
		PathParameters() []string
		PathParameterValues() []string
		Render(int, string, interface{}) error
		Request() *http.Request
		Reset(*http.Request, http.ResponseWriter, *Echo)
		Response() *Response
		Socket() *websocket.Conn
		Set(string, interface{})
		SetPathParameters([]string)
		SetPathParameterValue(int, string)
		SetRequest(*http.Request)
		SetSocket(*websocket.Conn)
		String(int, string, ...interface{}) error
	}


	RequestContext struct {
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
func NewContext(req *http.Request, res *Response, e *Echo) *RequestContext {
	return &RequestContext{
		request:  req,
		response: res,
		echo:     e,
		pvalues:  make([]string, *e.maxParam),
		store:    make(store),
	}
}

// Request returns *http.Request.
func (c *RequestContext) Request() *http.Request {
	return c.request
}

// SetRequest sets *http.Request.
func (c *RequestContext) SetRequest(r *http.Request) {
	c.request = r
}

// Response returns *Response.
func (c *RequestContext) Response() *Response {
	return c.response
}

// Socket returns *websocket.Conn.
func (c *RequestContext) Socket() *websocket.Conn {
	return c.socket
}

// SetSocket sets *websocket.Conn.
func (c *RequestContext) SetSocket(sock *websocket.Conn) {
	c.socket = sock
}

// SetPathParameters sets the path parameter names
func (c *RequestContext) SetPathParameters(l []string) {
	c.pnames = l
}

// SetPathParameters sets the path parameter names
func (c *RequestContext) SetPathParameterValue(n int, s string) {
	c.pvalues[n] = s
}

// PathParameters returns all of the path parameter names
func (c *RequestContext) PathParameters() []string {
	return c.pnames
}

// PathParameterValues returns all of the path parameter values
func (c *RequestContext) PathParameterValues() []string {
	return c.pvalues
}

// P returns path parameter by index.
func (c *RequestContext) P(i int) (value string) {
	l := len(c.pnames)
	if i < l {
		value = c.pvalues[i]
	}
	return
}

// Param returns path parameter by name.
func (c *RequestContext) Param(name string) (value string) {
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
func (c *RequestContext) Query(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

// Form returns form parameter by name.
func (c *RequestContext) Form(name string) string {
	return c.request.FormValue(name)
}

// Get retrieves data from the context.
func (c *RequestContext) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *RequestContext) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = val
}

// Bind binds the request body into specified type `i`. The default binder does
// it based on Content-Type header.
func (c *RequestContext) Bind(i interface{}) error {
	return c.echo.binder.Bind(c.request, i)
}

// Render renders a template with data and sends a text/html response with status
// code. Templates can be registered using `Echo.SetRenderer()`.
func (c *RequestContext) Render(code int, name string, data interface{}) error {
	if c.echo.renderer == nil {
		return RendererNotRegistered
	}
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	return c.echo.renderer.Render(c.response, name, data)
}

// HTML formats according to a format specifier and sends HTML response with
// status code.
func (c *RequestContext) HTML(code int, format string, a ...interface{}) (err error) {
	c.response.Header().Set(ContentType, TextHTMLCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = fmt.Fprintf(c.response, format, a...)
	return
}

// String formats according to a format specifier and sends text response with status
// code.
func (c *RequestContext) String(code int, format string, a ...interface{}) (err error) {
	c.response.Header().Set(ContentType, TextPlain)
	c.response.WriteHeader(code)
	_, err = fmt.Fprintf(c.response, format, a...)
	return
}

// JSON sends a JSON response with status code.
func (c *RequestContext) JSON(code int, i interface{}) error {
	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(i)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *RequestContext) JSONP(code int, callback string, i interface{}) (err error) {
	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(callback + "("))
	if err = json.NewEncoder(c.response).Encode(i); err == nil {
		c.response.Write([]byte(");"))
	}
	return
}

// XML sends an XML response with status code.
func (c *RequestContext) XML(code int, i interface{}) error {
	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	c.response.Write([]byte(xml.Header))
	return xml.NewEncoder(c.response).Encode(i)
}

// NoContent sends a response with no body and a status code.
func (c *RequestContext) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request using http.Redirect with status code.
func (c *RequestContext) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return InvalidRedirectCode
	}
	http.Redirect(c.response, c.request, url, code)
	return nil
}

// Error invokes the registered HTTP error handler. Generally used by middleware.
func (c *RequestContext) Error(err error) {
	c.echo.httpErrorHandler(err, c)
}

func (c *RequestContext) Reset(r *http.Request, w http.ResponseWriter, e *Echo) {
	c.request = r
	c.response.reset(w)
	c.query = nil
	c.store = nil
	c.echo = e
}

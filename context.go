package echo

import (
	"encoding/json"
	"net/http"
)

type (
	// Context represents context for the current request. It holds request and
	// response references, path parameters, data and registered handler.
	Context struct {
		Request  *http.Request
		Response *Response
		pnames   []string
		pvalues  []string
		store    store
		echo     *Echo
	}
	store map[string]interface{}
)

func NewContext(req *http.Request, res *Response, e *Echo) *Context {
	return &Context{
		Request:  req,
		Response: res,
		echo:     e,
		pnames:   make([]string, e.maxParam),
		pvalues:  make([]string, e.maxParam),
		store:    make(store),
	}
}

// P returns path parameter by index.
func (c *Context) P(i uint8) (value string) {
	l := uint8(len(c.pnames))
	if i <= l {
		value = c.pvalues[i]
	}
	return
}

// Param returns path parameter by name.
func (c *Context) Param(name string) (value string) {
	l := len(c.pnames)
	for i, n := range c.pnames {
		if n == name && i <= l {
			value = c.pvalues[i]
			break
		}
	}
	return
}

// Bind binds the request body into specified type v. Default binder does it
// based on Content-Type header.
func (c *Context) Bind(v interface{}) *HTTPError {
	return c.echo.binder(c.Request, v)
}

// Render invokes the registered HTML template renderer and sends a text/html
// response with status code.
func (c *Context) Render(code int, name string, data interface{}) *HTTPError {
	if c.echo.renderer == nil {
		return &HTTPError{Error: RendererNotRegistered}
	}
	c.Response.Header().Set(ContentType, TextHTML+"; charset=utf-8")
	c.Response.WriteHeader(code)
	return c.echo.renderer.Render(c.Response, name, data)
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(code int, v interface{}) *HTTPError {
	c.Response.Header().Set(ContentType, ApplicationJSON+"; charset=utf-8")
	c.Response.WriteHeader(code)
	if err := json.NewEncoder(c.Response).Encode(v); err != nil {
		return &HTTPError{Error: err}
	}
	return nil
}

// String sends a text/plain response with status code.
func (c *Context) String(code int, s string) *HTTPError {
	c.Response.Header().Set(ContentType, TextPlain+"; charset=utf-8")
	c.Response.WriteHeader(code)
	if _, err := c.Response.Write([]byte(s)); err != nil {
		return &HTTPError{Error: err}
	}
	return nil
}

// HTML sends a text/html response with status code.
func (c *Context) HTML(code int, html string) *HTTPError {
	c.Response.Header().Set(ContentType, TextHTML+"; charset=utf-8")
	c.Response.WriteHeader(code)
	if _, err := c.Response.Write([]byte(html)); err != nil {
		return &HTTPError{Error: err}
	}
	return nil
}

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) *HTTPError {
	c.Response.WriteHeader(code)
	return nil
}

// Error invokes the registered HTTP error handler.
func (c *Context) Error(he *HTTPError) {
	c.echo.httpErrorHandler(he, c)
}

// Get retrieves data from the context.
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *Context) Set(key string, val interface{}) {
	c.store[key] = val
}

// Redirect redirects the request using http.Redirect with status code.
func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.Response, c.Request, url, code)
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request, e *Echo) {
	c.Request = r
	c.Response.reset(w)
	c.echo = e
}

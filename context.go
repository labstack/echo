package echo

import (
	"encoding/json"
	"net/http"
	"strings"
)

type (
	// Context represents context for the current request. It holds request and
	// response references, path parameters, data and registered handler for
	// the route.
	Context struct {
		Request  *http.Request
		Response *response
		params   Params
		store    map[string]interface{}
		echo     *Echo
	}
	store map[string]interface{}
)

// P returns path parameter by index.
func (c *Context) P(i uint8) string {
	return c.params[i].Value
}

// Param returns path parameter by name.
func (c *Context) Param(n string) string {
	return c.params.Get(n)
}

// Bind decodes the payload into provided type based on Content-Type header.
func (c *Context) Bind(i interface{}) bool {
	var err error
	ct := c.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(ct, MIMEJSON) {
		dec := json.NewDecoder(c.Request.Body)
		err = dec.Decode(i)
	} else {
		// TODO:
	}
	if err != nil {
		c.echo.internalServerErrorHandler(c)
		return false
	}
	return true
}

// String sends a text/plain response with status code.
func (c *Context) String(n int, s string) {
	c.Response.Header().Set(HeaderContentType, MIMEText+"; charset=utf-8")
	c.Response.WriteHeader(n)
	c.Response.Write([]byte(s))
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(n int, i interface{}) {
	enc := json.NewEncoder(c.Response)
	c.Response.Header().Set(HeaderContentType, MIMEJSON+"; charset=utf-8")
	c.Response.WriteHeader(n)
	if err := enc.Encode(i); err != nil {
		c.echo.internalServerErrorHandler(c)
	}
}

// func (c *Context) File(n int, file, name string) {
// }

// Get retrieves data from the context.
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *Context) Set(key string, val interface{}) {
	c.store[key] = val
}

// Redirect redirects the request using http.Redirect with status code.
func (c *Context) Redirect(n int, url string) {
	http.Redirect(c.Response, c.Request, url, n)
}

func (c *Context) reset(rw http.ResponseWriter, r *http.Request) {
	c.Response.reset(rw)
	c.Request = r
}

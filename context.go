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
func (c *Context) Bind(i interface{}) (err error) {
	ct := c.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(ct, MIMEJSON) {
		dec := json.NewDecoder(c.Request.Body)
		if err = dec.Decode(i); err != nil {
			err = ErrBindJSON
		}
	} else if strings.HasPrefix(ct, MIMEForm) {
	} else {
		err = ErrUnsupportedContentType
	}
	return
}

// String sends a text/plain response with status code.
func (c *Context) String(n int, s string) {
	c.Response.Header().Set(HeaderContentType, MIMEText+"; charset=utf-8")
	c.Response.WriteHeader(n)
	c.Response.Write([]byte(s))
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(n int, i interface{}) (err error) {
	enc := json.NewEncoder(c.Response)
	c.Response.Header().Set(HeaderContentType, MIMEJSON+"; charset=utf-8")
	c.Response.WriteHeader(n)
	if err := enc.Encode(i); err != nil {
		err = ErrRenderJSON
	}
	return
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

func (c *Context) reset(rw http.ResponseWriter, r *http.Request, e *Echo) {
	c.Response.reset(rw)
	c.Request = r
	c.echo = e
}

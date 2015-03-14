package bolt

import (
	"encoding/json"
	"net/http"
	"strings"
)

type (
	// Context represents context for the current request. It holds request and
	// response references, path parameters, data, registered handlers for
	// the route. Context also drives the execution of middleware.
	Context struct {
		Request  *http.Request
		Response *response
		params   Params
		handlers []HandlerFunc
		store    map[string]interface{}
		l        int // Handlers' length
		i        int // Current handler index
		bolt     *Bolt
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
		c.bolt.internalServerErrorHandler(c)
		return false
	}
	return true
}

// JSON writes status and JSON to the response.
func (c *Context) JSON(n int, i interface{}) {
	enc := json.NewEncoder(c.Response)
	c.Response.Header().Set(HeaderContentType, MIMEJSON+"; charset=utf-8")
	c.Response.WriteHeader(n)
	if err := enc.Encode(i); err != nil {
		c.bolt.internalServerErrorHandler(c)
	}
}

// func (c *Context) File(n int, file, name string) {
// }

// Next executes the next handler in the chain.
func (c *Context) Next() {
	c.i++
	if c.i < c.l {
		c.handlers[c.i](c)
	}
}

// Get retrieves data from the context.
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *Context) Set(key string, val interface{}) {
	c.store[key] = val
}

// Redirect redirects the request using http.Redirect.
func (c *Context) Redirect(n int, url string) {
	http.Redirect(c.Response, c.Request, url, n)
}

func (c *Context) reset(rw http.ResponseWriter, r *http.Request) {
	c.Response.reset(rw)
	c.Request = r
	c.i = -1
}

// Halt halts the current request.
func (c *Context) Halt() {
	c.i = c.l
}

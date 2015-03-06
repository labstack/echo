package bolt

import (
	"encoding/json"
	"net/http"
	"strings"
)

type (
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

func (c *Context) P(i uint8) string {
	return c.params[i].Value
}

func (c *Context) Param(n string) string {
	return c.params.Get(n)
}

//**********
//   Bind
//**********
func (c *Context) Bind(i interface{}) bool {
	var err error
	ct := c.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(ct, MIME_JSON) {
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

//************
//   Render
//************
func (c *Context) JSON(n int, i interface{}) {
	enc := json.NewEncoder(c.Response)
	c.Response.Header().Set(HeaderContentType, MIME_JSON+"; charset=utf-8")
	c.Response.WriteHeader(n)
	if err := enc.Encode(i); err != nil {
		c.bolt.internalServerErrorHandler(c)
	}
}

// Next executes the next handler in the chain.
func (c *Context) Next() {
	c.i++
	if c.i < c.l {
		c.handlers[c.i](c)
	}
}

func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

func (c *Context) Set(key string, val interface{}) {
	c.store[key] = val
}

func (c *Context) Redirect(n int, url string) {
	http.Redirect(c.Response, c.Request, url, n)
}

func (c *Context) reset(rw http.ResponseWriter, r *http.Request) {
	c.Response.reset(rw)
	c.Request = r
	c.i = -1
}

func (c *Context) Halt() {
	c.i = c.l
}

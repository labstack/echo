package bolt

import (
	"encoding/json"
	"net/http"
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
func (c *Context) BindJSON(i interface{}) {
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(i); err != nil {
		c.bolt.internalServerErrorHandler(c)
	}
}

//************
//   Render
//************
func (c *Context) RenderJSON(n int, i interface{}) {
	enc := json.NewEncoder(c.Response)
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

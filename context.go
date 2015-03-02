package bolt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type (
	Context struct {
		Request  *http.Request
		Writer   *response
		Response *http.Response
		params   Params
		handlers []HandlerFunc
		store    map[string]interface{}
		l        int // Handlers' length
		i        int // Current handler index
	}
	store map[string]interface{}
)

func (c *Context) P(i uint8) string {
	return c.params[i].Value
}

func (c *Context) Param(n string) string {
	return c.params.Get(n)
}

func (c *Context) Bind(f Format, i interface{}) (err error) {
	switch f {
	case FmtJSON:
		dec := json.NewDecoder(c.Request.Body)
		if err = dec.Decode(i); err != nil {
			log.Printf("bolt: %s", err)
		}
	}
	return
}

// TODO: return error, streaming?
func (c *Context) Render(n int, f Format, i interface{}) (err error) {
	var body []byte
	switch f {
	case FmtJSON:
		body, err = json.Marshal(i)
	}
	if err != nil {
		return fmt.Errorf("bolt: %s", err)
	}
	// c.Writer.Header().Set(HEADER_CONTENT_TYPE, MIME_JSON)
	c.Writer.WriteHeader(n)
	_, err = c.Writer.Write(body)
	return
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

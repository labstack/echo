package bolt

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type (
	Context struct {
		Transport transport
		Request   *http.Request
		Writer    *response
		Response  *http.Response
		Socket    *Socket
		params    Params
		handlers  []HandlerFunc
		store     map[string]interface{}
		l         int // Handlers' length
		i         int // Current handler index
		client    bool
	}
	store map[string]interface{}
)

func (c *Context) P(i uint8) string {
	return c.params[i].Value
}

func (c *Context) Param(n string) string {
	return c.params.Get(n)
}

func (c *Context) Bind(f format, i interface{}) (err error) {
	var bd io.ReadCloser
	switch c.Transport {
	case TrnspHTTP:
		bd = c.Request.Body
	case TrnspWS, TrnspTCP:
		bd = c.Socket.Body
	}
	switch f {
	case FmtJSON:
		dec := json.NewDecoder(bd)
		if err = dec.Decode(i); err != nil {
			log.Printf("bolt: %s", err)
		}
	}
	return
}

func (c *Context) Decode(i interface{}) (err error) {
	var rd io.Reader

	switch c.Transport {
	case TrnspHTTP:
		rd = c.Request.Body
		if c.client {
			rd = c.Response.Body
			defer rd.(io.Closer).Close()
		}
	case TrnspWS, TrnspTCP:
		var cl int64
		cl, err = strconv.ParseInt(c.Request.Header.Get(HdrContentLength), 10, 64)
		if err != nil {
			return
		}
		rd = io.LimitReader(c.Socket.Reader, cl)
	}

	t := c.Request.Header.Get("Content-Type")
	if c.client {
		t = c.Request.Header.Get("Accept")
	}

	switch t {
	case MIME_MP:
	default: // JSON
		dec := json.NewDecoder(rd)
		if err = dec.Decode(i); err != nil {
			log.Println(err)
		}
	}

	return
}

// TODO: Streaming?
func (c *Context) Encode(f format, i interface{}) (b []byte, err error) {
	switch f {
	case FmtJSON:
		b, err = json.Marshal(i)
	}
	return
}

// TODO: return error
func (c *Context) Render(n int, f format, i interface{}) {
	bd, err := c.Encode(f, i)
	if err != nil {
		log.Printf("bolt: %s", err)
	}

	switch c.Transport {
	default:
		// c.Writer.Header().Set(HEADER_CONTENT_TYPE, MIME_JSON)
		// c.Writer.WriteHeader(int(n))
		// c.Writer.Write(body)
	case TrnspWS, TrnspTCP:
		binary.Write(c.Socket.Writer, binary.BigEndian, uint16(n))      // Status code
		binary.Write(c.Socket.Writer, binary.BigEndian, int64(len(bd))) // Body length
		c.Socket.Writer.Write(bd)                                       // Body
		c.Socket.Writer.Flush()
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

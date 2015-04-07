package echo

import (
	"encoding/json"
	"net/http"
	"strings"
)

type (
	// Context represents context for the current request. It holds request and
	// response references, path parameters, data and registered handler.
	Context struct {
		Request  *http.Request
		Response *response
		params   Params
		store    store
		echo     *Echo
	}
	store map[string]interface{}
)

// P returns path parameter by index.
func (c *Context) P(i uint8) string {
	return c.params[i].Value
}

// Param returns path parameter by name.
func (c *Context) Param(name string) (value string) {
	for _, p := range c.params {
		if p.Name == name {
			value = p.Value
		}
	}
	return
}

// Bind decodes the body into provided type based on Content-Type header.
func (c *Context) Bind(v interface{}) error {
	ct := c.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(ct, MIMEJSON) {
		return json.NewDecoder(c.Request.Body).Decode(v)
	} else if strings.HasPrefix(ct, MIMEForm) {
		return nil
	}
	return ErrUnsupportedMediaType
}

func (c *Context) Render(code int, name string, data interface{}) error {
	c.Response.Header().Set(HeaderContentType, MIMEHTML+"; charset=utf-8")
	c.Response.WriteHeader(code)
	return c.echo.renderFunc(c.Response.ResponseWriter, name, data)
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(code int, v interface{}) error {
	c.Response.Header().Set(HeaderContentType, MIMEJSON+"; charset=utf-8")
	c.Response.WriteHeader(code)
	return json.NewEncoder(c.Response).Encode(v)
}

// String sends a text/plain response with status code.
func (c *Context) String(code int, s string) (err error) {
	c.Response.Header().Set(HeaderContentType, MIMEText+"; charset=utf-8")
	c.Response.WriteHeader(code)
	_, err = c.Response.Write([]byte(s))
	return
}

// HTML sends a text/html response with status code.
func (c *Context) HTML(code int, html string) (err error) {
	c.Response.Header().Set(HeaderContentType, MIMEHTML+"; charset=utf-8")
	c.Response.WriteHeader(code)
	_, err = c.Response.Write([]byte(html))
	return
}

// func (c *Context) File(code int, file, name string) {
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
func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.Response, c.Request, url, code)
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request, e *Echo) {
	c.Response.reset(w)
	c.Request = r
	c.echo = e
}

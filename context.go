package echo

import (
	"encoding/json"
	"html/template"
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
func (c *Context) Bind(i interface{}) error {
	ct := c.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(ct, MIMEJSON) {
		return json.NewDecoder(c.Request.Body).Decode(i)
	} else if strings.HasPrefix(ct, MIMEForm) {
		return nil
	}
	return ErrUnsupportedMediaType
}

// Render encodes the provided type and sends a response with status code
// based on Accept header. If Accept header not set, it defaults to html/plain.
func (c *Context) Render(code int, i interface{}) error {
	a := c.Request.Header.Get(HeaderAccept)
	if strings.HasPrefix(a, MIMEJSON) {
		return c.JSON(code, i)
	} else if strings.HasPrefix(a, MIMEText) {
		return c.String(code, i.(string))
	} else if strings.HasPrefix(a, MIMEHTML) {
	}
	return c.HTMLString(code, i.(string))
}

// JSON sends an application/json response with status code.
func (c *Context) JSON(code int, i interface{}) error {
	c.Response.Header().Set(HeaderContentType, MIMEJSON+"; charset=utf-8")
	c.Response.WriteHeader(code)
	return json.NewEncoder(c.Response).Encode(i)
}

// String sends a text/plain response with status code.
func (c *Context) String(code int, s string) (err error) {
	c.Response.Header().Set(HeaderContentType, MIMEText+"; charset=utf-8")
	c.Response.WriteHeader(code)
	_, err = c.Response.Write([]byte(s))
	return
}

// HTMLString sends a text/html response with status code.
func (c *Context) HTMLString(code int, html string) (err error) {
	c.Response.Header().Set(HeaderContentType, MIMEHTML+"; charset=utf-8")
	c.Response.WriteHeader(code)
	_, err = c.Response.Write([]byte(html))
	return
}

// HTML applies the template associated with t that has the given name to
// the specified data object and sends a text/html response with status code.
func (c *Context) HTML(code int, t *template.Template, name string, data interface{}) (err error) {
	return t.ExecuteTemplate(c.Response, name, data)
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

func (c *Context) reset(rw http.ResponseWriter, r *http.Request, e *Echo) {
	c.Response.reset(rw)
	c.Request = r
	c.echo = e
}

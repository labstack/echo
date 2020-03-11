package fasthttp

import (
	"time"

	"github.com/valyala/fasthttp"
)

type (
	// Cookie implements `engine.Cookie`.
	Cookie struct {
		*fasthttp.Cookie
	}
)

// Name implements `engine.Cookie#Name` function.
func (c *Cookie) Name() string {
	return string(c.Cookie.Key())
}

// Value implements `engine.Cookie#Value` function.
func (c *Cookie) Value() string {
	return string(c.Cookie.Value())
}

// Path implements `engine.Cookie#Path` function.
func (c *Cookie) Path() string {
	return string(c.Cookie.Path())
}

// Domain implements `engine.Cookie#Domain` function.
func (c *Cookie) Domain() string {
	return string(c.Cookie.Domain())
}

// Expires implements `engine.Cookie#Expires` function.
func (c *Cookie) Expires() time.Time {
	return c.Cookie.Expire()
}

// Secure implements `engine.Cookie#Secure` function.
func (c *Cookie) Secure() bool {
	return c.Cookie.Secure()
}

// HTTPOnly implements `engine.Cookie#HTTPOnly` function.
func (c *Cookie) HTTPOnly() bool {
	return c.Cookie.HTTPOnly()
}

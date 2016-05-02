package standard

import (
	"net/http"
	"time"
)

type (
	// Cookie implements `engine.Cookie`.
	Cookie struct {
		*http.Cookie
	}
)

// Name implements `engine.Cookie#Name` function.
func (c *Cookie) Name() string {
	return c.Cookie.Name
}

// Value implements `engine.Cookie#Value` function.
func (c *Cookie) Value() string {
	return c.Cookie.Value
}

// Path implements `engine.Cookie#Path` function.
func (c *Cookie) Path() string {
	return c.Cookie.Path
}

// Domain implements `engine.Cookie#Domain` function.
func (c *Cookie) Domain() string {
	return c.Cookie.Domain
}

// Expires implements `engine.Cookie#Expires` function.
func (c *Cookie) Expires() time.Time {
	return c.Cookie.Expires
}

// Secure implements `engine.Cookie#Secure` function.
func (c *Cookie) Secure() bool {
	return c.Cookie.Secure
}

// HTTPOnly implements `engine.Cookie#HTTPOnly` function.
func (c *Cookie) HTTPOnly() bool {
	return c.Cookie.HttpOnly
}

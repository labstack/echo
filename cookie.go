package echo

import "time"

type (
	// Cookie defines the HTTP cookie.
	Cookie struct {
		name     string
		value    string
		path     string
		domain   string
		expires  time.Time
		secure   bool
		httpOnly bool
	}
)

// Name returns the cookie name.
func (c *Cookie) Name() string {
	return c.name
}

// SetName sets cookie name.
func (c *Cookie) SetName(name string) {
	c.name = name
}

// Value returns the cookie value.
func (c *Cookie) Value() string {
	return c.value
}

// SetValue sets the cookie value.
func (c *Cookie) SetValue(value string) {
	c.value = value
}

// Path returns the cookie path.
func (c *Cookie) Path() string {
	return c.path
}

// SetPath sets the cookie path.
func (c *Cookie) SetPath(path string) {
	c.path = path
}

// Domain returns the cookie domain.
func (c *Cookie) Domain() string {
	return c.domain
}

// SetDomain sets the cookie domain.
func (c *Cookie) SetDomain(domain string) {
	c.domain = domain
}

// Expires returns the cookie expiry time.
func (c *Cookie) Expires() time.Time {
	return c.expires
}

// SetExpires sets the cookie expiry time.
func (c *Cookie) SetExpires(expires time.Time) {
	c.expires = expires
}

// Secure indicates if cookie is Secure.
func (c *Cookie) Secure() bool {
	return c.secure
}

// SetSecure sets the cookie as Secure.
func (c *Cookie) SetSecure(secure bool) {
	c.secure = secure
}

// HTTPOnly indicates if cookie is HTTPOnly.
func (c *Cookie) HTTPOnly() bool {
	return c.httpOnly
}

// SetHTTPOnly sets the cookie as HTTPOnly.
func (c *Cookie) SetHTTPOnly(httpOnly bool) {
	c.httpOnly = httpOnly
}

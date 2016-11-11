package echo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCookie(t *testing.T) {
	c := new(Cookie)

	// Name
	c.SetName("name")
	assert.Equal(t, "name", c.Name())

	// Value
	c.SetValue("Jon Snow")
	assert.Equal(t, "Jon Snow", c.Value())

	// Path
	c.SetPath("/")
	assert.Equal(t, "/", c.Path())

	// Domain
	c.SetDomain("labstack.com")
	assert.Equal(t, "labstack.com", c.Domain())

	// Expires
	now := time.Now()
	c.SetExpires(now)
	assert.Equal(t, now, c.Expires())

	// Secure
	c.SetSecure(true)
	assert.Equal(t, true, c.Secure())

	// HTTPOnly
	c.SetHTTPOnly(true)
	assert.Equal(t, true, c.HTTPOnly())
}

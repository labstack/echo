package standard

import (
	"github.com/labstack/echo/engine/test"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestURL(t *testing.T) {
	u, _ := url.Parse("https://github.com/labstack/echo?param1=value1&param1=value2&param2=value3")
	mUrl := &URL{u, nil}
	test.URLTest(t, mUrl)

	mUrl.reset(&url.URL{})
	assert.Equal(t, "", mUrl.Host)
}

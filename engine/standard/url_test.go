package standard

import (
	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo/engine/test"
	"net/url"
	"testing"
)

func TestURL(t *testing.T) {
	u, _ := url.Parse("https://github.com/trafficstars/echo?param1=value1&param1=value2&param2=value3")
	mUrl := &URL{u, nil}
	test.URLTest(t, mUrl)

	mUrl.reset(&url.URL{})
	assert.Equal(t, "", mUrl.Host)
}

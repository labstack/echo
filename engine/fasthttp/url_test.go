package fasthttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo/engine/test"
	fast "github.com/valyala/fasthttp"
)

func TestURL(t *testing.T) {
	uri := &fast.URI{}
	uri.Parse([]byte("github.com"), []byte("/trafficstars/echo?param1=value1&param1=value2&param2=value3"))
	mUrl := &URL{uri}
	test.URLTest(t, mUrl)

	mUrl.reset(&fast.URI{})
	assert.Equal(t, "", string(mUrl.Host()))
}

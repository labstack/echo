package fasthttp

import (
	"github.com/labstack/echo/engine/test"
	"github.com/stretchr/testify/assert"
	fast "github.com/valyala/fasthttp"
	"testing"
)

func TestURL(t *testing.T) {
	uri := &fast.URI{}
	uri.Parse([]byte("github.com"), []byte("/labstack/echo?param1=value1&param1=value2&param2=value3"))
	mUrl := &URL{uri}
	test.URLTest(t, mUrl)

	mUrl.reset(&fast.URI{})
	assert.Equal(t, "", string(mUrl.Host()))
}

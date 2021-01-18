package fasthttp

import (
	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo/engine/test"
	fast "github.com/valyala/fasthttp"
	"testing"
)

func TestRequestHeader(t *testing.T) {
	header := &RequestHeader{&fast.RequestHeader{}}
	test.HeaderTest(t, header)

	header.reset(&fast.RequestHeader{})
	assert.Len(t, header.Keys(), 0)
}

func TestResponseHeader(t *testing.T) {
	header := &ResponseHeader{&fast.ResponseHeader{}}
	test.HeaderTest(t, header)

	header.reset(&fast.ResponseHeader{})
	assert.Len(t, header.Keys(), 1)
}

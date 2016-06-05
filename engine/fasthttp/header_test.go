package fasthttp

import (
	"github.com/labstack/echo/engine/test"
	"github.com/stretchr/testify/assert"
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

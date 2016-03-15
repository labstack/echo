// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	// RequestHeader holds `fasthttp.RequestHeader`.
	RequestHeader struct {
		*fasthttp.RequestHeader
	}

	// ResponseHeader holds `fasthttp.ResponseHeader`.
	ResponseHeader struct {
		*fasthttp.ResponseHeader
	}
)

// Add implements `engine.Header#Add` method.
func (h *RequestHeader) Add(key, val string) {
	// h.RequestHeader.Add(key, val)
}

// Del implements `engine.Header#Del` method.
func (h *RequestHeader) Del(key string) {
	h.RequestHeader.Del(key)
}

// Set implements `engine.Header#Set` method.
func (h *RequestHeader) Set(key, val string) {
	h.RequestHeader.Set(key, val)
}

// Get implements `engine.Header#Get` method.
func (h *RequestHeader) Get(key string) string {
	return string(h.Peek(key))
}

func (h *RequestHeader) reset(hdr *fasthttp.RequestHeader) {
	h.RequestHeader = hdr
}

// Add implements `engine.Header#Add` method.
func (h *ResponseHeader) Add(key, val string) {
	// TODO: https://github.com/valyala/fasthttp/issues/69
	// h.header.Add(key, val)
}

// Del implements `engine.Header#Del` method.
func (h *ResponseHeader) Del(key string) {
	h.ResponseHeader.Del(key)
}

// Get implements `engine.Header#Get` method.
func (h *ResponseHeader) Get(key string) string {
	return string(h.Peek(key))
}

// Set implements `engine.Header#Set` method.
func (h *ResponseHeader) Set(key, val string) {
	h.ResponseHeader.Set(key, val)
}

func (h *ResponseHeader) reset(hdr *fasthttp.ResponseHeader) {
	h.ResponseHeader = hdr
}

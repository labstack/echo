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

// Add implements `engine.Header#Add` function.
func (h *RequestHeader) Add(key, val string) {
	h.RequestHeader.Add(key, val)
}

// Del implements `engine.Header#Del` function.
func (h *RequestHeader) Del(key string) {
	h.RequestHeader.Del(key)
}

// Set implements `engine.Header#Set` function.
func (h *RequestHeader) Set(key, val string) {
	h.RequestHeader.Set(key, val)
}

// Get implements `engine.Header#Get` function.
func (h *RequestHeader) Get(key string) string {
	return string(h.Peek(key))
}

// Keys implements `engine.Header#Keys` function.
func (h *RequestHeader) Keys() (keys []string) {
	keys = make([]string, h.Len())
	i := 0
	h.VisitAll(func(k, v []byte) {
		keys[i] = string(k)
		i++
	})
	return
}

// Contains implements `engine.Header#Contains` function.
func (h *RequestHeader) Contains(key string) bool {
	return h.Peek(key) != nil
}

func (h *RequestHeader) reset(hdr *fasthttp.RequestHeader) {
	h.RequestHeader = hdr
}

// Add implements `engine.Header#Add` function.
func (h *ResponseHeader) Add(key, val string) {
	h.ResponseHeader.Add(key, val)
}

// Del implements `engine.Header#Del` function.
func (h *ResponseHeader) Del(key string) {
	h.ResponseHeader.Del(key)
}

// Get implements `engine.Header#Get` function.
func (h *ResponseHeader) Get(key string) string {
	return string(h.Peek(key))
}

// Set implements `engine.Header#Set` function.
func (h *ResponseHeader) Set(key, val string) {
	h.ResponseHeader.Set(key, val)
}

// Keys implements `engine.Header#Keys` function.
func (h *ResponseHeader) Keys() (keys []string) {
	keys = make([]string, h.Len())
	i := 0
	h.VisitAll(func(k, v []byte) {
		keys[i] = string(k)
		i++
	})
	return
}

// Contains implements `engine.Header#Contains` function.
func (h *ResponseHeader) Contains(key string) bool {
	return h.Peek(key) != nil
}

func (h *ResponseHeader) reset(hdr *fasthttp.ResponseHeader) {
	h.ResponseHeader = hdr
}

// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	RequestHeader struct {
		fasthttp.RequestHeader
	}

	ResponseHeader struct {
		fasthttp.ResponseHeader
	}
)

func (h *RequestHeader) Add(key, val string) {
	// h.RequestHeader.Add(key, val)
}

func (h *RequestHeader) Del(key string) {
	h.RequestHeader.Del(key)
}

func (h *RequestHeader) Get(key string) string {
	return string(h.Peek(key))
}

func (h *RequestHeader) Set(key, val string) {
	h.RequestHeader.Set(key, val)
}

func (h *RequestHeader) reset(hdr fasthttp.RequestHeader) {
	h.RequestHeader = hdr
}

func (h *ResponseHeader) Add(key, val string) {
	// h.header.Add(key, val)
}

func (h *ResponseHeader) Del(key string) {
	h.ResponseHeader.Del(key)
}

func (h *ResponseHeader) Get(key string) string {
	return string(h.Peek(key))
}

func (h *ResponseHeader) Set(key, val string) {
	h.ResponseHeader.Set(key, val)
}

func (h *ResponseHeader) reset(hdr fasthttp.ResponseHeader) {
	h.ResponseHeader = hdr
}

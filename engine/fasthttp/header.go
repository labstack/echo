package fasthttp

import "github.com/valyala/fasthttp"

type (
	RequestHeader struct {
		header fasthttp.RequestHeader
	}

	ResponseHeader struct {
		header fasthttp.ResponseHeader
	}
)

func (h *RequestHeader) Add(key, val string) {
	// h.RequestHeader.Add(key, val)
}

func (h *RequestHeader) Del(key string) {
	h.header.Del(key)
}

func (h *RequestHeader) Get(key string) string {
	// return h.header.Peek(key)
	return ""
}

func (h *RequestHeader) Set(key, val string) {
	h.header.Set(key, val)
}

func (h *ResponseHeader) Add(key, val string) {
	// h.header.Add(key, val)
}

func (h *ResponseHeader) Del(key string) {
	h.header.Del(key)
}

func (h *ResponseHeader) Get(key string) string {
	// return h.ResponseHeader.Get(key)
	return ""
}

func (h *ResponseHeader) Set(key, val string) {
	h.header.Set(key, val)
}

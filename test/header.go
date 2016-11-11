package test

import "net/http"

type (
	Header struct {
		header http.Header
	}
)

func (h *Header) Add(key, val string) {
	h.header.Add(key, val)
}

func (h *Header) Del(key string) {
	h.header.Del(key)
}

func (h *Header) Get(key string) string {
	return h.header.Get(key)
}

func (h *Header) Set(key, val string) {
	h.header.Set(key, val)
}

func (h *Header) Keys() (keys []string) {
	keys = make([]string, len(h.header))
	i := 0
	for k := range h.header {
		keys[i] = k
		i++
	}
	return
}

func (h *Header) Contains(key string) bool {
	_, ok := h.header[key]
	return ok
}

func (h *Header) reset(hdr http.Header) {
	h.header = hdr
}

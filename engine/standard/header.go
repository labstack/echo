package standard

import "net/http"

type (
	Header struct {
		http.Header
	}
)

func (h *Header) Add(key, val string) {
	h.Header.Add(key, val)
}

func (h *Header) Del(key string) {
	h.Header.Del(key)
}

func (h *Header) Get(key string) string {
	return h.Header.Get(key)
}

func (h *Header) Set(key, val string) {
	h.Header.Set(key, val)
}

func (h *Header) reset(hdr http.Header) {
	h.Header = hdr
}

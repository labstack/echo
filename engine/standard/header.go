package standard

import "net/http"

type (
	// Header implements `engine.Header`.
	Header struct {
		http.Header
	}
)

// Add implements `engine.Header#Add` function.
func (h *Header) Add(key, val string) {
	h.Header.Add(key, val)
}

// Del implements `engine.Header#Del` function.
func (h *Header) Del(key string) {
	h.Header.Del(key)
}

// Set implements `engine.Header#Set` function.
func (h *Header) Set(key, val string) {
	h.Header.Set(key, val)
}

// Get implements `engine.Header#Get` function.
func (h *Header) Get(key string) string {
	return h.Header.Get(key)
}

// Keys implements `engine.Header#Keys` function.
func (h *Header) Keys() (keys []string) {
	keys = make([]string, len(h.Header))
	i := 0
	for k := range h.Header {
		keys[i] = k
		i++
	}
	return
}

func (h *Header) reset(hdr http.Header) {
	h.Header = hdr
}

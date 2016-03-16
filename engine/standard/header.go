package standard

import "net/http"

type (
	// Header implements `engine.Header`.
	Header struct {
		http.Header
	}
)

// Add implements `engine.Header#Add` method.
func (h *Header) Add(key, val string) {
	h.Header.Add(key, val)
}

// Del implements `engine.Header#Del` method.
func (h *Header) Del(key string) {
	h.Header.Del(key)
}

// Set implements `engine.Header#Set` method.
func (h *Header) Set(key, val string) {
	h.Header.Set(key, val)
}

// Get implements `engine.Header#Get` method.
func (h *Header) Get(key string) string {
	return h.Header.Get(key)
}

func (h *Header) All() []string {
	headers := make([]string, len(h.Header))
	idx := 0
	for header, _ := range h.Header {
		headers[idx] = header
		idx++
	}
	return headers
}

func (h *Header) reset(hdr http.Header) {
	h.Header = hdr
}

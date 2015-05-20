package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse(t *testing.T) {
	r := NewResponse(httptest.NewRecorder())

	// Header
	if r.Header() == nil {
		t.Error("header should not be nil")
	}

	// WriteHeader
	r.WriteHeader(http.StatusOK)
	if r.status != http.StatusOK {
		t.Errorf("status should be %d", http.StatusOK)
	}
	if r.committed != true {
		t.Error("response should be true")
	}
	// Response already committed
	r.WriteHeader(http.StatusOK)

	// Status
	r.status = http.StatusOK
	if r.Status() != http.StatusOK {
		t.Errorf("status should be %d", http.StatusOK)
	}

	// Write & Size
	s := "echo"
	r.Write([]byte(s))
	if r.Size() != int64(len(s)) {
		t.Errorf("size should be %d", len(s))
	}
}

package echo

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := NewResponse(w)

	// SetWriter
	r.SetWriter(w)

	// Writer
	assert.Equal(t, w, r.Writer())

	// Header
	assert.NotNil(t, r.Header())

	// WriteHeader
	r.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, r.status)

	// Committed
	assert.True(t, r.committed)

	// Already committed
	r.WriteHeader(http.StatusTeapot)
	assert.NotEqual(t, http.StatusTeapot, r.Status())

	// Status
	r.status = http.StatusOK
	assert.Equal(t, http.StatusOK, r.Status())

	// Write
	s := "echo"
	_, err := r.Write([]byte(s))
	assert.NoError(t, err)

	// Flush
	r.Flush()

	// Size
	assert.EqualValues(t, len(s), r.Size())

	// Hijack
	assert.Panics(t, func() {
		r.Hijack()
	})

	// CloseNotify
	assert.Panics(t, func() {
		r.CloseNotify()
	})

	// reset
	r.reset(httptest.NewRecorder())
}

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	e := New()
	w := httptest.NewRecorder()
	r := NewResponse(w, e)

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

	// Committed
	assert.Equal(t, true, r.Committed())

	// Hijack
	assert.Panics(t, func() {
		r.Hijack()
	})

	// CloseNotify
	assert.Panics(t, func() {
		r.CloseNotify()
	})

	// reset
	r.reset(httptest.NewRecorder(), New())
}

func TestResponseWriteCommits(t *testing.T) {
	w := httptest.NewRecorder()
	r := NewResponse(w)
	r.SetWriter(w)

	// write body, it writes header if not committed yet
	s := "echo"
	r.Write([]byte(s))

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Body.String(), s)

	assert.Equal(t, r.Status(), 200)
	assert.Equal(t, r.Size(), int64(4))
	assert.True(t, r.Committed())

	// this is ignored with warning
	r.WriteHeader(400)

	assert.Equal(t, r.Status(), 200)
	assert.Equal(t, r.Size(), int64(4))
	assert.True(t, r.Committed())
}

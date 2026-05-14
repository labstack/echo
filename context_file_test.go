package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

// mockReadFromWriter implements io.ReaderFrom to trigger the optimization path
type mockReadFromWriter struct {
	*httptest.ResponseRecorder
	readFromCalled bool
}

func (m *mockReadFromWriter) ReadFrom(r io.Reader) (int64, error) {
	m.readFromCalled = true
	// Simulate sendfile/optimized copy
	return io.Copy(m.ResponseRecorder, r)
}

// mockSimpleResponseWriter ONLY implements http.ResponseWriter (no ReadFrom)
// This is used as a control group to force the original non-optimized path.
type mockSimpleResponseWriter struct {
	*httptest.ResponseRecorder
}

const readFromTestFile = "readfrom_test_data.txt"

func TestContext_File_ReadFrom_Optimization(t *testing.T) {
	e := New()
	tmpDir := t.TempDir()
	content := "hello optimization parity check content"
	err := os.WriteFile(filepath.Join(tmpDir, readFromTestFile), []byte(content), 0644)
	assert.NoError(t, err)
	e.Filesystem = os.DirFS(tmpDir)

	t.Run("Verify optimization triggers and parity", func(t *testing.T) {
		// Use e.NewContext and c.File for end-to-end functional parity check.
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// 1. Optimized Path Group
		recOpt := httptest.NewRecorder()
		mwOpt := &mockReadFromWriter{ResponseRecorder: recOpt}
		cOpt := e.NewContext(req, mwOpt)
		assert.NoError(t, cOpt.File(readFromTestFile))
		resOpt := cOpt.Response().(*Response)

		// 2. Original Path Group (Control)
		recOri := httptest.NewRecorder()
		mwOri := &mockSimpleResponseWriter{ResponseRecorder: recOri}
		cOri := e.NewContext(req, mwOri)
		assert.NoError(t, cOri.File(readFromTestFile))
		resOri := cOri.Response().(*Response)

		// ASSERTIONS:
		assert.True(t, mwOpt.readFromCalled, "Optimized path MUST trigger ReadFrom")
		assert.Equal(t, recOri.Code, recOpt.Code, "httptest.Recorder Code parity")
		assert.Equal(t, recOri.Body.String(), recOpt.Body.String(), "Body content parity")

		// Echo Response State Parity
		assert.Equal(t, resOri.Status, resOpt.Status, "Response.Status parity")
		assert.Equal(t, resOri.Size, resOpt.Size, "Response.Size parity")
		assert.Equal(t, resOri.Committed, resOpt.Committed, "Response.Committed parity")
	})

	t.Run("ReadFrom: Custom Status already set", func(t *testing.T) {
		// Manually construct the wrapper to bypass http.ServeContent's side effects
		// and surgically verify the Status/Before-hook bridging logic in ReadFrom.
		rec := httptest.NewRecorder()
		mw := &mockReadFromWriter{ResponseRecorder: rec}
		res := &Response{ResponseWriter: mw}
		w := &responseWithReadFrom{res}

		res.Status = http.StatusCreated
		n, err := w.ReadFrom(strings.NewReader("test data"))
		assert.NoError(t, err)
		assert.Equal(t, int64(9), n)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.True(t, res.Committed)
	})

	t.Run("ReadFrom: Already committed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		mw := &mockReadFromWriter{ResponseRecorder: rec}
		c := e.NewContext(req, mw)

		c.Response().WriteHeader(http.StatusAccepted) // Commit here
		assert.NoError(t, c.File(readFromTestFile))

		assert.True(t, mw.readFromCalled)
		assert.Equal(t, http.StatusAccepted, rec.Code)
		// Body should still be written because ServeContent continues after WriteHeader
		assert.Contains(t, rec.Body.String(), "hello optimization")
	})

	t.Run("ReadFrom: IO Error during Copy", func(t *testing.T) {
		// Directly test the wrapper to verify state updates (Size, Committed)
		// when an error occurs during the transfer.
		errReader := iotest.ErrReader(io.ErrUnexpectedEOF)

		res := &Response{ResponseWriter: &mockReadFromWriter{ResponseRecorder: httptest.NewRecorder()}}
		w := &responseWithReadFrom{res}

		n, err := w.ReadFrom(errReader)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		assert.Equal(t, int64(0), n)
		assert.Equal(t, int64(0), res.Size)
		assert.True(t, res.Committed)
	})

	t.Run("Hook Compatibility: Before hook triggers on ReadFrom", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		mw := &mockReadFromWriter{ResponseRecorder: rec}
		c := e.NewContext(req, mw)

		beforeTriggered := false
		c.Response().(*Response).Before(func() {
			beforeTriggered = true
		})

		assert.NoError(t, c.File(readFromTestFile))
		assert.True(t, mw.readFromCalled)
		assert.True(t, beforeTriggered, "Before hook must be called even on ReadFrom path")
	})

	t.Run("Hook Compatibility: After hook disables ReadFrom", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		mw := &mockReadFromWriter{ResponseRecorder: rec}
		c := e.NewContext(req, mw)

		afterCalls := 0
		c.Response().(*Response).After(func() {
			afterCalls++
		})

		assert.NoError(t, c.File(readFromTestFile))
		assert.False(t, mw.readFromCalled, "ReadFrom must be DISABLED when After hooks exist")
		assert.True(t, afterCalls > 0, "After hooks must be triggered via standard Write path")
	})

	t.Run("Error Parity: 416 Invalid Range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Range", "bytes=100-200")

		rec := httptest.NewRecorder()
		mw := &mockReadFromWriter{ResponseRecorder: rec}
		c := e.NewContext(req, mw)

		assert.NoError(t, c.File(readFromTestFile))
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rec.Code)
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, c.Response().(*Response).Status)
	})
}

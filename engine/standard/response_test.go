package standard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/log"
	"github.com/stretchr/testify/assert"
)

func TestResponseWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	res := NewResponse(rec, log.New(log.Prefix("test")))
	res.WriteHeader(http.StatusOK)
	assert.True(t, res.Committed())
	assert.Equal(t, http.StatusOK, res.Status())
}

func TestResponseWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	res := NewResponse(rec, log.New(log.Prefix("test")))
	res.Write([]byte("test"))
	assert.Equal(t, int64(4), res.Size())
	assert.Equal(t, "test", rec.Body.String())
	res.Flush()
	assert.True(t, rec.Flushed)
}

func TestResponseSetCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	res := NewResponse(rec, log.New(log.Prefix("test")))
	res.SetCookie(&Cookie{&http.Cookie{
		Name:  "name",
		Value: "Jon Snow",
	}})
	assert.Equal(t, "name=Jon Snow", rec.Header().Get("Set-Cookie"))
}

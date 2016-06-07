package standard

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

func TestResponseWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder, log.New("echo"))

	resp.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, resp.Status())

	assert.True(t, resp.Committed())
}

func TestResponseWrite(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder, log.New("echo"))
	resp.Write([]byte("Hello"))
	assert.Equal(t, int64(5), resp.Size())
	if body, err := ioutil.ReadAll(recorder.Body); assert.NoError(t, err) {
		assert.Equal(t, "Hello", string(body))
	}
	resp.Flush()
	assert.True(t, recorder.Flushed)
}

func TestResponseSetCookie(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder, log.New("echo"))

	resp.SetCookie(&Cookie{&http.Cookie{
		Name:     "session",
		Value:    "securetoken",
		Path:     "/",
		Domain:   "github.com",
		Expires:  time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
		Secure:   true,
		HttpOnly: true,
	}})

	assert.Equal(t, "session=securetoken; Path=/; Domain=github.com; Expires=Fri, 01 Jan 2016 00:00:00 GMT; HttpOnly; Secure", recorder.Header().Get("Set-Cookie"))
}

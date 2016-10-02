package xhandler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestTimeoutHandler(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := TimeoutHandler(time.Second)(&handler{})
	h := New(ctx, xh)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value with deadline", w.Body.String())
}

type closeNotifyWriter struct {
	*httptest.ResponseRecorder
	closed bool
}

func (w *closeNotifyWriter) CloseNotify() <-chan bool {
	notify := make(chan bool, 1)
	if w.closed {
		// return an already "closed" notifier
		notify <- true
	}
	return notify
}

func TestCloseHandlerClientClose(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := CloseHandler(&handler{})
	h := New(ctx, xh)
	w := &closeNotifyWriter{ResponseRecorder: httptest.NewRecorder(), closed: true}
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value canceled", w.Body.String())
}

func TestCloseHandlerRequestEnds(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := CloseHandler(&handler{})
	h := New(ctx, xh)
	w := &closeNotifyWriter{ResponseRecorder: httptest.NewRecorder(), closed: false}
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	h.ServeHTTP(w, r)
	assert.Equal(t, "value", w.Body.String())
}

func TestIf(t *testing.T) {
	trueHandler := HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/true", r.URL.Path)
	})
	falseHandler := HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, "/true", r.URL.Path)
	})
	ctx := context.WithValue(context.Background(), contextKey, "value")
	xh := If(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return r.URL.Path == "/true"
		},
		func(next HandlerC) HandlerC {
			return trueHandler
		},
	)(falseHandler)
	h := New(ctx, xh)
	r, _ := http.NewRequest("GET", "http://example.com/true", nil)
	h.ServeHTTP(nil, r)
	r, _ = http.NewRequest("GET", "http://example.com/false", nil)
	h.ServeHTTP(nil, r)
}

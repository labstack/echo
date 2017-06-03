package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type (
	closeNotifyRecorder struct {
		*httptest.ResponseRecorder
		closed chan bool
	}
)

func newCloseNotifyRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func TestProxy(t *testing.T) {
	// Setup
	t1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "target 1")
	}))
	defer t1.Close()
	t2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "target 2")
	}))
	defer t2.Close()
	config := ProxyConfig{
		Targets: []*ProxyTarget{
			&ProxyTarget{
				URL: t1.URL,
			},
			&ProxyTarget{
				URL: t2.URL,
			},
		},
	}

	// Random
	e := echo.New()
	e.Use(Proxy(config))
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := newCloseNotifyRecorder()
	e.ServeHTTP(rec, req)
	body := rec.Body.String()
	targets := map[string]bool{
		"target 1": true,
		"target 2": true,
	}
	assert.Condition(t, func() bool {
		return targets[body]
	})

	// Round-robin
	config.Balance = "round-robin"
	e = echo.New()
	e.Use(Proxy(config))
	rec = newCloseNotifyRecorder()
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, "target 1", body)
	rec = newCloseNotifyRecorder()
	e.ServeHTTP(rec, req)
	body = rec.Body.String()
	assert.Equal(t, "target 2", body)
}

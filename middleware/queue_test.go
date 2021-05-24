package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestQueueWithConfig(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		pts := c.Get("procTime")
		procTime, ok := pts.(int)
		if !ok {
			return c.NoContent(http.StatusInternalServerError)
		}

		time.Sleep(time.Duration(procTime) * time.Millisecond)

		return c.NoContent(http.StatusOK)
	}

	mw := QueueWithConfig(QueueConfig{
		QueueSize:     2,
		Workers:       1,
		QueueTimeout:  200 * time.Millisecond,
		WorkerTimeout: 100 * time.Millisecond,
	})

	testCases := []struct {
		procTime int // in Milliseconds
	}{
		{50},
		{95},
		{95},
		{95},
		{95},
		{120},
		{250},
	}

	ch := make(chan int, len(testCases))
	var wg sync.WaitGroup

	for _, tc := range testCases {

		wg.Add(1)

		go func(pt int) {

			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			c.Set("procTime", pt)

			_ = mw(handler)(c)

			ch <- rec.Code

		}(tc.procTime)

	}

	wg.Wait()

	var errQueueFull, errQueueTimeout, errInternalServerError bool

	for i := 0; i < len(testCases); i++ {
		c := <-ch

		t.Log(c)

		if c == http.StatusTooManyRequests {
			errQueueFull = true
		}

		if c == http.StatusRequestTimeout {
			errQueueTimeout = true
		}

		if c == http.StatusInternalServerError {
			errInternalServerError = true
		}
	}

	assert.Equal(t, true, errQueueFull)
	assert.Equal(t, true, errQueueTimeout)
	assert.Equal(t, false, errInternalServerError)

}

func TestQueueWithConfig_skipper(t *testing.T) {
	e := echo.New()

	var beforeFuncRan bool
	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := QueueWithConfig(QueueConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		BeforeFunc: func(c echo.Context) {
			beforeFuncRan = true
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, false, beforeFuncRan)
}

func TestQueueWithConfig_skipperNoSkip(t *testing.T) {
	e := echo.New()

	var beforeFuncRan bool
	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := QueueWithConfig(QueueConfig{
		Skipper: func(c echo.Context) bool {
			return false
		},
		BeforeFunc: func(c echo.Context) {
			beforeFuncRan = true
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, true, beforeFuncRan)
}

func TestQueueWithConfig_beforeFunc(t *testing.T) {
	e := echo.New()

	var beforeRan bool
	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mw := QueueWithConfig(QueueConfig{
		BeforeFunc: func(c echo.Context) {
			beforeRan = true
		},
	})

	_ = mw(handler)(c)

	assert.Equal(t, true, beforeRan)
}

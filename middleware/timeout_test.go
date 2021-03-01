// +build go1.13

package middleware

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTimeoutSkipper(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Skipper: func(context echo.Context) bool {
			return true
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		assert.NotEqual(t, "*context.timerCtx", reflect.TypeOf(c.Request().Context()).String())
		return nil
	})(c)

	assert.NoError(t, err)
}

func TestTimeoutWithTimeout0(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: 0,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		assert.NotEqual(t, "*context.timerCtx", reflect.TypeOf(c.Request().Context()).String())
		return nil
	})(c)

	assert.NoError(t, err)
}

func TestTimeoutIsCancelable(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: time.Minute,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		assert.EqualValues(t, "*context.timerCtx", reflect.TypeOf(c.Request().Context()).String())
		return nil
	})(c)

	assert.NoError(t, err)
}

func TestTimeoutErrorOutInHandler(t *testing.T) {
	t.Parallel()
	m := Timeout()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		return errors.New("err")
	})(c)

	assert.Error(t, err)
}

func TestTimeoutTimesOutAfterPredefinedTimeoutWithErrorHandler(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: time.Second,
		ErrorHandler: func(err error, e echo.Context) error {
			assert.EqualError(t, err, context.DeadlineExceeded.Error())
			return errors.New("err")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		time.Sleep(time.Minute)
		return nil
	})(c)

	assert.EqualError(t, err, errors.New("err").Error())
}

func TestTimeoutTimesOutAfterPredefinedTimeout(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: time.Second,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		time.Sleep(time.Minute)
		return nil
	})(c)

	assert.EqualError(t, err, context.DeadlineExceeded.Error())
}

func TestTimeoutTestRequestClone(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/uri?query=value", strings.NewReader(url.Values{"form": {"value"}}.Encode()))
	req.AddCookie(&http.Cookie{Name: "cookie", Value: "value"})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	m := TimeoutWithConfig(TimeoutConfig{
		// Timeout has to be defined or the whole flow for timeout middleware will be skipped
		Timeout: time.Second,
	})

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		// Cookie test
		cookie, err := c.Request().Cookie("cookie")
		if assert.NoError(t, err) {
			assert.EqualValues(t, "cookie", cookie.Name)
			assert.EqualValues(t, "value", cookie.Value)
		}

		// Form values
		if assert.NoError(t, c.Request().ParseForm()) {
			assert.EqualValues(t, "value", c.Request().FormValue("form"))
		}

		// Query string
		assert.EqualValues(t, "value", c.Request().URL.Query()["query"][0])
		return nil
	})(c)

	assert.NoError(t, err)

}

func TestTimeoutRecoversPanic(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: 25 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		panic("panic in handler")
	})(c)

	assert.Error(t, err, "panic recovered in timeout middleware: panic in handler")
}

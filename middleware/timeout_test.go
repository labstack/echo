package middleware

import (
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
		Timeout: 1 * time.Nanosecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		time.Sleep(25 * time.Microsecond)
		return errors.New("response from handler")
	})(c)

	// if not skipped we would have not returned error due context timeout logic
	assert.EqualError(t, err, "response from handler")
}

func TestTimeoutWithTimeout0(t *testing.T) {
	t.Parallel()
	m := Timeout()

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

func TestTimeoutErrorOutInHandler(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		// Timeout has to be defined or the whole flow for timeout middleware will be skipped
		Timeout: 50 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusTeapot, "err")
	})(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "{\"message\":\"err\"}\n", rec.Body.String())
}

func TestTimeoutSuccessfulRequest(t *testing.T) {
	t.Parallel()
	m := TimeoutWithConfig(TimeoutConfig{
		// Timeout has to be defined or the whole flow for timeout middleware will be skipped
		Timeout: 50 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]string{"data": "ok"})
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "{\"data\":\"ok\"}\n", rec.Body.String())
}

func TestTimeoutOnTimeoutRouteErrorHandler(t *testing.T) {
	t.Parallel()

	actualErrChan := make(chan error, 1)
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout: 1 * time.Millisecond,
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			actualErrChan <- err
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	stopChan := make(chan struct{}, 0)
	err := m(func(c echo.Context) error {
		<-stopChan
		return errors.New("error in route after timeout")
	})(c)
	stopChan <- struct{}{}
	assert.NoError(t, err)

	actualErr := <-actualErrChan
	assert.EqualError(t, actualErr, "error in route after timeout")
}

func TestTimeoutTestRequestClone(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/uri?query=value", strings.NewReader(url.Values{"form": {"value"}}.Encode()))
	req.AddCookie(&http.Cookie{Name: "cookie", Value: "value"})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	m := TimeoutWithConfig(TimeoutConfig{
		// Timeout has to be defined or the whole flow for timeout middleware will be skipped
		Timeout: 1 * time.Second,
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
	e := echo.New()
	e.Use(Recover()) // recover middleware will handler our panic
	e.Use(TimeoutWithConfig(TimeoutConfig{
		Timeout: 50 * time.Millisecond,
	}))

	e.GET("/", func(c echo.Context) error {
		panic("panic!!!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		e.ServeHTTP(rec, req)
	})
}

func TestTimeoutDataRace(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout:      timeout,
		ErrorMessage: "Timeout! change me",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		time.Sleep(timeout) // timeout and handler execution time difference is close to zero
		return c.String(http.StatusOK, "Hello, World!")
	})(c)

	assert.NoError(t, err)

	if rec.Code == http.StatusServiceUnavailable {
		assert.Equal(t, "Timeout! change me", rec.Body.String())
	} else {
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}
}

func TestTimeoutWithErrorMessage(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout:      timeout,
		ErrorMessage: "Timeout! change me",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	stopChan := make(chan struct{}, 0)
	err := m(func(c echo.Context) error {
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		<-stopChan
		return c.String(http.StatusOK, "Hello, World!")
	})(c)
	stopChan <- struct{}{}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "Timeout! change me", rec.Body.String())
}

func TestTimeoutWithDefaultErrorMessage(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout:      timeout,
		ErrorMessage: "",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	stopChan := make(chan struct{}, 0)
	err := m(func(c echo.Context) error {
		<-stopChan
		return c.String(http.StatusOK, "Hello, World!")
	})(c)
	stopChan <- struct{}{}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, `<html><head><title>Timeout</title></head><body><h1>Timeout</h1></body></html>`, rec.Body.String())
}

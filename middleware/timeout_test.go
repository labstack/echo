package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

	rec.Code = 1 // we want to be sure that even 200 will not be sent
	err := m(func(c echo.Context) error {
		// this error must not be written to the client response. Middlewares upstream of timeout middleware must be able
		// to handle returned error and this can be done only then handler has not yet committed (written status code)
		// the response.
		return echo.NewHTTPError(http.StatusTeapot, "err")
	})(c)

	assert.Error(t, err)
	assert.EqualError(t, err, "code=418, message=err")
	assert.Equal(t, 1, rec.Code)
	assert.Equal(t, "", rec.Body.String())
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

func TestTimeoutCanHandleContextDeadlineOnNextHandler(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := TimeoutWithConfig(TimeoutConfig{
		Timeout:      timeout,
		ErrorMessage: "Timeout! change me",
	})

	handlerFinishedExecution := make(chan bool)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	stopChan := make(chan struct{})
	err := m(func(c echo.Context) error {
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		<-stopChan

		// The Request Context should have a Deadline set by http.TimeoutHandler
		if _, ok := c.Request().Context().Deadline(); !ok {
			assert.Fail(t, "No timeout set on Request Context")
		}
		handlerFinishedExecution <- c.Request().Context().Err() == nil
		return c.String(http.StatusOK, "Hello, World!")
	})(c)
	stopChan <- struct{}{}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "Timeout! change me", rec.Body.String())
	assert.False(t, <-handlerFinishedExecution)
}

func TestTimeoutWithFullEchoStack(t *testing.T) {
	// test timeout with full http server stack running, do see what http.Server.ErrorLog contains
	var testCases = []struct {
		name                 string
		whenPath             string
		expectStatusCode     int
		expectResponse       string
		expectLogContains    []string
		expectLogNotContains []string
	}{
		{
			name:                 "404 - write response in global error handler",
			whenPath:             "/404",
			expectResponse:       "{\"message\":\"Not Found\"}\n",
			expectStatusCode:     http.StatusNotFound,
			expectLogNotContains: []string{"echo:http: superfluous response.WriteHeader call from"},
			expectLogContains:    []string{`"status":404,"error":"code=404, message=Not Found"`},
		},
		{
			name:                 "418 - write response in handler",
			whenPath:             "/",
			expectResponse:       "{\"message\":\"OK\"}\n",
			expectStatusCode:     http.StatusTeapot,
			expectLogNotContains: []string{"echo:http: superfluous response.WriteHeader call from"},
			expectLogContains:    []string{`"status":418,"error":"",`},
		},
		{
			name:             "503 - handler timeouts, write response in timeout middleware",
			whenPath:         "/?delay=50ms",
			expectResponse:   "<html><head><title>Timeout</title></head><body><h1>Timeout</h1></body></html>",
			expectStatusCode: http.StatusServiceUnavailable,
			expectLogNotContains: []string{
				"echo:http: superfluous response.WriteHeader call from",
				"{", // means that logger was not called.
			},
		},
	}

	e := echo.New()

	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)

	// NOTE: timeout middleware is first as it changes Response.Writer and causes data race for logger middleware if it is not first
	// FIXME: I have no idea how to fix this without adding mutexes.
	e.Use(TimeoutWithConfig(TimeoutConfig{
		Timeout: 15 * time.Millisecond,
	}))
	e.Use(Logger())
	e.Use(Recover())

	e.GET("/", func(c echo.Context) error {
		var delay time.Duration
		if err := echo.QueryParamsBinder(c).Duration("delay", &delay).BindError(); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delay)
		}
		return c.JSON(http.StatusTeapot, map[string]string{"message": "OK"})
	})

	server, addr, err := startServer(e)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	defer server.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset() // this is design this can not be run in parallel

			res, err := http.Get(fmt.Sprintf("http://%v%v", addr, tc.whenPath))
			if err != nil {
				assert.NoError(t, err)
				return
			}

			assert.Equal(t, tc.expectStatusCode, res.StatusCode)
			if body, err := ioutil.ReadAll(res.Body); err == nil {
				assert.Equal(t, tc.expectResponse, string(body))
			} else {
				assert.Fail(t, err.Error())
			}

			logged := buf.String()
			for _, subStr := range tc.expectLogContains {
				assert.True(t, strings.Contains(logged, subStr))
			}
			for _, subStr := range tc.expectLogNotContains {
				assert.False(t, strings.Contains(logged, subStr))
			}
		})
	}
}

func startServer(e *echo.Echo) (*http.Server, string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, "", err
	}

	s := http.Server{
		Handler:  e,
		ErrorLog: log.New(e.Logger.Output(), "echo:", 0),
	}

	errCh := make(chan error)
	go func() {
		if err := s.Serve(l); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-time.After(10 * time.Millisecond):
		return &s, l.Addr().String(), nil
	case err := <-errCh:
		return nil, "", err
	}
}

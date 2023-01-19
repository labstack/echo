package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func TestContextTimeoutSkipper(t *testing.T) {
	t.Parallel()
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
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

func TestContextTimeoutWithTimeout0(t *testing.T) {
	t.Parallel()
	m := ContextTimeout()

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

func TestContextTimeoutErrorOutInHandler(t *testing.T) {
	t.Parallel()
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
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

func TestContextTimeoutSuccessfulRequest(t *testing.T) {
	t.Parallel()
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
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

func TestContextTimeoutTestRequestClone(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/uri?query=value", strings.NewReader(url.Values{"form": {"value"}}.Encode()))
	req.AddCookie(&http.Cookie{Name: "cookie", Value: "value"})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
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

func TestContextTimeoutRecoversPanic(t *testing.T) {
	t.Parallel()
	e := echo.New()
	e.Use(Recover()) // recover middleware will handler our panic
	e.Use(ContextTimeoutWithConfig(ContextTimeoutConfig{
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

func TestContextTimeoutDataRace(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
		Timeout: timeout,
		//ErrorMessage: "Timeout! change me",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {

		if err := sleepWithContext(c.Request().Context(), timeout); err != nil {
			return err
		}
		return c.String(http.StatusOK, "Hello, World!")
	})(c)

	assert.NoError(t, err)

	if rec.Code == http.StatusServiceUnavailable {
		assert.Equal(t, "Timeout! change me", rec.Body.String())
	} else {
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}
}

func TestContextTimeoutWithErrorMessage(t *testing.T) {
	t.Parallel()

	timeoutErrorHandler := func(err error, c echo.Context) error {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return &echo.HTTPError{
					Code:    http.StatusServiceUnavailable,
					Message: "Timeout! change me",
				}
			}
			return err
		}
		return nil
	}

	timeout := 1 * time.Millisecond
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
		Timeout:      timeout,
		ErrorHandler: timeoutErrorHandler,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		if err := sleepWithContext(c.Request().Context(), time.Duration(2*time.Millisecond)); err != nil {
			return err
		}
		return c.String(http.StatusOK, "Hello, World!")
	})(c)

	assert.IsType(t, &echo.HTTPError{}, err)
	assert.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, err.(*echo.HTTPError).Code)
	assert.Equal(t, "Timeout! change me", err.(*echo.HTTPError).Message)
}

func TestContextTimeoutWithDefaultErrorMessage(t *testing.T) {
	t.Parallel()

	timeout := 1 * time.Millisecond
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
		Timeout: timeout,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		if err := sleepWithContext(c.Request().Context(), time.Duration(2*time.Millisecond)); err != nil {
			return err
		}
		return c.String(http.StatusOK, "Hello, World!")
	})(c)

	assert.IsType(t, &echo.HTTPError{}, err)
	assert.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, err.(*echo.HTTPError).Code)
	assert.Equal(t, "Service Unavailable", err.(*echo.HTTPError).Message)
}

func TestContextTimeoutCanHandleContextDeadlineOnNextHandler(t *testing.T) {
	t.Parallel()

	timeoutErrorHandler := func(err error, c echo.Context) error {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return &echo.HTTPError{
					Code:    http.StatusServiceUnavailable,
					Message: "Timeout! change me",
				}
			}
			return err
		}
		return nil
	}

	timeout := 1 * time.Millisecond
	m := ContextTimeoutWithConfig(ContextTimeoutConfig{
		Timeout:      timeout,
		ErrorHandler: timeoutErrorHandler,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err := m(func(c echo.Context) error {
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		if err := sleepWithContext(c.Request().Context(), time.Duration(2*time.Millisecond)); err != nil {
			return err
		}

		// The Request Context should have a Deadline set by http.ContextTimeoutHandler
		if _, ok := c.Request().Context().Deadline(); !ok {
			assert.Fail(t, "No timeout set on Request Context")
		}
		return c.String(http.StatusOK, "Hello, World!")
	})(c)

	assert.IsType(t, &echo.HTTPError{}, err)
	assert.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, err.(*echo.HTTPError).Code)
	assert.Equal(t, "Timeout! change me", err.(*echo.HTTPError).Message)
}

func TestContextTimeoutWithFullEchoStack(t *testing.T) {
	// test timeout with full http server stack running, do see what http.Server.ErrorLog contains
	var testCases = []struct {
		name                    string
		whenPath                string
		whenForceHandlerTimeout bool
		expectStatusCode        int
		expectResponse          string
		expectLogContains       []string
		expectLogNotContains    []string
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
			name:                    "503 - handler timeouts, write response in timeout middleware",
			whenForceHandlerTimeout: true,
			whenPath:                "/",
			expectResponse:          "{\"message\":\"Service Unavailable\"}\n",
			expectStatusCode:        http.StatusServiceUnavailable,
			expectLogNotContains: []string{
				"echo:http: superfluous response.WriteHeader call from",
			},
			expectLogContains: []string{"Service Unavailable"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			buf := new(coroutineSafeBuffer)
			e.Logger.SetOutput(buf)

			e.Use(Logger())
			// NOTE: timeout middleware is last after recover beacuse timeout middleware will change the response error as 503
			e.Use(ContextTimeoutWithConfig(ContextTimeoutConfig{
				Timeout: 15 * time.Millisecond,
			}))
			e.Use(Recover())

			e.GET("/", func(c echo.Context) error {
				if tc.whenForceHandlerTimeout {
					if err := sleepWithContext(c.Request().Context(), time.Duration(100*time.Millisecond)); err != nil {
						return err
					} // make `sleepWithContext` block until 100ms
				}
				return c.JSON(http.StatusTeapot, map[string]string{"message": "OK"})
			})

			server, addr, err := startServer(e)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			defer server.Close()

			res, err := http.Get(fmt.Sprintf("http://%v%v", addr, tc.whenPath))
			if err != nil {
				assert.NoError(t, err)
				return
			}
			if tc.whenForceHandlerTimeout {
				// shutdown waits for server to shutdown. this way we wait logger mw to be executed
				ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
				defer cancel()
				server.Shutdown(ctx)
			}

			assert.Equal(t, tc.expectStatusCode, res.StatusCode)
			if body, err := io.ReadAll(res.Body); err == nil {
				assert.Equal(t, tc.expectResponse, string(body))
			} else {
				assert.Fail(t, err.Error())
			}

			logged := buf.String()
			for _, subStr := range tc.expectLogContains {
				assert.True(t, strings.Contains(logged, subStr), "expected logs to contain: %v, logged: '%v'", subStr, logged)
			}
			for _, subStr := range tc.expectLogNotContains {
				assert.False(t, strings.Contains(logged, subStr), "expected logs not to contain: %v, logged: '%v'", subStr, logged)
			}
		})
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return context.DeadlineExceeded
	case <-timer.C:
	}
	return nil
}

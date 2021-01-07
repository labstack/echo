// +build go1.15

package echo

/**
	Since version 1.15 time.Time and time.Duration error message pattern has changed (values are wrapped now in \"\")
	So pre 1.15 these tests fail with similar error:

  expected: "code=400, message=failed to bind field value to Duration, internal=time: invalid duration \"nope\", field=param"
  actual  : "code=400, message=failed to bind field value to Duration, internal=time: invalid duration nope, field=param"
*/

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createTestContext15(URL string, body io.Reader, pathParams map[string]string) Context {
	e := New()
	req := httptest.NewRequest(http.MethodGet, URL, body)
	if body != nil {
		req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if len(pathParams) > 0 {
		names := make([]string, 0)
		values := make([]string, 0)
		for name, value := range pathParams {
			names = append(names, name)
			values = append(values, value)
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}

	return c
}

func TestValueBinder_TimeError(t *testing.T) {
	var testCases = []struct {
		name            string
		givenFailFast   bool
		givenBindErrors []error
		whenURL         string
		whenMust        bool
		whenLayout      string
		expectValue     time.Time
		expectError     string
	}{
		{
			name:        "nok, conversion fails, value is not changed",
			whenURL:     "/search?param=nope&param=100",
			expectValue: time.Time{},
			expectError: "code=400, message=failed to bind field value to Time, internal=parsing time \"nope\": extra text: \"nope\", field=param",
		},
		{
			name:        "nok (must), conversion fails, value is not changed",
			whenMust:    true,
			whenURL:     "/search?param=nope&param=100",
			expectValue: time.Time{},
			expectError: "code=400, message=failed to bind field value to Time, internal=parsing time \"nope\": extra text: \"nope\", field=param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := createTestContext15(tc.whenURL, nil, nil)
			b := QueryParamsBinder(c).FailFast(tc.givenFailFast)
			if tc.givenFailFast {
				b.errors = []error{errors.New("previous error")}
			}

			dest := time.Time{}
			var err error
			if tc.whenMust {
				err = b.MustTime("param", &dest, tc.whenLayout).BindError()
			} else {
				err = b.Time("param", &dest, tc.whenLayout).BindError()
			}

			assert.Equal(t, tc.expectValue, dest)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValueBinder_TimesError(t *testing.T) {
	var testCases = []struct {
		name            string
		givenFailFast   bool
		givenBindErrors []error
		whenURL         string
		whenMust        bool
		whenLayout      string
		expectValue     []time.Time
		expectError     string
	}{
		{
			name:          "nok, fail fast without binding value",
			givenFailFast: true,
			whenURL:       "/search?param=1&param=100",
			expectValue:   []time.Time(nil),
			expectError:   "code=400, message=failed to bind field value to Time, internal=parsing time \"1\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"1\" as \"2006\", field=param",
		},
		{
			name:        "nok, conversion fails, value is not changed",
			whenURL:     "/search?param=nope&param=100",
			expectValue: []time.Time(nil),
			expectError: "code=400, message=failed to bind field value to Time, internal=parsing time \"nope\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"nope\" as \"2006\", field=param",
		},
		{
			name:        "nok (must), conversion fails, value is not changed",
			whenMust:    true,
			whenURL:     "/search?param=nope&param=100",
			expectValue: []time.Time(nil),
			expectError: "code=400, message=failed to bind field value to Time, internal=parsing time \"nope\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"nope\" as \"2006\", field=param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := createTestContext15(tc.whenURL, nil, nil)
			b := QueryParamsBinder(c).FailFast(tc.givenFailFast)
			b.errors = tc.givenBindErrors

			layout := time.RFC3339
			if tc.whenLayout != "" {
				layout = tc.whenLayout
			}

			var dest []time.Time
			var err error
			if tc.whenMust {
				err = b.MustTimes("param", &dest, layout).BindError()
			} else {
				err = b.Times("param", &dest, layout).BindError()
			}

			assert.Equal(t, tc.expectValue, dest)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValueBinder_DurationError(t *testing.T) {
	var testCases = []struct {
		name            string
		givenFailFast   bool
		givenBindErrors []error
		whenURL         string
		whenMust        bool
		expectValue     time.Duration
		expectError     string
	}{
		{
			name:        "nok, conversion fails, value is not changed",
			whenURL:     "/search?param=nope&param=100",
			expectValue: 0,
			expectError: "code=400, message=failed to bind field value to Duration, internal=time: invalid duration \"nope\", field=param",
		},
		{
			name:        "nok (must), conversion fails, value is not changed",
			whenMust:    true,
			whenURL:     "/search?param=nope&param=100",
			expectValue: 0,
			expectError: "code=400, message=failed to bind field value to Duration, internal=time: invalid duration \"nope\", field=param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := createTestContext15(tc.whenURL, nil, nil)
			b := QueryParamsBinder(c).FailFast(tc.givenFailFast)
			if tc.givenFailFast {
				b.errors = []error{errors.New("previous error")}
			}

			var dest time.Duration
			var err error
			if tc.whenMust {
				err = b.MustDuration("param", &dest).BindError()
			} else {
				err = b.Duration("param", &dest).BindError()
			}

			assert.Equal(t, tc.expectValue, dest)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValueBinder_DurationsError(t *testing.T) {
	var testCases = []struct {
		name            string
		givenFailFast   bool
		givenBindErrors []error
		whenURL         string
		whenMust        bool
		expectValue     []time.Duration
		expectError     string
	}{
		{
			name:          "nok, fail fast without binding value",
			givenFailFast: true,
			whenURL:       "/search?param=1&param=100",
			expectValue:   []time.Duration(nil),
			expectError:   "code=400, message=failed to bind field value to Duration, internal=time: missing unit in duration \"1\", field=param",
		},
		{
			name:        "nok, conversion fails, value is not changed",
			whenURL:     "/search?param=nope&param=100",
			expectValue: []time.Duration(nil),
			expectError: "code=400, message=failed to bind field value to Duration, internal=time: invalid duration \"nope\", field=param",
		},
		{
			name:        "nok (must), conversion fails, value is not changed",
			whenMust:    true,
			whenURL:     "/search?param=nope&param=100",
			expectValue: []time.Duration(nil),
			expectError: "code=400, message=failed to bind field value to Duration, internal=time: invalid duration \"nope\", field=param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := createTestContext15(tc.whenURL, nil, nil)
			b := QueryParamsBinder(c).FailFast(tc.givenFailFast)
			b.errors = tc.givenBindErrors

			var dest []time.Duration
			var err error
			if tc.whenMust {
				err = b.MustDurations("param", &dest).BindError()
			} else {
				err = b.Durations("param", &dest).BindError()
			}

			assert.Equal(t, tc.expectValue, dest)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

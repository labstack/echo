package echo

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJsonLogger_Write(t *testing.T) {
	var testCases = []struct {
		name   string
		when   []byte
		expect string
	}{
		{
			name:   "ok, write non JSONlike message",
			when:   []byte("version: %v, build: %v"),
			expect: `{"time":"2021-09-07T23:09:37+03:00","level":"WARN","prefix":"echo","message":"version: %v, build: %v"}` + "\n",
		},
		{
			name:   "ok, write quoted message",
			when:   []byte(`version: "%v"`),
			expect: `{"time":"2021-09-07T23:09:37+03:00","level":"WARN","prefix":"echo","message":"version: \"%v\""}` + "\n",
		},
		{
			name:   "ok, write JSON",
			when:   []byte(`{"version": 123}` + "\n"),
			expect: `{"version": 123}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			logger := newJSONLogger(buf)
			logger.timeNow = func() time.Time {
				return time.Unix(1631045377, 0)
			}

			_, err := logger.Write(tc.when)

			result := buf.String()
			assert.Equal(t, tc.expect, result)
			assert.NoError(t, err)
		})
	}
}

func TestJsonLogger_Error(t *testing.T) {
	var testCases = []struct {
		name      string
		whenError error
		expect    string
	}{
		{
			name:      "ok",
			whenError: ErrForbidden,
			expect:    `{"time":"2021-09-07T23:09:37+03:00","level":"ERROR","prefix":"echo","message":"code=403, message=Forbidden"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			logger := newJSONLogger(buf)
			logger.timeNow = func() time.Time {
				return time.Unix(1631045377, 0)
			}

			logger.Error(tc.whenError)

			result := buf.String()
			assert.Equal(t, tc.expect, result)
		})
	}
}

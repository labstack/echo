// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// deserializeJSON decodes body into target via the default serializer using a
// fresh context. It does not touch *testing.T so it is safe to call from
// goroutines (used by the concurrency test).
func deserializeJSON(e *Echo, body string, target any) error {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c := e.NewContext(req, httptest.NewRecorder())
	return DefaultJSONSerializer{}.Deserialize(c, target)
}

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Encode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Echo
	assert.Equal(t, e, c.Echo())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	err := enc.Serialize(c, user{ID: 1, Name: "Jon Snow"}, "")
	if assert.NoError(t, err) {
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Serialize(c, user{ID: 1, Name: "Jon Snow"}, "  ")
	if assert.NoError(t, err) {
		assert.Equal(t, userJSONPretty+"\n", rec.Body.String())
	}
}

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Decode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Echo
	assert.Equal(t, e, c.Echo())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	var u = user{}
	err := enc.Deserialize(c, &u)
	if assert.NoError(t, err) {
		assert.Equal(t, u, user{ID: 1, Name: "Jon Snow"})
	}

	var userUnmarshalSyntaxError = user{}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(invalidContent))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Deserialize(c, &userUnmarshalSyntaxError)
	assert.IsType(t, &HTTPError{}, err)
	assert.EqualError(t, err, "code=400, message=Bad Request, err=invalid character 'i' looking for beginning of value")

	var userUnmarshalTypeError = struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = enc.Deserialize(c, &userUnmarshalTypeError)
	assert.IsType(t, &HTTPError{}, err)
	assert.EqualError(t, err, "code=400, message=Bad Request, err=json: cannot unmarshal number into Go struct field .id of type string")

}

// TestDefaultJSONCodec_Decode_RejectsTrailingData documents an intentional
// behavior change: Deserialize uses json.Unmarshal, which (unlike a streaming
// json.Decoder) rejects any content after the first top-level JSON value.
func TestDefaultJSONCodec_Decode_RejectsTrailingData(t *testing.T) {
	e := New()
	for _, body := range []string{
		userJSON + `{"id":2,"name":"second"}`, // a second JSON object
		userJSON + ` trailing garbage`,        // trailing non-JSON
		userJSON + `,`,                        // trailing token
	} {
		var u user
		err := deserializeJSON(e, body, &u)
		if assert.Error(t, err, "body %q should be rejected", body) {
			assert.IsType(t, &HTTPError{}, err)
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
		}
	}
}

// TestDefaultJSONCodec_Decode_PooledBufferReuse guards against stale bytes
// bleeding between requests through the reused pooled buffer: a long body
// followed by a short one must each decode to exactly their own input.
func TestDefaultJSONCodec_Decode_PooledBufferReuse(t *testing.T) {
	e := New()
	for i := 0; i < 50; i++ {
		longName := strings.Repeat("x", 1000+i)
		var long user
		err := deserializeJSON(e, fmt.Sprintf(`{"id":%d,"name":%q}`, i, longName), &long)
		assert.NoError(t, err)
		assert.Equal(t, user{ID: i, Name: longName}, long)

		var short user
		err = deserializeJSON(e, `{"id":7,"name":"a"}`, &short)
		assert.NoError(t, err)
		assert.Equal(t, user{ID: 7, Name: "a"}, short)
	}
}

// TestDefaultJSONCodec_Decode_PooledBufferConcurrent exercises the pooled
// buffer from many goroutines at once; run under -race it catches any aliasing
// or missing-reset regression that would let one request's body corrupt another.
func TestDefaultJSONCodec_Decode_PooledBufferConcurrent(t *testing.T) {
	e := New()
	const n = 64
	var wg sync.WaitGroup
	errs := make([]error, n)
	got := make([]user, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"id":%d,"name":%q}`, i, strings.Repeat("n", i+1))
			errs[i] = deserializeJSON(e, body, &got[i])
		}(i)
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		assert.NoError(t, errs[i])
		assert.Equal(t, user{ID: i, Name: strings.Repeat("n", i+1)}, got[i])
	}
}

// TestDefaultJSONCodec_Decode_LargeBodyThenNormal covers the buffer-cap path: a
// body larger than maxPooledJSONBuf must decode correctly, and its oversized
// buffer (dropped from the pool rather than retained) must not affect the next
// normal-sized request.
func TestDefaultJSONCodec_Decode_LargeBodyThenNormal(t *testing.T) {
	e := New()
	bigName := strings.Repeat("z", 100*1024) // 100 KiB > 64 KiB cap
	var big user
	err := deserializeJSON(e, fmt.Sprintf(`{"id":1,"name":%q}`, bigName), &big)
	assert.NoError(t, err)
	assert.Equal(t, user{ID: 1, Name: bigName}, big)

	var small user
	err = deserializeJSON(e, userJSON, &small)
	assert.NoError(t, err)
	assert.Equal(t, user{ID: 1, Name: "Jon Snow"}, small)
}

// errReader is an io.ReadCloser whose Read always fails, used to exercise the
// body-read error branch of Deserialize.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errReader) Close() error             { return nil }

// TestDefaultJSONCodec_Decode_BodyReadError verifies a failing request body read
// surfaces as a 400, matching the pre-existing decoder behavior.
func TestDefaultJSONCodec_Decode_BodyReadError(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	req.Body = errReader{}
	c := e.NewContext(req, httptest.NewRecorder())

	var u user
	err := DefaultJSONSerializer{}.Deserialize(c, &u)
	if assert.Error(t, err) {
		assert.IsType(t, &HTTPError{}, err)
		assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
	}
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBodyLimit(t *testing.T) {
	e := echo.New()
	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	// Based on content length (within limit)
	if assert.NoError(t, BodyLimit("2M")(h)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.Bytes())
	}

	// Based on content length (overlimit)
	he := BodyLimit("2B")(h)(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code)

	// Based on content read (within limit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, BodyLimit("2M")(h)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// Based on content read (overlimit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	he = BodyLimit("2B")(h)(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code)
}

func TestBodyLimitReader(t *testing.T) {
	hw := []byte("Hello, World!")

	config := BodyLimitConfig{
		Skipper: DefaultSkipper,
		Limit:   "2B",
		limit:   2,
	}
	reader := &limitedReader{
		BodyLimitConfig: config,
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// read all should return ErrStatusRequestEntityTooLarge
	_, err := io.ReadAll(reader)
	he := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code)

	// A new request gets a fresh reader.
	bt := make([]byte, 2)
	reader = &limitedReader{BodyLimitConfig: config, reader: io.NopCloser(bytes.NewReader(hw))}
	n, err := reader.Read(bt)
	assert.Equal(t, 2, n)
	assert.Equal(t, nil, err)
}

func TestBodyLimitReaderDoesNotExposeExtraBytes(t *testing.T) {
	for _, chunkSize := range []int{1, 64} {
		t.Run(fmt.Sprint(chunkSize), func(t *testing.T) {
			reader := &limitedReader{
				BodyLimitConfig: BodyLimitConfig{limit: 5},
				reader:          io.NopCloser(bytes.NewReader([]byte("123456789"))),
			}
			buf := make([]byte, chunkSize)
			var got []byte
			for {
				n, err := reader.Read(buf)
				got = append(got, buf[:n]...)
				if err != nil {
					assert.Equal(t, []byte("12345"), got)
					assertBodyLimitError(t, err)
					break
				}
			}
			n, err := reader.Read(buf)
			assert.Zero(t, n)
			assertBodyLimitError(t, err)
		})
	}
}

func TestBodyLimitReaderBoundaryAndLargeLimit(t *testing.T) {
	for _, limit := range []int64{5, math.MaxInt64} {
		reader := &limitedReader{
			BodyLimitConfig: BodyLimitConfig{limit: limit},
			reader:          io.NopCloser(bytes.NewReader([]byte("12345"))),
		}
		got, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "12345", string(got))
	}
}

func TestBodyLimitReaderKeepsOrdinaryReadErrors(t *testing.T) {
	broken := errors.New("source failed")
	reader := &limitedReader{
		BodyLimitConfig: BodyLimitConfig{limit: 5},
		reader:          io.NopCloser(&errorOnceReader{err: broken}),
	}
	buf := make([]byte, 2)
	_, err := reader.Read(buf)
	assert.ErrorIs(t, err, broken)
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte('x'), buf[0])
}

type errorOnceReader struct {
	err error
}

func (r *errorOnceReader) Read(b []byte) (int, error) {
	if r.err != nil {
		err := r.err
		r.err = nil
		return 0, err
	}
	return copy(b, "x"), nil
}

func TestBodyLimitReaderRemainsAttachedToRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("123456"))
	req.ContentLength = -1
	c := e.NewContext(req, httptest.NewRecorder())
	err := BodyLimit("5B")(func(c echo.Context) error { return nil })(c)
	assert.NoError(t, err)
	got, err := io.ReadAll(req.Body)
	assert.Equal(t, "12345", string(got))
	assertBodyLimitError(t, err)
}

func TestBodyLimitBindRejectsOversizeBodies(t *testing.T) {
	var multipartBody bytes.Buffer
	writer := multipart.NewWriter(&multipartBody)
	if err := writer.WriteField("x", "12345"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, contentType string
		body              []byte
	}{
		{"JSON", echo.MIMEApplicationJSON, []byte(`{"x":"12345"}`)},
		{"XML", echo.MIMEApplicationXML, []byte(`<x>12345</x>`)},
		{"form", echo.MIMEApplicationForm, []byte(`x=12345`)},
		{"multipart", writer.FormDataContentType(), multipartBody.Bytes()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(BodyLimit("5B"))
			e.POST("/", func(c echo.Context) error {
				var payload struct {
					X string `json:"x" xml:",chardata" form:"x"`
				}
				return c.Bind(&payload)
			})
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.body))
			req.ContentLength = -1
			req.Header.Set(echo.HeaderContentType, tc.contentType)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code, rec.Body.String())
		})
	}
}

func assertBodyLimitError(t *testing.T, err error) {
	t.Helper()
	var httpError *echo.HTTPError
	if !errors.As(err, &httpError) {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	assert.Equal(t, http.StatusRequestEntityTooLarge, httpError.Code)
}

func TestBodyLimitWithConfig_Skipper(t *testing.T) {
	e := echo.New()
	h := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}
	mw := BodyLimitWithConfig(BodyLimitConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		Limit: "2B", // if not skipped this limit would make request to fail limit check
	})

	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, hw, rec.Body.Bytes())
}

func TestBodyLimitWithConfig(t *testing.T) {
	var testCases = []struct {
		name        string
		givenLimit  string
		whenBody    []byte
		expectBody  []byte
		expectError string
	}{
		{
			name:        "ok, body is less than limit",
			givenLimit:  "10B",
			whenBody:    []byte("123456789"),
			expectBody:  []byte("123456789"),
			expectError: "",
		},
		{
			name:        "nok, body is more than limit",
			givenLimit:  "9B",
			whenBody:    []byte("1234567890"),
			expectBody:  []byte(nil),
			expectError: "code=413, message=Request Entity Too Large",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			h := func(c echo.Context) error {
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				return c.String(http.StatusOK, string(body))
			}
			mw := BodyLimitWithConfig(BodyLimitConfig{
				Limit: tc.givenLimit,
			})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.whenBody))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := mw(h)(c)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			// not testing status as middlewares return error instead of committing it and OK cases are anyway 200
			assert.Equal(t, tc.expectBody, rec.Body.Bytes())
		})
	}
}

func TestBodyLimit_panicOnInvalidLimit(t *testing.T) {
	assert.PanicsWithError(
		t,
		"echo: invalid body-limit=",
		func() { BodyLimit("") },
	)
	assert.PanicsWithError(
		t,
		"echo: invalid body-limit=-1B",
		func() { BodyLimit("-1B") },
	)
}

func TestBodyLimit_Middleware_BodyRestoration(t *testing.T) {
	e := echo.New()

	e.Use(BodyLimit("1KB"))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}

			c.Request().Body.Close()

			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			return next(c)
		}
	})

	e.POST("/", func(c echo.Context) error {
		type Payload struct {
			Message string `json:"message"`
		}
		p := new(Payload)
		if err := c.Bind(p); err != nil {
			return err
		}
		return c.String(http.StatusOK, p.Message)
	})

	t.Run("valid request under 1KB binds successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"message": "hello"}`)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello", rec.Body.String())
	})

	t.Run("request exceeding 1KB returns 413 at middleware read phase", func(t *testing.T) {
		largePayload := `{"message": "` + string(bytes.Repeat([]byte("A"), 2000)) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(largePayload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})
}

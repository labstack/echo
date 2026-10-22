// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestDecompressSizeLimits(t *testing.T) {
	for _, limit := range []int64{-1, 0, 7, 8, 9} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			zipped, err := gzipString("abcdefgh")
			require.NoError(t, err)
			e := echo.New()
			e.Use(DecompressWithConfig(DecompressConfig{MaxDecompressedSize: limit}))
			e.POST("/", func(c echo.Context) error {
				require.EqualValues(t, -1, c.Request().ContentLength)
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				return c.String(http.StatusOK, string(body))
			})
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(zipped))
			req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if limit == 7 {
				require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
			} else {
				require.Equal(t, http.StatusOK, rec.Code)
				require.Equal(t, "abcdefgh", rec.Body.String())
			}
		})
	}
}

func TestDecompressReadBoundary(t *testing.T) {
	for _, payload := range []string{"A", "AB"} {
		t.Run(payload, func(t *testing.T) {
			zipped, err := gzipString(payload)
			require.NoError(t, err)
			gr, err := gzip.NewReader(bytes.NewReader(zipped))
			require.NoError(t, err)
			defer gr.Close()
			reader := &limitedGzipReader{Reader: gr, remaining: 1}
			buf := make([]byte, 1)
			_, err = io.ReadFull(reader, buf)
			require.NoError(t, err)
			_, err = reader.Read(nil)
			require.NoError(t, err)
			for i := 0; i < 2; i++ {
				n, err := reader.Read(buf)
				require.Zero(t, n)
				if payload == "A" {
					require.ErrorIs(t, err, io.EOF)
				} else {
					require.ErrorIs(t, err, echo.ErrStatusRequestEntityTooLarge)
				}
			}
		})
	}
}

func TestDecompressDefaultLimitStreaming(t *testing.T) {
	var zipped bytes.Buffer
	zw := gzip.NewWriter(&zipped)
	// Generate compressed input incrementally; the test never buffers a 101 MiB body.
	for i := 0; i < 101; i++ {
		_, err := io.Copy(zw, strings.NewReader(strings.Repeat("x", 1<<20)))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	e := echo.New()
	e.Use(Decompress())
	var read int64
	e.POST("/", func(c echo.Context) error {
		var err error
		read, err = io.Copy(io.Discard, c.Request().Body)
		return err
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(zipped.Bytes()))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	require.EqualValues(t, 100<<20, read)
}

func TestDecompressBindingLimitStatus(t *testing.T) {
	for _, tc := range []struct {
		name, contentType, payload string
	}{
		{"json", echo.MIMEApplicationJSON, `{"value":"` + strings.Repeat("x", 32) + `"}`},
		{"xml", echo.MIMEApplicationXML, `<payload><value>` + strings.Repeat("x", 32) + `</value></payload>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			zipped, err := gzipString(tc.payload)
			require.NoError(t, err)
			e := echo.New()
			e.Use(DecompressWithConfig(DecompressConfig{MaxDecompressedSize: 16}))
			accepted := false
			var bindErr error
			e.POST("/", func(c echo.Context) error {
				var value struct {
					Value string `json:"value" xml:"value"`
				}
				if bindErr = c.Bind(&value); bindErr != nil {
					return bindErr
				}
				accepted = true
				return c.NoContent(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(zipped))
			req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
			req.Header.Set(echo.HeaderContentType, tc.contentType)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.ErrorIs(t, bindErr, echo.ErrStatusRequestEntityTooLarge)
			require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code, "v4 binding preserves the size-limit read error")
			require.False(t, accepted)
		})
	}
}

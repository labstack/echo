// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestBodyDumpCaptureLimits(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		reqLimit, resLimit     int64
		size, reqWant, resWant int
	}{
		{"configured", 3, 4, 8, 3, 4},
		{"exact", 8, 8, 8, 8, 8},
		{"unlimited", -1, -1, 8, 8, 8},
		{"default", 0, 0, 6 << 20, 5 << 20, 5 << 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := strings.Repeat("x", tc.size)
			var requestDump, responseDump []byte
			e := echo.New()
			e.Use(BodyDumpWithConfig(BodyDumpConfig{MaxRequestBytes: tc.reqLimit, MaxResponseBytes: tc.resLimit,
				Handler: func(_ echo.Context, req, res []byte) { requestDump, responseDump = req, res }}))
			e.POST("/", func(c echo.Context) error {
				body, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}
				require.Equal(t, payload, string(body))
				return c.String(http.StatusOK, string(body))
			})
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload)))
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, payload, rec.Body.String())
			require.Equal(t, payload[:tc.reqWant], string(requestDump))
			require.Equal(t, payload[:tc.resWant], string(responseDump))
		})
	}
}

type bodyDumpCloseTracker struct {
	io.Reader
	closed bool
}

func (b *bodyDumpCloseTracker) Close() error { b.closed = true; return nil }

func TestBodyDumpPreservesRemainderErrorAndClose(t *testing.T) {
	wantErr := errors.New("remainder failed")
	body := &bodyDumpCloseTracker{Reader: io.MultiReader(strings.NewReader("abcdefgh"), iotest.ErrReader(wantErr))}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = body
	e := echo.New()
	c := e.NewContext(req, httptest.NewRecorder())
	var dumped []byte
	err := BodyDumpWithConfig(BodyDumpConfig{MaxRequestBytes: 3, Handler: func(_ echo.Context, req, res []byte) { dumped = req }})(func(c echo.Context) error {
		require.False(t, body.closed)
		data, err := io.ReadAll(c.Request().Body)
		require.Equal(t, "abcdefgh", string(data))
		require.ErrorIs(t, err, wantErr)
		require.NoError(t, c.Request().Body.Close())
		require.True(t, body.closed)
		return err
	})(c)
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, "abc", string(dumped))
}

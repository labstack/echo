// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2026 LabStack LLC and Echo contributors

package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestBodyDumpPreservesBodyLimit(t *testing.T) {
	for _, dumpFirst := range []bool{false, true} {
		for _, knownLength := range []bool{false, true} {
			t.Run(fmt.Sprintf("dumpFirst=%t/knownLength=%t", dumpFirst, knownLength), func(t *testing.T) {
				e := echo.New()
				dump := BodyDumpWithConfig(BodyDumpConfig{MaxRequestBytes: 4, Handler: func(*echo.Context, []byte, []byte, error) {}})
				limit := BodyLimit(5)
				if dumpFirst {
					e.Use(dump, limit)
				} else {
					e.Use(limit, dump)
				}
				accepted := false
				e.POST("/", func(c *echo.Context) error {
					if _, err := io.ReadAll(c.Request().Body); err != nil {
						return err
					}
					accepted = true
					return c.NoContent(http.StatusOK)
				})
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("ABCDEFGH"))
				if !knownLength {
					req.ContentLength = -1
					req.TransferEncoding = []string{"chunked"}
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
				require.False(t, accepted)
			})
		}
	}
}

func TestBodyDumpPreservesBodyLimitChunkedHTTP(t *testing.T) {
	e := echo.New()
	e.Use(BodyLimit(5), BodyDumpWithConfig(BodyDumpConfig{MaxRequestBytes: 4, Handler: func(*echo.Context, []byte, []byte, error) {}}))
	e.POST("/", func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	})
	server := httptest.NewServer(e)
	defer server.Close()
	req, err := http.NewRequest(http.MethodPost, server.URL, io.NopCloser(strings.NewReader("ABCDEFGH")))
	require.NoError(t, err)
	req.TransferEncoding = []string{"chunked"}
	response, err := server.Client().Do(req)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
}

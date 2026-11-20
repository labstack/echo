// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipPreservesProxyContentEncoding(t *testing.T) {
	payload := []byte("encoded upstream response")
	for _, encoding := range []string{"gzip", "deflate"} {
		var encoded bytes.Buffer
		var compressor io.WriteCloser
		if encoding == "gzip" {
			compressor = gzip.NewWriter(&encoded)
		} else {
			compressor = zlib.NewWriter(&encoded)
		}
		_, err := compressor.Write(payload)
		require.NoError(t, err)
		require.NoError(t, compressor.Close())
		wire := append([]byte(nil), encoded.Bytes()...)
		for _, minLength := range []int{0, 8, 4096} {
			t.Run(fmt.Sprintf("%s/min=%d", encoding, minLength), func(t *testing.T) {
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set(echo.HeaderContentEncoding, encoding)
					w.Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
					w.Header().Set(echo.HeaderContentLength, strconv.Itoa(len(wire)))
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write(wire)
				}))
				defer upstream.Close()
				target, err := url.Parse(upstream.URL)
				require.NoError(t, err)
				e := echo.New()
				e.Use(GzipWithConfig(GzipConfig{MinLength: minLength}))
				e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
				proxy := httptest.NewServer(e)
				defer proxy.Close()
				transport := &http.Transport{DisableCompression: true}
				defer transport.CloseIdleConnections()
				client := &http.Client{Transport: transport}
				req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
				require.NoError(t, err)
				req.Header.Set(echo.HeaderAcceptEncoding, "gzip, deflate")
				res, err := client.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, res.StatusCode)
				assert.Equal(t, encoding, res.Header.Get(echo.HeaderContentEncoding))
				assert.Equal(t, wire, body, "proxy must preserve upstream encoded bytes")
				var decoder io.ReadCloser
				if encoding == "gzip" {
					decoder, err = gzip.NewReader(bytes.NewReader(body))
				} else {
					decoder, err = zlib.NewReader(bytes.NewReader(body))
				}
				require.NoError(t, err)
				defer decoder.Close()
				decoded, err := io.ReadAll(decoder)
				require.NoError(t, err)
				assert.Equal(t, payload, decoded, "one decoding must recover the original response")
			})
		}
	}
}

func TestGzipPreservesEncodedHandlerResponse(t *testing.T) {
	wire := []byte("already encoded body")
	for _, mode := range []string{"explicit header", "implicit header", "flush", "head"} {
		t.Run(mode, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			handler := Gzip()(func(c *echo.Context) error {
				w := c.Response()
				w.Header().Set(echo.HeaderContentEncoding, gzipScheme)
				w.Header().Set(echo.HeaderContentLength, strconv.Itoa(len(wire)))
				if mode != "implicit header" {
					w.WriteHeader(http.StatusCreated)
				}
				if mode == "head" {
					return nil
				}
				if mode == "flush" {
					require.NoError(t, http.NewResponseController(w).Flush())
				}
				n, err := w.Write(wire)
				assert.Equal(t, len(wire), n)
				return err
			})
			require.NoError(t, handler(c))
			expectedStatus := http.StatusCreated
			if mode == "implicit header" {
				expectedStatus = http.StatusOK
			}
			assert.Equal(t, expectedStatus, rec.Code)
			assert.Equal(t, gzipScheme, rec.Result().Header.Get(echo.HeaderContentEncoding))
			assert.Equal(t, strconv.Itoa(len(wire)), rec.Result().Header.Get(echo.HeaderContentLength))
			if mode == "head" {
				assert.Empty(t, rec.Body.Bytes())
			} else {
				assert.Equal(t, wire, rec.Body.Bytes())
			}
		})
	}
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
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

func TestGzip_PreservesProxyContentEncoding(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		minLength          int
		emptyEncodingField bool
	}{
		{"get", http.MethodGet, 0, false},
		{"get above threshold", http.MethodGet, 8, false},
		{"get below threshold", http.MethodGet, 4096, false},
		{"head", http.MethodHead, 0, false},
		{"empty first encoding field", http.MethodGet, 0, true},
	}
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
		wire := encoded.Bytes()

		for _, tt := range tests {
			t.Run(encoding+"/"+tt.name, func(t *testing.T) {
				encodings := []string{encoding}
				if tt.emptyEncodingField {
					encodings = []string{"", encoding}
				}
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, tt.method, r.Method)
					for _, value := range encodings {
						w.Header().Add(echo.HeaderContentEncoding, value)
					}
					w.Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
					w.Header().Set(echo.HeaderContentLength, strconv.Itoa(len(wire)))
					w.WriteHeader(http.StatusCreated)
					if r.Method != http.MethodHead {
						_, _ = w.Write(wire)
					}
				}))
				defer upstream.Close()
				target, err := url.Parse(upstream.URL)
				require.NoError(t, err)

				e := echo.New()
				e.Use(GzipWithConfig(GzipConfig{MinLength: tt.minLength}))
				e.Use(Proxy(NewRoundRobinBalancer([]*ProxyTarget{{URL: target}})))
				proxy := httptest.NewServer(e)
				defer proxy.Close()

				transport := &http.Transport{DisableCompression: true}
				defer transport.CloseIdleConnections()
				client := &http.Client{Transport: transport}
				req, err := http.NewRequest(tt.method, proxy.URL, nil)
				require.NoError(t, err)
				req.Header.Set(echo.HeaderAcceptEncoding, "gzip, deflate")
				res, err := client.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, http.StatusCreated, res.StatusCode)
				assert.Equal(t, encodings, res.Header.Values(echo.HeaderContentEncoding))
				assert.Equal(t, int64(len(wire)), res.ContentLength)
				if tt.method == http.MethodHead {
					assert.Empty(t, body)
				} else {
					assert.Equal(t, wire, body, "proxy must preserve the upstream encoded bytes")
				}
			})
		}
	}
}

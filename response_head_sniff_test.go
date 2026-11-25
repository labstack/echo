// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouterAutoHandleHEAD_SniffContentType(t *testing.T) {
	for _, tc := range []struct {
		name     string
		chunks   []string
		headers  http.Header
		status   int
		wantType string
	}{
		{name: "HTML", chunks: []string{"<!DOCTYPE html><html><body>hello</body></html>"}, wantType: "text/html; charset=utf-8"},
		{name: "plain text", chunks: []string{"hello world"}, wantType: "text/plain; charset=utf-8"},
		{name: "split HTML", chunks: []string{"<", "!DOC", "TYPE html><html>hello</html>"}, wantType: "text/html; charset=utf-8"},
		{name: "leading empty write", chunks: []string{"", "<html>hello</html>"}, wantType: "text/html; charset=utf-8"},
		{name: "binary", chunks: []string{string([]byte{0, 1, 2, 3})}, wantType: "application/octet-stream"},
		{name: "bounded prefix", chunks: []string{string(bytes.Repeat([]byte(" "), 512)), "<html>hello</html>"}, wantType: "text/plain; charset=utf-8"},
		{name: "explicit type", chunks: []string{"<html>hello</html>"}, headers: http.Header{HeaderContentType: {"application/custom"}}, wantType: "application/custom"},
		{name: "suppressed type", chunks: []string{"<html>hello</html>"}, headers: http.Header{HeaderContentType: nil}},
		{name: "encoded content", chunks: []string{"<html>hello</html>"}, headers: http.Header{HeaderContentEncoding: {"gzip"}}},
		{name: "empty encoding", chunks: []string{"<html>hello</html>"}, headers: http.Header{HeaderContentEncoding: {""}}, wantType: "text/html; charset=utf-8"},
		{name: "no writes"},
		{name: "empty write", chunks: []string{""}},
		{name: "accepted", chunks: []string{"<html>hello</html>"}, status: http.StatusAccepted, wantType: "text/html; charset=utf-8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewWithConfig(Config{Router: NewRouter(RouterConfig{AutoHandleHEAD: true})})
			e.GET("/", func(c *Context) error {
				for k, v := range tc.headers {
					c.Response().Header()[k] = v
				}
				if tc.status != 0 {
					c.Response().WriteHeader(tc.status)
				}
				for _, chunk := range tc.chunks {
					if _, err := io.WriteString(c.Response(), chunk); err != nil {
						return err
					}
				}
				return nil
			})
			server := httptest.NewServer(e)
			defer server.Close()
			client := server.Client()
			client.Transport.(*http.Transport).DisableCompression = true
			for _, method := range []string{http.MethodGet, http.MethodHead} {
				req, err := http.NewRequest(method, server.URL, nil)
				require.NoError(t, err)
				res, err := client.Do(req)
				require.NoError(t, err)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
				require.Equal(t, tc.wantType, res.Header.Get(HeaderContentType), method)
				require.Equal(t, max(http.StatusOK, tc.status), res.StatusCode, method)
				if method == http.MethodHead {
					require.Empty(t, body)
				}
			}
		})
	}
}

func TestHeadResponseWriter_SniffPrefixBounded(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}
	payload := bytes.Repeat([]byte("x"), 1<<20)
	for range 3 {
		n, err := w.Write(payload)
		require.NoError(t, err)
		require.Equal(t, len(payload), n)
	}
	require.Len(t, w.sniffBuf, 512)
	require.Equal(t, int64(3*len(payload)), w.bodyBytes)
	require.Empty(t, rec.Body.String())
}

func TestHeadResponseWriter_SniffSkipsBodylessStatusAndTransferEncoding(t *testing.T) {
	for _, status := range []int{http.StatusSwitchingProtocols, http.StatusNoContent, http.StatusResetContent, http.StatusNotModified} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := &headResponseWriter{rw: rec}
			w.WriteHeader(status)
			_, err := w.Write([]byte("<html>ignored</html>"))
			require.NoError(t, err)
			w.commit()
			require.Empty(t, rec.Header().Get(HeaderContentType))
		})
	}
	rec := httptest.NewRecorder()
	rec.Header().Set("Transfer-Encoding", "chunked")
	w := &headResponseWriter{rw: rec}
	_, err := w.Write([]byte("<html>ignored</html>"))
	require.NoError(t, err)
	w.commit()
	require.Empty(t, rec.Header().Get(HeaderContentType))
}

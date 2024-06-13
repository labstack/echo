package middleware

import (
	"net/http"
	"testing"
)

func Test_getScheme(t *testing.T) {
	tests := []struct {
		name       string
		r          *http.Request
		headerName string
		whenHeader string
		want       string
	}{
		{
			name:       "test only X-Forwarded-Proto: https",
			headerName: "X-Forwarded-Proto",
			whenHeader: "https",
			want:       "https",
		},
		{
			name:       "test only X-Forwarded-Proto: http",
			headerName: "X-Forwarded-Proto",
			whenHeader: "http",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Proto: HTTP",
			headerName: "X-Forwarded-Proto",
			whenHeader: "HTTP",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Protocol: https",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "https",
			want:       "https",
		},
		{
			name:       "test only X-Forwarded-Protocol: http",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "http",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Protocol: HTTP",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "HTTP",
			want:       "http",
		},
		{
			name:       "test only Forwarded https",
			headerName: "Forwarded",
			whenHeader: "proto=https",
			want:       "https",
		},
		{
			name:       "test only Forwarded: http",
			headerName: "Forwarded",
			whenHeader: "proto=http",
			want:       "http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{
					tt.headerName: []string{tt.whenHeader},
				},
			}

			if got := getScheme(req); got != tt.want {
				t.Errorf("getScheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

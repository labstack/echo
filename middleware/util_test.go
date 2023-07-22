package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_matchScheme(t *testing.T) {
	tests := []struct {
		domain, pattern string
		expected        bool
	}{
		{
			domain:   "http://example.com",
			pattern:  "http://example.com",
			expected: true,
		},
		{
			domain:   "https://example.com",
			pattern:  "https://example.com",
			expected: true,
		},
		{
			domain:   "http://example.com",
			pattern:  "https://example.com",
			expected: false,
		},
		{
			domain:   "https://example.com",
			pattern:  "http://example.com",
			expected: false,
		},
	}

	for _, v := range tests {
		assert.Equal(t, v.expected, matchScheme(v.domain, v.pattern))
	}
}

func Test_matchSubdomain(t *testing.T) {
	tests := []struct {
		domain, pattern string
		expected        bool
	}{
		{
			domain:   "http://aaa.example.com",
			pattern:  "http://*.example.com",
			expected: true,
		},
		{
			domain:   "http://bbb.aaa.example.com",
			pattern:  "http://*.example.com",
			expected: true,
		},
		{
			domain:   "http://bbb.aaa.example.com",
			pattern:  "http://*.aaa.example.com",
			expected: true,
		},
		{
			domain:   "http://aaa.example.com:8080",
			pattern:  "http://*.example.com:8080",
			expected: true,
		},

		{
			domain:   "http://fuga.hoge.com",
			pattern:  "http://*.example.com",
			expected: false,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://*.aaa.example.com",
			expected: false,
		},
		{
			domain: `http://1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
      .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
      .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
      .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.example.com`,
			pattern:  "http://*.example.com",
			expected: false,
		},
		{
			domain:   "http://ccc.bbb.example.com",
			pattern:  "http://example.com",
			expected: false,
		},
	}

	for _, v := range tests {
		assert.Equal(t, v.expected, matchSubdomain(v.domain, v.pattern))
	}
}

func TestRandomString(t *testing.T) {
	var testCases = []struct {
		name       string
		whenLength uint8
		expect     string
	}{
		{
			name:       "ok, 16",
			whenLength: 16,
		},
		{
			name:       "ok, 32",
			whenLength: 32,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uid := randomString(tc.whenLength)
			assert.Len(t, uid, int(tc.whenLength))
		})
	}
}

func TestRandomStringBias(t *testing.T) {
	t.Parallel()
	const slen = 33
	const loop = 100000

	counts := make(map[rune]int)
	var count int64

	for i := 0; i < loop; i++ {
		s := randomString(slen)
		require.Equal(t, slen, len(s))
		for _, b := range s {
			counts[b]++
			count++
		}
	}

	require.Equal(t, randomStringCharsetLen, len(counts))

	avg := float64(count) / float64(len(counts))
	for k, n := range counts {
		diff := float64(n) / avg
		if diff < 0.95 || diff > 1.05 {
			t.Errorf("Bias on '%c': expected average %f, got %d", k, avg, n)
		}
	}
}

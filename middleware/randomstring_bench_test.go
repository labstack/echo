// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bufio"
	"io"
	"testing"
)

// randomStringUnpooled is the previous (pre-pooling) implementation, kept here only to A/B benchmark
// against the current pooled randomString.
func randomStringUnpooled(length uint8) string {
	reader := randomReaderPool.Get().(*bufio.Reader)
	defer randomReaderPool.Put(reader)

	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	var i uint8 = 0
	for {
		if _, err := io.ReadFull(reader, r); err != nil {
			panic("unexpected error reading from crypto/rand")
		}
		for _, rb := range r {
			if rb > randomStringMaxByte {
				continue
			}
			b[i] = randomStringCharset[rb%randomStringCharsetLen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

func BenchmarkRandomString_Unpooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = randomStringUnpooled(32)
	}
}

func BenchmarkRandomString_Pooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = randomString(32)
	}
}

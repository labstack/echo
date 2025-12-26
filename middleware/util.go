// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
)

const (
	_ = int64(1 << (10 * iota)) // ignore first value by assigning to blank identifier
	// KB is 1 KiloByte = 1024 bytes
	KB
	// MB is 1 Megabyte = 1_048_576 bytes
	MB
	// GB is 1 Gigabyte = 1_073_741_824 bytes
	GB
	// TB is 1 Terabyte = 1_099_511_627_776 bytes
	TB
	// PB is 1 Petabyte = 1_125_899_906_842_624 bytes
	PB
	// EB is 1 Exabyte = 1_152_921_504_606_847_000 bytes
	EB
)

func matchScheme(domain, pattern string) bool {
	didx := strings.Index(domain, ":")
	pidx := strings.Index(pattern, ":")
	return didx != -1 && pidx != -1 && domain[:didx] == pattern[:pidx]
}

func createRandomStringGenerator(length uint8) func() string {
	return func() string {
		return randomString(length)
	}
}

// https://tip.golang.org/doc/go1.19#:~:text=Read%20no%20longer%20buffers%20random%20data%20obtained%20from%20the%20operating%20system%20between%20calls
var randomReaderPool = sync.Pool{New: func() any {
	return bufio.NewReader(rand.Reader)
}}

const randomStringCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const randomStringCharsetLen = 52 // len(randomStringCharset)
const randomStringMaxByte = 255 - (256 % randomStringCharsetLen)

func randomString(length uint8) string {
	reader := randomReaderPool.Get().(*bufio.Reader)
	defer randomReaderPool.Put(reader)

	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // perf: avoid read from rand.Reader many times
	var i uint8 = 0

	// security note:
	// we can't just simply do b[i]=randomStringCharset[rb%len(randomStringCharset)],
	// len(len(randomStringCharset)) is 52, and rb is [0, 255], 256 = 52 * 4 + 48.
	// make the first 48 characters more possibly to be generated then others.
	// So we have to skip bytes when rb > randomStringMaxByt

	for {
		_, err := io.ReadFull(reader, r)
		if err != nil {
			panic("unexpected error happened when reading from bufio.NewReader(crypto/rand.Reader)")
		}
		for _, rb := range r {
			if rb > randomStringMaxByte {
				// Skip this number to avoid bias.
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

func validateOrigins(origins []string, what string) error {
	for _, o := range origins {
		if err := validateOrigin(o, what); err != nil {
			return err
		}
	}
	return nil
}

func validateOrigin(origin string, what string) error {
	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("can not parse %s: %w", what, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%s is missing scheme or host: %s", what, origin)
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("%s can not have path, query, and fragments: %s", what, origin)
	}
	return nil
}

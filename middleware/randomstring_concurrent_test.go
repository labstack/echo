// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"strings"
	"sync"
	"testing"
)

// TestRandomStringConcurrent guards the pooled scratch buffers in randomString: concurrent callers
// must not share/alias a buffer and corrupt each other's output. Run with -race.
func TestRandomStringConcurrent(t *testing.T) {
	const goroutines, iterations = 100, 300
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				s := randomString(32)
				if len(s) != 32 {
					t.Errorf("expected length 32, got %d (%q)", len(s), s)
					return
				}
				for _, r := range s {
					if !strings.ContainsRune(randomStringCharset, r) {
						t.Errorf("char %q not in charset (%q)", r, s)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

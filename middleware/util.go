package middleware

import (
	"crypto/rand"
	"fmt"
	"strings"
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

// matchSubdomain compares authority with wildcard
func matchSubdomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}
	didx := strings.Index(domain, "://")
	pidx := strings.Index(pattern, "://")
	if didx == -1 || pidx == -1 {
		return false
	}
	domAuth := domain[didx+3:]
	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}
	patAuth := pattern[pidx+3:]

	domComp := strings.Split(domAuth, ".")
	patComp := strings.Split(patAuth, ".")
	for i := len(domComp)/2 - 1; i >= 0; i-- {
		opp := len(domComp) - 1 - i
		domComp[i], domComp[opp] = domComp[opp], domComp[i]
	}
	for i := len(patComp)/2 - 1; i >= 0; i-- {
		opp := len(patComp) - 1 - i
		patComp[i], patComp[opp] = patComp[opp], patComp[i]
	}

	for i, v := range domComp {
		if len(patComp) <= i {
			return false
		}
		p := patComp[i]
		if p == "*" {
			return true
		}
		if p != v {
			return false
		}
	}
	return false
}

func createRandomStringGenerator(length uint8) func() string {
	return func() string {
		return randomString(length)
	}
}

func randomString(length uint8) string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// we are out of random. let the request fail
		panic(fmt.Errorf("echo randomString failed to read random bytes: %w", err))
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

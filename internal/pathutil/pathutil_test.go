// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package pathutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasEncodedPathSeparator(t *testing.T) {
	for s, want := range map[string]bool{
		"foo/bar.txt": false,
		"100%25.txt":  false, // encoded percent, not a separator
		"a%2Fb":       true,
		"a%2fb":       true,
		"a%5Cb":       true,
		"a%5cb":       true,
		"trailing%2F": true,
		"%2F":         true,
		"%2":          false, // truncated, not a full sequence
		"":            false,
	} {
		assert.Equal(t, want, HasEncodedPathSeparator(s), "input=%q", s)
	}
}

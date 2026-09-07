// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echotest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type loadBytesOpts func([]byte) []byte

// TrimNewlineEnd instructs LoadBytes to remove `\n` from the end of loaded file.
func TrimNewlineEnd(bytes []byte) []byte {
	bLen := len(bytes)
	if bLen > 1 && bytes[bLen-1] == '\n' {
		bytes = bytes[:bLen-1]
	}
	return bytes
}

// LoadBytes is helper to load file contents relative to current (where test file is) package
// directory.
func LoadBytes(t *testing.T, name string, opts ...loadBytesOpts) []byte {
	bytes := loadBytes(t, name, 2)

	for _, f := range opts {
		bytes = f(bytes)
	}

	return bytes
}

func loadBytes(t *testing.T, name string, callDepth int) []byte {
	_, b, _, _ := runtime.Caller(callDepth)
	basepath := filepath.Dir(b)

	path := filepath.Join(basepath, name) // relative path
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes[:]
}

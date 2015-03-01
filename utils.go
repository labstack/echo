package bolt

import (
	"bytes"
	"io"
)

type nopCloser struct {
	*bytes.Buffer
}

func (nopCloser) Close() error {
	return nil
}

// NopCloser returns a ReadWriteCloser with a no-op Close method wrapping
// the provided Buffer.
func NopCloser(b *bytes.Buffer) io.ReadWriteCloser {
	return nopCloser{b}
}

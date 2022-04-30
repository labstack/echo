//go:build !go1.16
// +build !go1.16

package echo

type filesystem struct {
}

func createFilesystem() filesystem {
	return filesystem{}
}

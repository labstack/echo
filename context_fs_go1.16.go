//go:build go1.16
// +build go1.16

package echo

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
)

func (c *context) File(file string) error {
	return fsFile(c, file, c.echo.Filesystem)
}

func (c *context) FileFS(file string, filesystem fs.FS) error {
	return fsFile(c, file, filesystem)
}

func fsFile(c Context, file string, filesystem fs.FS) error {
	f, err := filesystem.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.ToSlash(filepath.Join(file, indexPage))
		f, err = filesystem.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return errors.New("file does not implement io.ReadSeeker")
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), ff)
	return nil
}

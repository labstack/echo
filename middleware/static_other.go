// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"io/fs"
	"os"
)

// We ignore these errors as there could be handler that matches request path.
func isIgnorableOpenFileError(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	// As of Go 1.20 Windows path checks are more strict on the provided path and considers [UNC](https://en.wikipedia.org/wiki/Path_(computing)#UNC)
	// paths with missing host etc parts as invalid. Previously it would result you `fs.ErrNotExist`.
	// Also `fs.Open` on all OSes does not accept ``, `.`, `..` at all.
	//
	// so we need to treat those errors the same as `fs.ErrNotExists` so we can continue handling
	// errors in the middleware/handler chain. Otherwise we might end up with status 500 instead of finding a route
	// or return 404 not found.
	var pErr *fs.PathError
	if errors.As(err, &pErr) {
		err = pErr.Err
		return err.Error() == "invalid argument"
	}
	return false
}

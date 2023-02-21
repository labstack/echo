package middleware

import (
	"os"
)

// We ignore these errors as there could be handler that matches request path.
//
// As of 1.20 on Windows filepath.Clean has different behaviour on OS related filesystems so we need to use path.Clean
// which is more suitable for path coming from web but this has some caveats on Windows. When we eventually end up in
// os related filesystem Open methods we are getting different errors as earlier versions. As of 1.20 path checks are
// more strict on path you provide and consider path with [UNC](https://en.wikipedia.org/wiki/Path_(computing)#UNC)
// but missing host etc parts as invalid. Previously it would result you `fs.ErrNotExist`.
//
// So for 1.20@Windows we need to consider it as same not exist so we can continue next middleware/handler and not error
// which would result status 500 instead of potential route hit or 404.
func isIgnorableOpenFileError(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	errTxt := err.Error()
	return errTxt == "http: invalid or unsafe file path" || errTxt == "invalid path"
}

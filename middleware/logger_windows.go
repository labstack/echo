// +build !appengine

package middleware

import (
	"github.com/mattn/go-colorable"
)

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}

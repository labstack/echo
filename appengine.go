// +build appengine

package echo

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

type (
	appengineLogAdapter struct {
		context.Context
	}
)

func init() {
	RegisterPreRequestHandlerFunc(setupAppEngineContext)
}

func setupAppEngineContext(c *Context) {
	c.Context = appengine.NewContext(c.Request())
}

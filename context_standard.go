// +build !appengine

package echo

import "github.com/labstack/echo/engine"

func (c *echoContext) Reset(req engine.Request, res engine.Response) {
	c.reset(req, res)
}

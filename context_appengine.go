// +build appengine

package echo

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
)

func (c *echoContext) Reset(req engine.Request, res engine.Response) {
	if r, ok := req.Context().Value(RequestContextKey).(*http.Request); ok {
		c.logger = log.New(log.WithContext(appengine.NewContext(r)))
		c.logger.SetLevel(c.echo.Logger().Level())
	}
	c.reset(req, res)
}

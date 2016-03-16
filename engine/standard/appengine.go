// +build appengine

package standard

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

// SetHandler implements `engine.Engine#SetHandler` method.
func (s *Server) SetHandler(h engine.Handler) {
	e := h.(*echo.Echo)
	e.SetContectFunc(func(req engine.Request) context.Context {
		return appengine.NewContext(req.(*Request).Request)
	})

	s.handler = h
}

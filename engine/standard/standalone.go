// +build !appengine

package standard

import (
	"github.com/labstack/echo/engine"
)

func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

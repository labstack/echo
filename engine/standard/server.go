package standard

import (
	"log"
	"net/http"

	"github.com/labstack/echo/engine"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
	}
)

func NewServer(config *engine.Config, handler engine.HandlerFunc) *Server {
	return &Server{
		Server:  new(http.Server),
		config:  config,
		handler: handler,
	}
}

func (s *Server) Start() {
	s.Addr = s.config.Address
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handler(NewRequest(r), NewResponse(w))
	})
	log.Fatal(s.ListenAndServe())
}

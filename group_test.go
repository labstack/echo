package echo

import "testing"

// TODO: Fix me
func TestGroup(t *testing.T) {
	g := New().Group("/group")
	h := func(Context) error { return nil }
	g.CONNECT("/", h)
	g.Connect("/", h)
	g.DELETE("/", h)
	g.Delete("/", h)
	g.GET("/", h)
	g.Get("/", h)
	g.HEAD("/", h)
	g.Head("/", h)
	g.OPTIONS("/", h)
	g.Options("/", h)
	g.PATCH("/", h)
	g.Patch("/", h)
	g.POST("/", h)
	g.Post("/", h)
	g.PUT("/", h)
	g.Put("/", h)
	g.TRACE("/", h)
	g.Trace("/", h)
	g.Any("/", h)
	g.Match([]string{GET, POST}, "/", h)
	g.Static("/static", "/tmp")
	g.File("/walle", "_fixture/images//walle.png")
}

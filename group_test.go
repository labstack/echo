package echo

import "testing"

func TestGroup(t *testing.T) {
	g := New().Group("/group")
	h := HandlerFunc(func(Context) error { return nil })
	g.Connect("/", h)
	g.Delete("/", h)
	g.Get("/", h)
	g.Head("/", h)
	g.Options("/", h)
	g.Patch("/", h)
	g.Post("/", h)
	g.Put("/", h)
	g.Trace("/", h)
}

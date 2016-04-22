package echo

import "testing"

func TestGroup(t *testing.T) {
	g := New().Group("/group")
	h := func(Context) error { return nil }
	g.CONNECT("/", h)
	g.DELETE("/", h)
	g.GET("/", h)
	g.HEAD("/", h)
	g.OPTIONS("/", h)
	g.PATCH("/", h)
	g.POST("/", h)
	g.PUT("/", h)
	g.TRACE("/", h)
	g.Any("/", h)
	g.Match([]string{GET, POST}, "/", h)
	g.Static("/static", "/tmp")
	g.File("/walle", "_fixture/images//walle.png")
}

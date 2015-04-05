package echo

import (
	"fmt"
	"testing"
)

func TestRouterStatic(t *testing.T) {
	r := New().Router
	r.Add(MethodGET, "/folders/files/echo.gif", func(c *Context) {}, nil)
	h, _, _ := r.Find(MethodGET, "/folders/files/echo.gif")
	if h == nil {
		t.Fatal("handle not found")
	}
}

func TestRouterParam(t *testing.T) {
	r := New().Router
	r.Add(MethodGET, "/users/:id", func(c *Context) {}, nil)
	h, c, _ := r.Find(MethodGET, "/users/1")
	if h == nil {
		t.Fatal("handle not found")
	}
	if c.P(0) != "1" {
		t.Error("param id should be 1")
	}
}

func TestRouterTwoParam(t *testing.T) {
	r := New().Router
	r.Add(MethodGET, "/users/:uid/files/:fid", func(c *Context) {}, nil)
	h, c, _ := r.Find(MethodGET, "/users/1/files/1")
	if h == nil {
		t.Fatal("handle not found")
	}
	if c.P(0) != "1" {
		t.Error("param uid should be 1")
	}
	if c.P(1) != "1" {
		t.Error("param fid should be 1")
	}
}

func TestRouterCatchAll(t *testing.T) {
	r := New().Router
	r.Add(MethodGET, "/static/*", func(c *Context) {}, nil)
	h, _, _ := r.Find(MethodGET, "/static/*")
	if h == nil {
		t.Fatal("handle not found")
	}
}

func TestRouterMicroParam(t *testing.T) {
	r := New().Router
	r.Add(MethodGET, "/:a/:b/:c", func(c *Context) {}, nil)
	h, c, _ := r.Find(MethodGET, "/1/2/3")
	if h == nil {
		t.Fatal("handle not found")
	}
	if c.P(0) != "1" {
		t.Error("param a should be 1")
	}
	if c.P(1) != "2" {
		t.Error("param b should be 2")
	}
	if c.P(2) != "3" {
		t.Error("param c should be 3")
	}
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s has=%d, h=%v, echo=%v\n", p, n.prefix, n.has, n.handler, n.echo)

	nodes := n.edges
	l := len(nodes)
	p = prefix(tail, pfx, "    ", "│   ")
	for i := 0; i < l-1; i++ {
		nodes[i].printTree(p, false)
	}
	if l > 0 {
		nodes[l-1].printTree(p, true)
	}
}

func prefix(tail bool, p, on, off string) string {
	if tail {
		return fmt.Sprintf("%s%s", p, on)
	}
	return fmt.Sprintf("%s%s", p, off)
}

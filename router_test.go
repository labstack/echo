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
	p := c.Param("id")
	if p != "1" {
		t.Errorf("id should be equal to 1, found %s", p)
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
	h, c, _ := r.Find(MethodGET, "/a/b/c")
	if h == nil {
		t.Fatal("handle not found")
	}
	p1 := c.P(0)
	if p1 != "a" {
		t.Errorf("p1 should be equal to a, found %s", p1)
	}
	p2 := c.P(1)
	if p2 != "b" {
		t.Errorf("p2 should be equal to b, found %s", p2)
	}
	p3 := c.P(2)
	if p3 != "c" {
		t.Errorf("p3 should be equal to c, found %s", p3)
	}
}

func TestRouterConflict(t *testing.T) {
	r := New().Router
	r.Add("GET", "/users/new", func(*Context) {}, nil)
	r.Add("GET", "/users/wen", func(*Context) {}, nil)
	r.Add("GET", "/users/:id", func(*Context) {}, nil)
	r.trees["GET"].printTree("", true)
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s has=%d, h=%v, echo=%d\n", p, n.prefix, n.has, n.handler, n.echo)

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

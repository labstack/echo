package echo

import "testing"

func TestRouterStatic(t *testing.T) {
	r := New().Router
	r.Add("GET", "/folders/files/bolt.gif", func(c *Context) {})
	h, _, _ := r.Find("GET", "/folders/files/bolt.gif")
	if h == nil {
		t.Fatal("handle not found")
	}
}

func TestRouterParam(t *testing.T) {
	r := New().Router
	r.Add("GET", "/users/:id", func(c *Context) {})
	h, c, _ := r.Find("GET", "/users/1")
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
	r.Add("GET", "/static/*", func(c *Context) {})
	h, _, _ := r.Find("GET", "/static/*")
	if h == nil {
		t.Fatal("handle not found")
	}
}

func TestRouterMicroParam(t *testing.T) {
	r := New().Router
	r.Add("GET", "/:a/:b/:c", func(c *Context) {})
	h, c, _ := r.Find("GET", "/a/b/c")
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

func TestPrintTree(t *testing.T) {
	r := New().Router
	r.Add("GET", "/users", nil)
	r.Add("GET", "/users/:id", nil)
	r.Add("GET", "/users/:id/books", nil)
	r.Add("GET", "/users/:id/files", nil)
	r.Add("POST", "/files", nil)
	r.printTree()
}

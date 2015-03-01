package bolt

import "testing"

func TestStatic(t *testing.T) {
	r := New().Router
	r.Add("GET", "/users/joe/books", func(c *Context) {})
	h, _, _ := r.Find("GET", "/users/joe/books")
	if h == nil {
		t.Fatal("handle not found")
	}
}

func TestParam(t *testing.T) {
	r := New().Router
	r.Add("GET", "/users/:name", func(c *Context) {})
	h, c, _ := r.Find("GET", "/users/joe")
	if h == nil {
		t.Fatal("handle not found")
	}
	p := c.Param("name")
	if p != "joe" {
		t.Errorf("name should be equal to joe, found %s", p)
	}
}

func TestMicroParam(t *testing.T) {
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

package echo

import "testing"

func TestGroupConnect(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, CONNECT, "/")
}

func TestGroupDelete(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, DELETE, "/")
}

func TestGroupGet(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, GET, "/")
}

func TestGroupHead(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, HEAD, "/")
}

func TestGroupOptions(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, OPTIONS, "/")
}

func TestGroupPatch(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, PATCH, "/")
}

func TestGroupPost(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, POST, "/")
}

func TestGroupPut(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, PUT, "/")
}

func TestGroupTrace(t *testing.T) {
	g := New().Group("/group")
	testMethod(t, &g.echo, g, TRACE, "/")
}

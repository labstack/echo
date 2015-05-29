package echo

import "testing"

func TestGroupConnect(t *testing.T) {
	g := New().Group("/group")
	testMethod(CONNECT, "/", &g.echo, g, t)
}

func TestGroupDelete(t *testing.T) {
	g := New().Group("/group")
	testMethod(DELETE, "/", &g.echo, g, t)
}

func TestGroupGet(t *testing.T) {
	g := New().Group("/group")
	testMethod(GET, "/", &g.echo, g, t)
}

func TestGroupHead(t *testing.T) {
	g := New().Group("/group")
	testMethod(HEAD, "/", &g.echo, g, t)
}

func TestGroupOptions(t *testing.T) {
	g := New().Group("/group")
	testMethod(OPTIONS, "/", &g.echo, g, t)
}

func TestGroupPatch(t *testing.T) {
	g := New().Group("/group")
	testMethod(PATCH, "/", &g.echo, g, t)
}

func TestGroupPost(t *testing.T) {
	g := New().Group("/group")
	testMethod(POST, "/", &g.echo, g, t)
}

func TestGroupPut(t *testing.T) {
	g := New().Group("/group")
	testMethod(PUT, "/", &g.echo, g, t)
}

func TestGroupTrace(t *testing.T) {
	g := New().Group("/group")
	testMethod(TRACE, "/", &g.echo, g, t)
}

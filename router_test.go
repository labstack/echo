package echo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type route struct {
	method string
	path   string
}

var (
	context = NewContext(nil, nil, New())
	api     = []route{
		// OAuth Authorizations
		{"GET", "/authorizations"},
		{"GET", "/authorizations/:id"},
		{"POST", "/authorizations"},
		//{"PUT", "/authorizations/clients/:client_id"},
		//{"PATCH", "/authorizations/:id"},
		{"DELETE", "/authorizations/:id"},
		{"GET", "/applications/:client_id/tokens/:access_token"},
		{"DELETE", "/applications/:client_id/tokens"},
		{"DELETE", "/applications/:client_id/tokens/:access_token"},

		// Activity
		{"GET", "/events"},
		{"GET", "/repos/:owner/:repo/events"},
		{"GET", "/networks/:owner/:repo/events"},
		{"GET", "/orgs/:org/events"},
		{"GET", "/users/:user/received_events"},
		{"GET", "/users/:user/received_events/public"},
		{"GET", "/users/:user/events"},
		{"GET", "/users/:user/events/public"},
		{"GET", "/users/:user/events/orgs/:org"},
		{"GET", "/feeds"},
		{"GET", "/notifications"},
		{"GET", "/repos/:owner/:repo/notifications"},
		{"PUT", "/notifications"},
		{"PUT", "/repos/:owner/:repo/notifications"},
		{"GET", "/notifications/threads/:id"},
		//{"PATCH", "/notifications/threads/:id"},
		{"GET", "/notifications/threads/:id/subscription"},
		{"PUT", "/notifications/threads/:id/subscription"},
		{"DELETE", "/notifications/threads/:id/subscription"},
		{"GET", "/repos/:owner/:repo/stargazers"},
		{"GET", "/users/:user/starred"},
		{"GET", "/user/starred"},
		{"GET", "/user/starred/:owner/:repo"},
		{"PUT", "/user/starred/:owner/:repo"},
		{"DELETE", "/user/starred/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/subscribers"},
		{"GET", "/users/:user/subscriptions"},
		{"GET", "/user/subscriptions"},
		{"GET", "/repos/:owner/:repo/subscription"},
		{"PUT", "/repos/:owner/:repo/subscription"},
		{"DELETE", "/repos/:owner/:repo/subscription"},
		{"GET", "/user/subscriptions/:owner/:repo"},
		{"PUT", "/user/subscriptions/:owner/:repo"},
		{"DELETE", "/user/subscriptions/:owner/:repo"},

		// Gists
		{"GET", "/users/:user/gists"},
		{"GET", "/gists"},
		//{"GET", "/gists/public"},
		//{"GET", "/gists/starred"},
		{"GET", "/gists/:id"},
		{"POST", "/gists"},
		//{"PATCH", "/gists/:id"},
		{"PUT", "/gists/:id/star"},
		{"DELETE", "/gists/:id/star"},
		{"GET", "/gists/:id/star"},
		{"POST", "/gists/:id/forks"},
		{"DELETE", "/gists/:id"},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
		{"POST", "/repos/:owner/:repo/git/blobs"},
		{"GET", "/repos/:owner/:repo/git/commits/:sha"},
		{"POST", "/repos/:owner/:repo/git/commits"},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
		{"GET", "/repos/:owner/:repo/git/refs"},
		{"POST", "/repos/:owner/:repo/git/refs"},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
		{"GET", "/repos/:owner/:repo/git/tags/:sha"},
		{"POST", "/repos/:owner/:repo/git/tags"},
		{"GET", "/repos/:owner/:repo/git/trees/:sha"},
		{"POST", "/repos/:owner/:repo/git/trees"},

		// Issues
		{"GET", "/issues"},
		{"GET", "/user/issues"},
		{"GET", "/orgs/:org/issues"},
		{"GET", "/repos/:owner/:repo/issues"},
		{"GET", "/repos/:owner/:repo/issues/:number"},
		{"POST", "/repos/:owner/:repo/issues"},
		//{"PATCH", "/repos/:owner/:repo/issues/:number"},
		{"GET", "/repos/:owner/:repo/assignees"},
		{"GET", "/repos/:owner/:repo/assignees/:assignee"},
		{"GET", "/repos/:owner/:repo/issues/:number/comments"},
		//{"GET", "/repos/:owner/:repo/issues/comments"},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
		{"POST", "/repos/:owner/:repo/issues/:number/comments"},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
		{"GET", "/repos/:owner/:repo/issues/:number/events"},
		//{"GET", "/repos/:owner/:repo/issues/events"},
		//{"GET", "/repos/:owner/:repo/issues/events/:id"},
		{"GET", "/repos/:owner/:repo/labels"},
		{"GET", "/repos/:owner/:repo/labels/:name"},
		{"POST", "/repos/:owner/:repo/labels"},
		//{"PATCH", "/repos/:owner/:repo/labels/:name"},
		{"DELETE", "/repos/:owner/:repo/labels/:name"},
		{"GET", "/repos/:owner/:repo/issues/:number/labels"},
		{"POST", "/repos/:owner/:repo/issues/:number/labels"},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
		{"GET", "/repos/:owner/:repo/milestones"},
		{"GET", "/repos/:owner/:repo/milestones/:number"},
		{"POST", "/repos/:owner/:repo/milestones"},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
		{"DELETE", "/repos/:owner/:repo/milestones/:number"},

		// Miscellaneous
		{"GET", "/emojis"},
		{"GET", "/gitignore/templates"},
		{"GET", "/gitignore/templates/:name"},
		{"POST", "/markdown"},
		{"POST", "/markdown/raw"},
		{"GET", "/meta"},
		{"GET", "/rate_limit"},

		// Organizations
		{"GET", "/users/:user/orgs"},
		{"GET", "/user/orgs"},
		{"GET", "/orgs/:org"},
		//{"PATCH", "/orgs/:org"},
		{"GET", "/orgs/:org/members"},
		{"GET", "/orgs/:org/members/:user"},
		{"DELETE", "/orgs/:org/members/:user"},
		{"GET", "/orgs/:org/public_members"},
		{"GET", "/orgs/:org/public_members/:user"},
		{"PUT", "/orgs/:org/public_members/:user"},
		{"DELETE", "/orgs/:org/public_members/:user"},
		{"GET", "/orgs/:org/teams"},
		{"GET", "/teams/:id"},
		{"POST", "/orgs/:org/teams"},
		//{"PATCH", "/teams/:id"},
		{"DELETE", "/teams/:id"},
		{"GET", "/teams/:id/members"},
		{"GET", "/teams/:id/members/:user"},
		{"PUT", "/teams/:id/members/:user"},
		{"DELETE", "/teams/:id/members/:user"},
		{"GET", "/teams/:id/repos"},
		{"GET", "/teams/:id/repos/:owner/:repo"},
		{"PUT", "/teams/:id/repos/:owner/:repo"},
		{"DELETE", "/teams/:id/repos/:owner/:repo"},
		{"GET", "/user/teams"},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls"},
		{"GET", "/repos/:owner/:repo/pulls/:number"},
		{"POST", "/repos/:owner/:repo/pulls"},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
		{"GET", "/repos/:owner/:repo/pulls/:number/files"},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
		//{"GET", "/repos/:owner/:repo/pulls/comments"},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

		// Repositories
		{"GET", "/user/repos"},
		{"GET", "/users/:user/repos"},
		{"GET", "/orgs/:org/repos"},
		{"GET", "/repositories"},
		{"POST", "/user/repos"},
		{"POST", "/orgs/:org/repos"},
		{"GET", "/repos/:owner/:repo"},
		//{"PATCH", "/repos/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/contributors"},
		{"GET", "/repos/:owner/:repo/languages"},
		{"GET", "/repos/:owner/:repo/teams"},
		{"GET", "/repos/:owner/:repo/tags"},
		{"GET", "/repos/:owner/:repo/branches"},
		{"GET", "/repos/:owner/:repo/branches/:branch"},
		{"DELETE", "/repos/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/collaborators"},
		{"GET", "/repos/:owner/:repo/collaborators/:user"},
		{"PUT", "/repos/:owner/:repo/collaborators/:user"},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
		{"GET", "/repos/:owner/:repo/comments"},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
		{"GET", "/repos/:owner/:repo/comments/:id"},
		//{"PATCH", "/repos/:owner/:repo/comments/:id"},
		{"DELETE", "/repos/:owner/:repo/comments/:id"},
		{"GET", "/repos/:owner/:repo/commits"},
		{"GET", "/repos/:owner/:repo/commits/:sha"},
		{"GET", "/repos/:owner/:repo/readme"},
		//{"GET", "/repos/:owner/:repo/contents/*path"},
		//{"PUT", "/repos/:owner/:repo/contents/*path"},
		//{"DELETE", "/repos/:owner/:repo/contents/*path"},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
		{"GET", "/repos/:owner/:repo/keys"},
		{"GET", "/repos/:owner/:repo/keys/:id"},
		{"POST", "/repos/:owner/:repo/keys"},
		//{"PATCH", "/repos/:owner/:repo/keys/:id"},
		{"DELETE", "/repos/:owner/:repo/keys/:id"},
		{"GET", "/repos/:owner/:repo/downloads"},
		{"GET", "/repos/:owner/:repo/downloads/:id"},
		{"DELETE", "/repos/:owner/:repo/downloads/:id"},
		{"GET", "/repos/:owner/:repo/forks"},
		{"POST", "/repos/:owner/:repo/forks"},
		{"GET", "/repos/:owner/:repo/hooks"},
		{"GET", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/hooks"},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
		{"DELETE", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/merges"},
		{"GET", "/repos/:owner/:repo/releases"},
		{"GET", "/repos/:owner/:repo/releases/:id"},
		{"POST", "/repos/:owner/:repo/releases"},
		//{"PATCH", "/repos/:owner/:repo/releases/:id"},
		{"DELETE", "/repos/:owner/:repo/releases/:id"},
		{"GET", "/repos/:owner/:repo/releases/:id/assets"},
		{"GET", "/repos/:owner/:repo/stats/contributors"},
		{"GET", "/repos/:owner/:repo/stats/commit_activity"},
		{"GET", "/repos/:owner/:repo/stats/code_frequency"},
		{"GET", "/repos/:owner/:repo/stats/participation"},
		{"GET", "/repos/:owner/:repo/stats/punch_card"},
		{"GET", "/repos/:owner/:repo/statuses/:ref"},
		{"POST", "/repos/:owner/:repo/statuses/:ref"},

		// Search
		{"GET", "/search/repositories"},
		{"GET", "/search/code"},
		{"GET", "/search/issues"},
		{"GET", "/search/users"},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
		{"GET", "/legacy/repos/search/:keyword"},
		{"GET", "/legacy/user/search/:keyword"},
		{"GET", "/legacy/user/email/:email"},

		// Users
		{"GET", "/users/:user"},
		{"GET", "/user"},
		//{"PATCH", "/user"},
		{"GET", "/users"},
		{"GET", "/user/emails"},
		{"POST", "/user/emails"},
		{"DELETE", "/user/emails"},
		{"GET", "/users/:user/followers"},
		{"GET", "/user/followers"},
		{"GET", "/users/:user/following"},
		{"GET", "/user/following"},
		{"GET", "/user/following/:user"},
		{"GET", "/users/:user/following/:target_user"},
		{"PUT", "/user/following/:user"},
		{"DELETE", "/user/following/:user"},
		{"GET", "/users/:user/keys"},
		{"GET", "/user/keys"},
		{"GET", "/user/keys/:id"},
		{"POST", "/user/keys"},
		//{"PATCH", "/user/keys/:id"},
		{"DELETE", "/user/keys/:id"},
	}
)

func TestRouterStatic(t *testing.T) {
	r := New().Router
	b := new(bytes.Buffer)
	path := "/folders/a/files/echo.gif"
	r.Add(GET, path, func(*Context) *HTTPError {
		b.WriteString(path)
		return nil
	}, nil)
	h, _ := r.Find(GET, path, context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(nil)
		if b.String() != path {
			t.Errorf("buffer should %s", path)
		}
	}
}

func TestRouterParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users/:id", func(c *Context) *HTTPError {
		return nil
	}, nil)
	h, _ := r.Find(GET, "/users/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.P(0) != "1" {
			t.Error("param id should be 1")
		}
	}
}

func TestRouterTwoParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users/:uid/files/:fid", func(*Context) *HTTPError {
		return nil
	}, nil)

	h, _ := r.Find(GET, "/users/1/files/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.P(0) != "1" {
			t.Error("param uid should be 1")
		}
		if context.P(1) != "1" {
			t.Error("param fid should be 1")
		}
	}

	h, _ = r.Find(GET, "/users/1", context)
	if h != nil {
		t.Error("should not found handler")
	}
}

func TestRouterMatchAny(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users/*", func(*Context) *HTTPError {
		return nil
	}, nil)

	h, _ := r.Find(GET, "/users/", context)
	if h == nil {
		t.Error("should match empty value")
	} else {
		if context.P(0) != "" {
			t.Error("value should be empty")
		}
	}

	h, _ = r.Find(GET, "/users/joe", context)
	if h == nil {
		t.Error("should match non-empty value")
	} else {
		if context.P(0) != "joe" {
			t.Error("value should be joe")
		}
	}
}

func TestRouterMicroParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/:a/:b/:c", func(c *Context) *HTTPError {
		return nil
	}, nil)
	h, _ := r.Find(GET, "/1/2/3", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.P(0) != "1" {
			t.Error("param a should be 1")
		}
		if context.P(1) != "2" {
			t.Error("param b should be 2")
		}
		if context.P(2) != "3" {
			t.Error("param c should be 3")
		}
	}
}

func TestRouterMultiRoute(t *testing.T) {
	r := New().Router
	b := new(bytes.Buffer)

	// Routes
	r.Add(GET, "/users", func(*Context) *HTTPError {
		b.WriteString("/users")
		return nil
	}, nil)
	r.Add(GET, "/users/:id", func(c *Context) *HTTPError {
		return nil
	}, nil)

	// Route > /users
	h, _ := r.Find(GET, "/users", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(nil)
		if b.String() != "/users" {
			t.Errorf("buffer should be /users")
		}
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.P(0) != "1" {
			t.Error("param id should be 1")
		}
	}

	// Route > /user
	h, _ = r.Find(GET, "/user", context)
	if h != nil {
		t.Error("handler should be nil")
	}
}

func TestRouterPriority(t *testing.T) {
	r := New().Router

	// Routes
	r.Add(GET, "/users", func(c *Context) *HTTPError {
		c.Set("a", 1)
		return nil
	}, nil)
	r.Add(GET, "/users/new", func(c *Context) *HTTPError {
		c.Set("b", 2)
		return nil
	}, nil)
	r.Add(GET, "/users/:id", func(c *Context) *HTTPError {
		c.Set("c", 3)
		return nil
	}, nil)
	r.Add(GET, "/users/dew", func(c *Context) *HTTPError {
		c.Set("d", 4)
		return nil
	}, nil)
	r.Add(GET, "/users/:id/files", func(c *Context) *HTTPError {
		c.Set("e", 5)
		return nil
	}, nil)
	r.Add(GET, "/users/newsee", func(c *Context) *HTTPError {
		c.Set("f", 6)
		return nil
	}, nil)
	r.Add(GET, "/users/*", func(c *Context) *HTTPError {
		c.Set("g", 7)
		return nil
	}, nil)

	// Route > /users
	h, _ := r.Find(GET, "/users", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("a") != 1 {
			t.Error("a should map to 1")
		}
	}

	// Route > /users/new
	h, _ = r.Find(GET, "/users/new", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("b") != 2 {
			t.Error("b should map to 2")
		}
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("c") != 3 {
			t.Error("c should map to 3")
		}
	}

	// Route > /users/dew
	h, _ = r.Find(GET, "/users/dew", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("d") != 4 {
			t.Error("d should map to 4")
		}
	}

	// Route > /users/:id/files
	h, _ = r.Find(GET, "/users/1/files", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("e") != 5 {
			t.Error("e should map to 5")
		}
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/news", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("c") != 3 {
			t.Error("c should map to 3")
		}
	}

	// Route > /users/*
	h, _ = r.Find(GET, "/users/joe/books", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if context.Get("g") != 7 {
			t.Error("g should map to 7")
		}
	}
}

func TestRouterParamNames(t *testing.T) {
	r := New().Router
	b := new(bytes.Buffer)

	// Routes
	r.Add(GET, "/users", func(*Context) *HTTPError {
		b.WriteString("/users")
		return nil
	}, nil)
	r.Add(GET, "/users/:id", func(c *Context) *HTTPError {
		return nil
	}, nil)
	r.Add(GET, "/users/:uid/files/:fid", func(c *Context) *HTTPError {
		return nil
	}, nil)

	// Route > /users
	h, _ := r.Find(GET, "/users", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		h(context)
		if b.String() != "/users" {
			t.Errorf("buffer should be /users")
		}
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.pnames[0] != "id" {
			t.Error("param name should be id")
		}
		if context.P(0) != "1" {
			t.Error("param id should be 1")
		}
	}

	// Route > /users/:uid/files/:fid
	h, _ = r.Find(GET, "/users/1/files/1", context)
	if h == nil {
		t.Error("handler not found")
	} else {
		if context.pnames[0] != "uid" {
			t.Error("param name should be id")
		}
		if context.P(0) != "1" {
			t.Error("param id should be 1")
		}
		if context.pnames[1] != "fid" {
			t.Error("param name should be id")
		}
		if context.P(1) != "1" {
			t.Error("param id should be 1")
		}
	}
}

func TestRouterAPI(t *testing.T) {
	r := New().Router
	for _, route := range api {
		r.Add(route.method, route.path, func(c *Context) *HTTPError {
			for i, n := range c.pnames {
				if n != "" {
					if ":"+n != c.P(uint8(i)) {
						t.Errorf("param not found, method=%s, path=%s", route.method, route.path)
					}
				}
			}
			return nil
		}, nil)
		h, _ := r.Find(route.method, route.path, context)
		if h == nil {
			t.Fatalf("handler not found, method=%s, path=%s", route.method, route.path)
		} else {
			h(context)
		}
	}
}

func TestRouterServeHTTP(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users", func(*Context) *HTTPError {
		return nil
	}, nil)

	// OK
	req, _ := http.NewRequest(GET, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// NotFound handler
	req, _ = http.NewRequest(GET, "/files", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s, %p: type=%d, parent=%p, handler=%v\n", p, n.prefix, n, n.typ, n.parent, n.handler)

	children := n.children
	l := len(children)
	p = prefix(tail, pfx, "    ", "│   ")
	for i := 0; i < l-1; i++ {
		children[i].printTree(p, false)
	}
	if l > 0 {
		children[l-1].printTree(p, true)
	}
}

func prefix(tail bool, p, on, off string) string {
	if tail {
		return fmt.Sprintf("%s%s", p, on)
	}
	return fmt.Sprintf("%s%s", p, off)
}

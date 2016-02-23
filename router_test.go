package echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	api = []Route{
		// OAuth Authorizations
		{"GET", "/authorizations", nil},
		{"GET", "/authorizations/:id", nil},
		{"POST", "/authorizations", nil},
		//{"PUT", "/authorizations/clients/:client_id", nil},
		//{"PATCH", "/authorizations/:id", nil},
		{"DELETE", "/authorizations/:id", nil},
		{"GET", "/applications/:client_id/tokens/:access_token", nil},
		{"DELETE", "/applications/:client_id/tokens", nil},
		{"DELETE", "/applications/:client_id/tokens/:access_token", nil},

		// Activity
		{"GET", "/events", nil},
		{"GET", "/repos/:owner/:repo/events", nil},
		{"GET", "/networks/:owner/:repo/events", nil},
		{"GET", "/orgs/:org/events", nil},
		{"GET", "/users/:user/received_events", nil},
		{"GET", "/users/:user/received_events/public", nil},
		{"GET", "/users/:user/events", nil},
		{"GET", "/users/:user/events/public", nil},
		{"GET", "/users/:user/events/orgs/:org", nil},
		{"GET", "/feeds", nil},
		{"GET", "/notifications", nil},
		{"GET", "/repos/:owner/:repo/notifications", nil},
		{"PUT", "/notifications", nil},
		{"PUT", "/repos/:owner/:repo/notifications", nil},
		{"GET", "/notifications/threads/:id", nil},
		//{"PATCH", "/notifications/threads/:id", nil},
		{"GET", "/notifications/threads/:id/subscription", nil},
		{"PUT", "/notifications/threads/:id/subscription", nil},
		{"DELETE", "/notifications/threads/:id/subscription", nil},
		{"GET", "/repos/:owner/:repo/stargazers", nil},
		{"GET", "/users/:user/starred", nil},
		{"GET", "/user/starred", nil},
		{"GET", "/user/starred/:owner/:repo", nil},
		{"PUT", "/user/starred/:owner/:repo", nil},
		{"DELETE", "/user/starred/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/subscribers", nil},
		{"GET", "/users/:user/subscriptions", nil},
		{"GET", "/user/subscriptions", nil},
		{"GET", "/repos/:owner/:repo/subscription", nil},
		{"PUT", "/repos/:owner/:repo/subscription", nil},
		{"DELETE", "/repos/:owner/:repo/subscription", nil},
		{"GET", "/user/subscriptions/:owner/:repo", nil},
		{"PUT", "/user/subscriptions/:owner/:repo", nil},
		{"DELETE", "/user/subscriptions/:owner/:repo", nil},

		// Gists
		{"GET", "/users/:user/gists", nil},
		{"GET", "/gists", nil},
		//{"GET", "/gists/public", nil},
		//{"GET", "/gists/starred", nil},
		{"GET", "/gists/:id", nil},
		{"POST", "/gists", nil},
		//{"PATCH", "/gists/:id", nil},
		{"PUT", "/gists/:id/star", nil},
		{"DELETE", "/gists/:id/star", nil},
		{"GET", "/gists/:id/star", nil},
		{"POST", "/gists/:id/forks", nil},
		{"DELETE", "/gists/:id", nil},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/blobs", nil},
		{"GET", "/repos/:owner/:repo/git/commits/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/commits", nil},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/refs", nil},
		{"POST", "/repos/:owner/:repo/git/refs", nil},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref", nil},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/tags/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/tags", nil},
		{"GET", "/repos/:owner/:repo/git/trees/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/trees", nil},

		// Issues
		{"GET", "/issues", nil},
		{"GET", "/user/issues", nil},
		{"GET", "/orgs/:org/issues", nil},
		{"GET", "/repos/:owner/:repo/issues", nil},
		{"GET", "/repos/:owner/:repo/issues/:number", nil},
		{"POST", "/repos/:owner/:repo/issues", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/:number", nil},
		{"GET", "/repos/:owner/:repo/assignees", nil},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/comments", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id", nil},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/events", nil},
		//{"GET", "/repos/:owner/:repo/issues/events", nil},
		//{"GET", "/repos/:owner/:repo/issues/events/:id", nil},
		{"GET", "/repos/:owner/:repo/labels", nil},
		{"GET", "/repos/:owner/:repo/labels/:name", nil},
		{"POST", "/repos/:owner/:repo/labels", nil},
		//{"PATCH", "/repos/:owner/:repo/labels/:name", nil},
		{"DELETE", "/repos/:owner/:repo/labels/:name", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", nil},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number", nil},
		{"POST", "/repos/:owner/:repo/milestones", nil},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number", nil},
		{"DELETE", "/repos/:owner/:repo/milestones/:number", nil},

		// Miscellaneous
		{"GET", "/emojis", nil},
		{"GET", "/gitignore/templates", nil},
		{"GET", "/gitignore/templates/:name", nil},
		{"POST", "/markdown", nil},
		{"POST", "/markdown/raw", nil},
		{"GET", "/meta", nil},
		{"GET", "/rate_limit", nil},

		// Organizations
		{"GET", "/users/:user/orgs", nil},
		{"GET", "/user/orgs", nil},
		{"GET", "/orgs/:org", nil},
		//{"PATCH", "/orgs/:org", nil},
		{"GET", "/orgs/:org/members", nil},
		{"GET", "/orgs/:org/members/:user", nil},
		{"DELETE", "/orgs/:org/members/:user", nil},
		{"GET", "/orgs/:org/public_members", nil},
		{"GET", "/orgs/:org/public_members/:user", nil},
		{"PUT", "/orgs/:org/public_members/:user", nil},
		{"DELETE", "/orgs/:org/public_members/:user", nil},
		{"GET", "/orgs/:org/teams", nil},
		{"GET", "/teams/:id", nil},
		{"POST", "/orgs/:org/teams", nil},
		//{"PATCH", "/teams/:id", nil},
		{"DELETE", "/teams/:id", nil},
		{"GET", "/teams/:id/members", nil},
		{"GET", "/teams/:id/members/:user", nil},
		{"PUT", "/teams/:id/members/:user", nil},
		{"DELETE", "/teams/:id/members/:user", nil},
		{"GET", "/teams/:id/repos", nil},
		{"GET", "/teams/:id/repos/:owner/:repo", nil},
		{"PUT", "/teams/:id/repos/:owner/:repo", nil},
		{"DELETE", "/teams/:id/repos/:owner/:repo", nil},
		{"GET", "/user/teams", nil},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number", nil},
		{"POST", "/repos/:owner/:repo/pulls", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", nil},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", nil},

		// Repositories
		{"GET", "/user/repos", nil},
		{"GET", "/users/:user/repos", nil},
		{"GET", "/orgs/:org/repos", nil},
		{"GET", "/repositories", nil},
		{"POST", "/user/repos", nil},
		{"POST", "/orgs/:org/repos", nil},
		{"GET", "/repos/:owner/:repo", nil},
		//{"PATCH", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/contributors", nil},
		{"GET", "/repos/:owner/:repo/languages", nil},
		{"GET", "/repos/:owner/:repo/teams", nil},
		{"GET", "/repos/:owner/:repo/tags", nil},
		{"GET", "/repos/:owner/:repo/branches", nil},
		{"GET", "/repos/:owner/:repo/branches/:branch", nil},
		{"DELETE", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/collaborators", nil},
		{"GET", "/repos/:owner/:repo/collaborators/:user", nil},
		{"PUT", "/repos/:owner/:repo/collaborators/:user", nil},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user", nil},
		{"GET", "/repos/:owner/:repo/comments", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"GET", "/repos/:owner/:repo/comments/:id", nil},
		//{"PATCH", "/repos/:owner/:repo/comments/:id", nil},
		{"DELETE", "/repos/:owner/:repo/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/commits", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha", nil},
		{"GET", "/repos/:owner/:repo/readme", nil},
		//{"GET", "/repos/:owner/:repo/contents/*path", nil},
		//{"PUT", "/repos/:owner/:repo/contents/*path", nil},
		//{"DELETE", "/repos/:owner/:repo/contents/*path", nil},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref", nil},
		{"GET", "/repos/:owner/:repo/keys", nil},
		{"GET", "/repos/:owner/:repo/keys/:id", nil},
		{"POST", "/repos/:owner/:repo/keys", nil},
		//{"PATCH", "/repos/:owner/:repo/keys/:id", nil},
		{"DELETE", "/repos/:owner/:repo/keys/:id", nil},
		{"GET", "/repos/:owner/:repo/downloads", nil},
		{"GET", "/repos/:owner/:repo/downloads/:id", nil},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", nil},
		{"GET", "/repos/:owner/:repo/forks", nil},
		{"POST", "/repos/:owner/:repo/forks", nil},
		{"GET", "/repos/:owner/:repo/hooks", nil},
		{"GET", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks", nil},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests", nil},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/merges", nil},
		{"GET", "/repos/:owner/:repo/releases", nil},
		{"GET", "/repos/:owner/:repo/releases/:id", nil},
		{"POST", "/repos/:owner/:repo/releases", nil},
		//{"PATCH", "/repos/:owner/:repo/releases/:id", nil},
		{"DELETE", "/repos/:owner/:repo/releases/:id", nil},
		{"GET", "/repos/:owner/:repo/releases/:id/assets", nil},
		{"GET", "/repos/:owner/:repo/stats/contributors", nil},
		{"GET", "/repos/:owner/:repo/stats/commit_activity", nil},
		{"GET", "/repos/:owner/:repo/stats/code_frequency", nil},
		{"GET", "/repos/:owner/:repo/stats/participation", nil},
		{"GET", "/repos/:owner/:repo/stats/punch_card", nil},
		{"GET", "/repos/:owner/:repo/statuses/:ref", nil},
		{"POST", "/repos/:owner/:repo/statuses/:ref", nil},

		// Search
		{"GET", "/search/repositories", nil},
		{"GET", "/search/code", nil},
		{"GET", "/search/issues", nil},
		{"GET", "/search/users", nil},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", nil},
		{"GET", "/legacy/repos/search/:keyword", nil},
		{"GET", "/legacy/user/search/:keyword", nil},
		{"GET", "/legacy/user/email/:email", nil},

		// Users
		{"GET", "/users/:user", nil},
		{"GET", "/user", nil},
		//{"PATCH", "/user", nil},
		{"GET", "/users", nil},
		{"GET", "/user/emails", nil},
		{"POST", "/user/emails", nil},
		{"DELETE", "/user/emails", nil},
		{"GET", "/users/:user/followers", nil},
		{"GET", "/user/followers", nil},
		{"GET", "/users/:user/following", nil},
		{"GET", "/user/following", nil},
		{"GET", "/user/following/:user", nil},
		{"GET", "/users/:user/following/:target_user", nil},
		{"PUT", "/user/following/:user", nil},
		{"DELETE", "/user/following/:user", nil},
		{"GET", "/users/:user/keys", nil},
		{"GET", "/user/keys", nil},
		{"GET", "/user/keys/:id", nil},
		{"POST", "/user/keys", nil},
		//{"PATCH", "/user/keys/:id", nil},
		{"DELETE", "/user/keys/:id", nil},
	}
)

func TestRouterStatic(t *testing.T) {
	e := New()
	r := e.router
	path := "/folders/a/files/echo.gif"
	r.Add(GET, path, func(c *Context) error {
		c.Set("path", path)
		return nil
	}, e)
	c := NewContext(nil, nil, e)
	h, _ := r.Find(GET, path, c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, path, c.Get("path"))
	}
}

func TestRouterParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/users/:id", func(c *Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)
	h, _ := r.Find(GET, "/users/1", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", c.P(0))
	}
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/users/:uid/files/:fid", func(*Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	h, _ := r.Find(GET, "/users/1/files/1", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", c.P(0))
		assert.Equal(t, "1", c.P(1))
	}
}

func TestRouterMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/", func(*Context) error {
		return nil
	}, e)
	r.Add(GET, "/*", func(*Context) error {
		return nil
	}, e)
	r.Add(GET, "/users/*", func(*Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	h, _ := r.Find(GET, "/", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "", c.P(0))
	}

	h, _ = r.Find(GET, "/download", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "download", c.P(0))
	}

	h, _ = r.Find(GET, "/users/joe", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "joe", c.P(0))
	}
}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/:a/:b/:c", func(c *Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)
	h, _ := r.Find(GET, "/1/2/3", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", c.P(0))
		assert.Equal(t, "2", c.P(1))
		assert.Equal(t, "3", c.P(2))
	}
}

func TestRouterMixParamMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Route
	r.Add(GET, "/users/:id/*", func(c *Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	h, _ := r.Find(GET, "/users/joe/comments", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, "joe", c.P(0))
	}
}

func TestRouterMultiRoute(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c *Context) error {
		c.Set("path", "/users")
		return nil
	}, e)
	r.Add(GET, "/users/:id", func(c *Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	// Route > /users
	h, _ := r.Find(GET, "/users", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, "/users", c.Get("path"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", c.P(0))
	}

	// Route > /user
	h, _ = r.Find(GET, "/user", c)
	if assert.IsType(t, new(HTTPError), h(c)) {
		he := h(c).(*HTTPError)
		assert.Equal(t, http.StatusNotFound, he.code)
	}
}

func TestRouterPriority(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c *Context) error {
		c.Set("a", 1)
		return nil
	}, e)
	r.Add(GET, "/users/new", func(c *Context) error {
		c.Set("b", 2)
		return nil
	}, e)
	r.Add(GET, "/users/:id", func(c *Context) error {
		c.Set("c", 3)
		return nil
	}, e)
	r.Add(GET, "/users/dew", func(c *Context) error {
		c.Set("d", 4)
		return nil
	}, e)
	r.Add(GET, "/users/:id/files", func(c *Context) error {
		c.Set("e", 5)
		return nil
	}, e)
	r.Add(GET, "/users/newsee", func(c *Context) error {
		c.Set("f", 6)
		return nil
	}, e)
	r.Add(GET, "/users/*", func(c *Context) error {
		c.Set("g", 7)
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	// Route > /users
	h, _ := r.Find(GET, "/users", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 1, c.Get("a"))
	}

	// Route > /users/new
	h, _ = r.Find(GET, "/users/new", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 2, c.Get("b"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 3, c.Get("c"))
	}

	// Route > /users/dew
	h, _ = r.Find(GET, "/users/dew", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 4, c.Get("d"))
	}

	// Route > /users/:id/files
	h, _ = r.Find(GET, "/users/1/files", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 5, c.Get("e"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/news", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 3, c.Get("c"))
	}

	// Route > /users/*
	h, _ = r.Find(GET, "/users/joe/books", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 7, c.Get("g"))
		assert.Equal(t, "joe/books", c.Param("_*"))
	}
}

func TestRouterParamNames(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c *Context) error {
		c.Set("path", "/users")
		return nil
	}, e)
	r.Add(GET, "/users/:id", func(c *Context) error {
		return nil
	}, e)
	r.Add(GET, "/users/:uid/files/:fid", func(c *Context) error {
		return nil
	}, e)
	c := NewContext(nil, nil, e)

	// Route > /users
	h, _ := r.Find(GET, "/users", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, "/users", c.Get("path"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "id", c.pnames[0])
		assert.Equal(t, "1", c.P(0))
	}

	// Route > /users/:uid/files/:fid
	h, _ = r.Find(GET, "/users/1/files/1", c)
	if assert.NotNil(t, h) {
		assert.Equal(t, "uid", c.pnames[0])
		assert.Equal(t, "1", c.P(0))
		assert.Equal(t, "fid", c.pnames[1])
		assert.Equal(t, "1", c.P(1))
	}
}

func TestRouterAPI(t *testing.T) {
	e := New()
	r := e.router

	for _, route := range api {
		r.Add(route.Method, route.Path, func(c *Context) error {
			return nil
		}, e)
	}
	c := NewContext(nil, nil, e)
	for _, route := range api {
		h, _ := r.Find(route.Method, route.Path, c)
		if assert.NotNil(t, h) {
			for i, n := range c.pnames {
				if assert.NotEmpty(t, n) {
					assert.Equal(t, ":"+n, c.P(i))
				}
			}
			h(c)
		}
	}
}

func TestRouterServeHTTP(t *testing.T) {
	e := New()
	r := e.router

	r.Add(GET, "/users", func(*Context) error {
		return nil
	}, e)

	// OK
	req, _ := http.NewRequest(GET, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Not found
	req, _ = http.NewRequest(GET, "/files", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s, %p: type=%d, parent=%p, handler=%v\n", p, n.prefix, n, n.kind, n.parent, n.methodHandler)

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

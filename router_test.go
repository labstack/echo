package echo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	api = []Route{
		// OAuth Authorizations
		{"GET", "/authorizations", ""},
		{"GET", "/authorizations/:id", ""},
		{"POST", "/authorizations", ""},
		//{"PUT", "/authorizations/clients/:client_id", ""},
		//{"PATCH", "/authorizations/:id", ""},
		{"DELETE", "/authorizations/:id", ""},
		{"GET", "/applications/:client_id/tokens/:access_token", ""},
		{"DELETE", "/applications/:client_id/tokens", ""},
		{"DELETE", "/applications/:client_id/tokens/:access_token", ""},

		// Activity
		{"GET", "/events", ""},
		{"GET", "/repos/:owner/:repo/events", ""},
		{"GET", "/networks/:owner/:repo/events", ""},
		{"GET", "/orgs/:org/events", ""},
		{"GET", "/users/:user/received_events", ""},
		{"GET", "/users/:user/received_events/public", ""},
		{"GET", "/users/:user/events", ""},
		{"GET", "/users/:user/events/public", ""},
		{"GET", "/users/:user/events/orgs/:org", ""},
		{"GET", "/feeds", ""},
		{"GET", "/notifications", ""},
		{"GET", "/repos/:owner/:repo/notifications", ""},
		{"PUT", "/notifications", ""},
		{"PUT", "/repos/:owner/:repo/notifications", ""},
		{"GET", "/notifications/threads/:id", ""},
		//{"PATCH", "/notifications/threads/:id", ""},
		{"GET", "/notifications/threads/:id/subscription", ""},
		{"PUT", "/notifications/threads/:id/subscription", ""},
		{"DELETE", "/notifications/threads/:id/subscription", ""},
		{"GET", "/repos/:owner/:repo/stargazers", ""},
		{"GET", "/users/:user/starred", ""},
		{"GET", "/user/starred", ""},
		{"GET", "/user/starred/:owner/:repo", ""},
		{"PUT", "/user/starred/:owner/:repo", ""},
		{"DELETE", "/user/starred/:owner/:repo", ""},
		{"GET", "/repos/:owner/:repo/subscribers", ""},
		{"GET", "/users/:user/subscriptions", ""},
		{"GET", "/user/subscriptions", ""},
		{"GET", "/repos/:owner/:repo/subscription", ""},
		{"PUT", "/repos/:owner/:repo/subscription", ""},
		{"DELETE", "/repos/:owner/:repo/subscription", ""},
		{"GET", "/user/subscriptions/:owner/:repo", ""},
		{"PUT", "/user/subscriptions/:owner/:repo", ""},
		{"DELETE", "/user/subscriptions/:owner/:repo", ""},

		// Gists
		{"GET", "/users/:user/gists", ""},
		{"GET", "/gists", ""},
		//{"GET", "/gists/public", ""},
		//{"GET", "/gists/starred", ""},
		{"GET", "/gists/:id", ""},
		{"POST", "/gists", ""},
		//{"PATCH", "/gists/:id", ""},
		{"PUT", "/gists/:id/star", ""},
		{"DELETE", "/gists/:id/star", ""},
		{"GET", "/gists/:id/star", ""},
		{"POST", "/gists/:id/forks", ""},
		{"DELETE", "/gists/:id", ""},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha", ""},
		{"POST", "/repos/:owner/:repo/git/blobs", ""},
		{"GET", "/repos/:owner/:repo/git/commits/:sha", ""},
		{"POST", "/repos/:owner/:repo/git/commits", ""},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref", ""},
		{"GET", "/repos/:owner/:repo/git/refs", ""},
		{"POST", "/repos/:owner/:repo/git/refs", ""},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref", ""},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref", ""},
		{"GET", "/repos/:owner/:repo/git/tags/:sha", ""},
		{"POST", "/repos/:owner/:repo/git/tags", ""},
		{"GET", "/repos/:owner/:repo/git/trees/:sha", ""},
		{"POST", "/repos/:owner/:repo/git/trees", ""},

		// Issues
		{"GET", "/issues", ""},
		{"GET", "/user/issues", ""},
		{"GET", "/orgs/:org/issues", ""},
		{"GET", "/repos/:owner/:repo/issues", ""},
		{"GET", "/repos/:owner/:repo/issues/:number", ""},
		{"POST", "/repos/:owner/:repo/issues", ""},
		//{"PATCH", "/repos/:owner/:repo/issues/:number", ""},
		{"GET", "/repos/:owner/:repo/assignees", ""},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", ""},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", ""},
		//{"GET", "/repos/:owner/:repo/issues/comments", ""},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id", ""},
		{"POST", "/repos/:owner/:repo/issues/:number/comments", ""},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id", ""},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id", ""},
		{"GET", "/repos/:owner/:repo/issues/:number/events", ""},
		//{"GET", "/repos/:owner/:repo/issues/events", ""},
		//{"GET", "/repos/:owner/:repo/issues/events/:id", ""},
		{"GET", "/repos/:owner/:repo/labels", ""},
		{"GET", "/repos/:owner/:repo/labels/:name", ""},
		{"POST", "/repos/:owner/:repo/labels", ""},
		//{"PATCH", "/repos/:owner/:repo/labels/:name", ""},
		{"DELETE", "/repos/:owner/:repo/labels/:name", ""},
		{"GET", "/repos/:owner/:repo/issues/:number/labels", ""},
		{"POST", "/repos/:owner/:repo/issues/:number/labels", ""},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", ""},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels", ""},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels", ""},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels", ""},
		{"GET", "/repos/:owner/:repo/milestones", ""},
		{"GET", "/repos/:owner/:repo/milestones/:number", ""},
		{"POST", "/repos/:owner/:repo/milestones", ""},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number", ""},
		{"DELETE", "/repos/:owner/:repo/milestones/:number", ""},

		// Miscellaneous
		{"GET", "/emojis", ""},
		{"GET", "/gitignore/templates", ""},
		{"GET", "/gitignore/templates/:name", ""},
		{"POST", "/markdown", ""},
		{"POST", "/markdown/raw", ""},
		{"GET", "/meta", ""},
		{"GET", "/rate_limit", ""},

		// Organizations
		{"GET", "/users/:user/orgs", ""},
		{"GET", "/user/orgs", ""},
		{"GET", "/orgs/:org", ""},
		//{"PATCH", "/orgs/:org", ""},
		{"GET", "/orgs/:org/members", ""},
		{"GET", "/orgs/:org/members/:user", ""},
		{"DELETE", "/orgs/:org/members/:user", ""},
		{"GET", "/orgs/:org/public_members", ""},
		{"GET", "/orgs/:org/public_members/:user", ""},
		{"PUT", "/orgs/:org/public_members/:user", ""},
		{"DELETE", "/orgs/:org/public_members/:user", ""},
		{"GET", "/orgs/:org/teams", ""},
		{"GET", "/teams/:id", ""},
		{"POST", "/orgs/:org/teams", ""},
		//{"PATCH", "/teams/:id", ""},
		{"DELETE", "/teams/:id", ""},
		{"GET", "/teams/:id/members", ""},
		{"GET", "/teams/:id/members/:user", ""},
		{"PUT", "/teams/:id/members/:user", ""},
		{"DELETE", "/teams/:id/members/:user", ""},
		{"GET", "/teams/:id/repos", ""},
		{"GET", "/teams/:id/repos/:owner/:repo", ""},
		{"PUT", "/teams/:id/repos/:owner/:repo", ""},
		{"DELETE", "/teams/:id/repos/:owner/:repo", ""},
		{"GET", "/user/teams", ""},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number", ""},
		{"POST", "/repos/:owner/:repo/pulls", ""},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", ""},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", ""},
		//{"GET", "/repos/:owner/:repo/pulls/comments", ""},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number", ""},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", ""},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", ""},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", ""},

		// Repositories
		{"GET", "/user/repos", ""},
		{"GET", "/users/:user/repos", ""},
		{"GET", "/orgs/:org/repos", ""},
		{"GET", "/repositories", ""},
		{"POST", "/user/repos", ""},
		{"POST", "/orgs/:org/repos", ""},
		{"GET", "/repos/:owner/:repo", ""},
		//{"PATCH", "/repos/:owner/:repo", ""},
		{"GET", "/repos/:owner/:repo/contributors", ""},
		{"GET", "/repos/:owner/:repo/languages", ""},
		{"GET", "/repos/:owner/:repo/teams", ""},
		{"GET", "/repos/:owner/:repo/tags", ""},
		{"GET", "/repos/:owner/:repo/branches", ""},
		{"GET", "/repos/:owner/:repo/branches/:branch", ""},
		{"DELETE", "/repos/:owner/:repo", ""},
		{"GET", "/repos/:owner/:repo/collaborators", ""},
		{"GET", "/repos/:owner/:repo/collaborators/:user", ""},
		{"PUT", "/repos/:owner/:repo/collaborators/:user", ""},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user", ""},
		{"GET", "/repos/:owner/:repo/comments", ""},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments", ""},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments", ""},
		{"GET", "/repos/:owner/:repo/comments/:id", ""},
		//{"PATCH", "/repos/:owner/:repo/comments/:id", ""},
		{"DELETE", "/repos/:owner/:repo/comments/:id", ""},
		{"GET", "/repos/:owner/:repo/commits", ""},
		{"GET", "/repos/:owner/:repo/commits/:sha", ""},
		{"GET", "/repos/:owner/:repo/readme", ""},
		//{"GET", "/repos/:owner/:repo/contents/*path", ""},
		//{"PUT", "/repos/:owner/:repo/contents/*path", ""},
		//{"DELETE", "/repos/:owner/:repo/contents/*path", ""},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref", ""},
		{"GET", "/repos/:owner/:repo/keys", ""},
		{"GET", "/repos/:owner/:repo/keys/:id", ""},
		{"POST", "/repos/:owner/:repo/keys", ""},
		//{"PATCH", "/repos/:owner/:repo/keys/:id", ""},
		{"DELETE", "/repos/:owner/:repo/keys/:id", ""},
		{"GET", "/repos/:owner/:repo/downloads", ""},
		{"GET", "/repos/:owner/:repo/downloads/:id", ""},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", ""},
		{"GET", "/repos/:owner/:repo/forks", ""},
		{"POST", "/repos/:owner/:repo/forks", ""},
		{"GET", "/repos/:owner/:repo/hooks", ""},
		{"GET", "/repos/:owner/:repo/hooks/:id", ""},
		{"POST", "/repos/:owner/:repo/hooks", ""},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id", ""},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests", ""},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", ""},
		{"POST", "/repos/:owner/:repo/merges", ""},
		{"GET", "/repos/:owner/:repo/releases", ""},
		{"GET", "/repos/:owner/:repo/releases/:id", ""},
		{"POST", "/repos/:owner/:repo/releases", ""},
		//{"PATCH", "/repos/:owner/:repo/releases/:id", ""},
		{"DELETE", "/repos/:owner/:repo/releases/:id", ""},
		{"GET", "/repos/:owner/:repo/releases/:id/assets", ""},
		{"GET", "/repos/:owner/:repo/stats/contributors", ""},
		{"GET", "/repos/:owner/:repo/stats/commit_activity", ""},
		{"GET", "/repos/:owner/:repo/stats/code_frequency", ""},
		{"GET", "/repos/:owner/:repo/stats/participation", ""},
		{"GET", "/repos/:owner/:repo/stats/punch_card", ""},
		{"GET", "/repos/:owner/:repo/statuses/:ref", ""},
		{"POST", "/repos/:owner/:repo/statuses/:ref", ""},

		// Search
		{"GET", "/search/repositories", ""},
		{"GET", "/search/code", ""},
		{"GET", "/search/issues", ""},
		{"GET", "/search/users", ""},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", ""},
		{"GET", "/legacy/repos/search/:keyword", ""},
		{"GET", "/legacy/user/search/:keyword", ""},
		{"GET", "/legacy/user/email/:email", ""},

		// Users
		{"GET", "/users/:user", ""},
		{"GET", "/user", ""},
		//{"PATCH", "/user", ""},
		{"GET", "/users", ""},
		{"GET", "/user/emails", ""},
		{"POST", "/user/emails", ""},
		{"DELETE", "/user/emails", ""},
		{"GET", "/users/:user/followers", ""},
		{"GET", "/user/followers", ""},
		{"GET", "/users/:user/following", ""},
		{"GET", "/user/following", ""},
		{"GET", "/user/following/:user", ""},
		{"GET", "/users/:user/following/:target_user", ""},
		{"PUT", "/user/following/:user", ""},
		{"DELETE", "/user/following/:user", ""},
		{"GET", "/users/:user/keys", ""},
		{"GET", "/user/keys", ""},
		{"GET", "/user/keys/:id", ""},
		{"POST", "/user/keys", ""},
		//{"PATCH", "/user/keys/:id", ""},
		{"DELETE", "/user/keys/:id", ""},
	}
)

func TestRouterStatic(t *testing.T) {
	e := New()
	r := e.router
	path := "/folders/a/files/echo.gif"
	r.Add(GET, path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(GET, path, c)
	c.handler(c)
	assert.Equal(t, path, c.Get("path"))
}

func TestRouterParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "1", c.P(0))
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "1", c.P(0))
	assert.Equal(t, "1", c.P(1))
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	e := New()
	r := e.router

	r.Add(GET, "/a/:b/c/d/:e", func(c Context) error {
		return nil
	})

	r.Add(GET, "/a/:b/c/:d/:f", func(c Context) error {
		return nil
	})

	c := e.NewContext(nil, nil).(*context)
	assert.NotPanics(t, func() {
		r.Find(GET, "/a/1/c/d/2/3", c)
	})
}

func TestRouterMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/", func(Context) error {
		return nil
	})
	r.Add(GET, "/*", func(Context) error {
		return nil
	})
	r.Add(GET, "/users/*", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(GET, "/", c)
	assert.Equal(t, "", c.P(0))

	r.Find(GET, "/download", c)
	assert.Equal(t, "download", c.P(0))

	r.Find(GET, "/users/joe", c)
	assert.Equal(t, "joe", c.P(0))
}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(GET, "/1/2/3", c)
	assert.Equal(t, "1", c.P(0))
	assert.Equal(t, "2", c.P(1))
	assert.Equal(t, "3", c.P(2))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Route
	r.Add(GET, "/users/:id/*", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(GET, "/users/joe/comments", c)
	c.handler(c)
	assert.Equal(t, "joe", c.P(0))
}

func TestRouterMultiRoute(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add(GET, "/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(GET, "/users", c)
	c.handler(c)
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "1", c.P(0))

	// Route > /user
	c = e.NewContext(nil, nil).(*context)
	r.Find(GET, "/user", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestRouterPriority(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(GET, "/users/new", func(c Context) error {
		c.Set("b", 2)
		return nil
	})
	r.Add(GET, "/users/:id", func(c Context) error {
		c.Set("c", 3)
		return nil
	})
	r.Add(GET, "/users/dew", func(c Context) error {
		c.Set("d", 4)
		return nil
	})
	r.Add(GET, "/users/:id/files", func(c Context) error {
		c.Set("e", 5)
		return nil
	})
	r.Add(GET, "/users/newsee", func(c Context) error {
		c.Set("f", 6)
		return nil
	})
	r.Add(GET, "/users/*", func(c Context) error {
		c.Set("g", 7)
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(GET, "/users", c)
	c.handler(c)
	assert.Equal(t, 1, c.Get("a"))

	// Route > /users/new
	r.Find(GET, "/users/new", c)
	c.handler(c)
	assert.Equal(t, 2, c.Get("b"))

	// Route > /users/:id
	r.Find(GET, "/users/1", c)
	c.handler(c)
	assert.Equal(t, 3, c.Get("c"))

	// Route > /users/dew
	r.Find(GET, "/users/dew", c)
	c.handler(c)
	assert.Equal(t, 4, c.Get("d"))

	// Route > /users/:id/files
	r.Find(GET, "/users/1/files", c)
	c.handler(c)
	assert.Equal(t, 5, c.Get("e"))

	// Route > /users/:id
	r.Find(GET, "/users/news", c)
	c.handler(c)
	assert.Equal(t, 3, c.Get("c"))

	// Route > /users/*
	r.Find(GET, "/users/joe/books", c)
	c.handler(c)
	assert.Equal(t, 7, c.Get("g"))
	assert.Equal(t, "joe/books", c.Param("_*"))
}

// Issue #372
func TestRouterPriorityNotFound(t *testing.T) {
	e := New()
	r := e.router
	c := e.NewContext(nil, nil).(*context)

	// Add
	r.Add(GET, "/a/foo", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(GET, "/a/bar", func(c Context) error {
		c.Set("b", 2)
		return nil
	})

	// Find
	r.Find(GET, "/a/foo", c)
	c.handler(c)
	assert.Equal(t, 1, c.Get("a"))

	r.Find(GET, "/a/bar", c)
	c.handler(c)
	assert.Equal(t, 2, c.Get("b"))

	c = e.NewContext(nil, nil).(*context)
	r.Find(GET, "/abc/def", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestRouterParamNames(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(GET, "/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add(GET, "/users/:id", func(c Context) error {
		return nil
	})
	r.Add(GET, "/users/:uid/files/:fid", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(GET, "/users", c)
	c.handler(c)
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "id", c.pnames[0])
	assert.Equal(t, "1", c.P(0))

	// Route > /users/:uid/files/:fid
	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "uid", c.pnames[0])
	assert.Equal(t, "1", c.P(0))
	assert.Equal(t, "fid", c.pnames[1])
	assert.Equal(t, "1", c.P(1))
}

// Issue #623
func TestRouterStaticDynamicConflict(t *testing.T) {
	e := New()
	r := e.router
	c := e.NewContext(nil, nil)

	r.Add(GET, "/dictionary/skills", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(GET, "/dictionary/:name", func(c Context) error {
		c.Set("b", 2)
		return nil
	})
	r.Add(GET, "/server", func(c Context) error {
		c.Set("c", 3)
		return nil
	})

	r.Find(GET, "/dictionary/skills", c)
	c.Handler()(c)
	assert.Equal(t, 1, c.Get("a"))
	c = e.NewContext(nil, nil)
	r.Find(GET, "/dictionary/type", c)
	c.Handler()(c)
	assert.Equal(t, 2, c.Get("b"))
	c = e.NewContext(nil, nil)
	r.Find(GET, "/server", c)
	c.Handler()(c)
	assert.Equal(t, 3, c.Get("c"))
}

func TestRouterAPI(t *testing.T) {
	e := New()
	r := e.router

	for _, route := range api {
		r.Add(route.Method, route.Path, func(c Context) error {
			return nil
		})
	}
	c := e.NewContext(nil, nil).(*context)
	for _, route := range api {
		r.Find(route.Method, route.Path, c)
		for i, n := range c.pnames {
			if assert.NotEmpty(t, n) {
				assert.Equal(t, ":"+n, c.P(i))
			}
		}
	}
}

func BenchmarkRouterGitHubAPI(b *testing.B) {
	e := New()
	r := e.router
	b.ReportAllocs()

	// Add routes
	for _, route := range api {
		r.Add(route.Method, route.Path, func(c Context) error {
			return nil
		})
	}

	// Find routes
	for i := 0; i < b.N; i++ {
		for _, route := range api {
			// c := e.pool.Get().(*context)
			c := e.AcquireContext()
			r.Find(route.Method, route.Path, c)
			// router.Find(r.Method, r.Path, c)
			e.ReleaseContext(c)
			// e.pool.Put(c)
		}
	}
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

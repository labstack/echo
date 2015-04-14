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
	params = make(Params, 5)
	api    = []route{
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
	r.Add(GET, path, func(*Context) {
		b.WriteString(path)
	}, nil)
	h, _ := r.Find(GET, path, params)
	if h == nil {
		t.Fatal("handler not found")
	}
	h(nil)
	if b.String() != path {
		t.Errorf("buffer should %s", path)
	}
}

func TestRouterParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users/:id", func(c *Context) {}, nil)
	h, _ := r.Find(GET, "/users/1", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param id should be 1")
	}
}

func TestRouterTwoParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users/:uid/files/:fid", func(*Context) {}, nil)
	h, _ := r.Find(GET, "/users/1/files/1", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param uid should be 1")
	}
	if params[1].Value != "1" {
		t.Error("param fid should be 1")
	}
}

func TestRouterCatchAll(t *testing.T) {
	r := New().Router
	r.Add(GET, "/static/*", func(*Context) {}, nil)
	h, _ := r.Find(GET, "/static/echo.gif", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "echo.gif" {
		t.Error("value should be echo.gif")
	}
}

func TestRouterMicroParam(t *testing.T) {
	r := New().Router
	r.Add(GET, "/:a/:b/:c", func(c *Context) {}, nil)
	h, _ := r.Find(GET, "/1/2/3", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param a should be 1")
	}
	if params[1].Value != "2" {
		t.Error("param b should be 2")
	}
	if params[2].Value != "3" {
		t.Error("param c should be 3")
	}
}

func TestRouterMultiRoute(t *testing.T) {
	r := New().Router
	b := new(bytes.Buffer)

	// Routes
	r.Add(GET, "/users", func(*Context) {
		b.WriteString("/users")
	}, nil)
	r.Add(GET, "/users/:id", func(c *Context) {}, nil)

	// Route > /users
	h, _ := r.Find(GET, "/users", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	h(nil)
	if b.String() != "/users" {
		t.Errorf("buffer should be /users")
	}

	// Route > /users/:id > /users/1
	h, _ = r.Find(GET, "/users/1", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param id should be 1")
	}

	// Route > /user
	h, _ = r.Find(GET, "/user", params)
	if h != nil {
		t.Fatal("handler should be nil")
	}
}

func TestRouterConflictingRoute(t *testing.T) {
	r := New().Router
	b := new(bytes.Buffer)

	// Routes
	r.Add(GET, "/users", func(*Context) {
		b.WriteString("/users")
	}, nil)
	r.Add(GET, "/users/new", func(*Context) {
		b.Reset()
		b.WriteString("/users/new")
	}, nil)
	r.Add(GET, "/users/:id", func(c *Context) {}, nil)
	r.Add(GET, "/users/new/moon", func(*Context) {
		b.Reset()
		b.WriteString("/users/new/moon")
	}, nil)
	r.Add(GET, "/users/new/:id", func(*Context) {}, nil)

	// Route > /users
	h, _ := r.Find(GET, "/users", params)
	if h == nil {
		t.Fatal("handler not found")
	}

	// Route > /users/new
	h, _ = r.Find(GET, "/users/new", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	h(nil)
	if b.String() != "/users/new" {
		t.Error("buffer should be /users/new")
	}

	// Route > /users/:id > /users/1
	h, _ = r.Find(GET, "/users/1", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param id should be 1")
	}

	// Route > /users/:id > /users/nil
	h, _ = r.Find(GET, "/users/nil", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "nil" {
		t.Error("param id should be nil")
	}

	// Route > /users/:id > /users/news
	h, _ = r.Find(GET, "/users/news", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "news" {
		t.Error("param id should be news")
	}

	//***************//
	//   Two level   //
	//***************//
	// Route > /users/new/moon > /users/new/moon
	h, _ = r.Find(GET, "/users/new/moon", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	h(nil)
	if b.String() != "/users/new/moon" {
		t.Error("buffer should be /users/new/moon")
	}

	// Route > /users/new/:id > /users/new/1
	h, _ = r.Find(GET, "/users/new/1", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "1" {
		t.Error("param id should be 1")
	}

	// Route > /users/new/:id > /users/new/me
	h, _ = r.Find(GET, "/users/new/me", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "me" {
		t.Error("param id should be me")
	}

	// Route > /users/new/:id > /users/new/moons
	h, _ = r.Find(GET, "/users/new/moons", params)
	if h == nil {
		t.Fatal("handler not found")
	}
	if params[0].Value != "moons" {
		t.Error("param id should be moons")
	}
}

func TestRouterAPI(t *testing.T) {
	r := New().Router
	for _, route := range api {
		r.Add(route.method, route.path, func(c *Context) {
			for _, p := range c.params {
				if p.Name != "" {
					if ":"+p.Name != p.Value {
						t.Errorf("param not found, method=%s, path=%s", route.method, route.path)
					}
				}
			}
		}, nil)

		h, _ := r.Find(route.method, route.path, params)
		if h == nil {
			t.Errorf("handler not found, method=%s, path=%s", route.method, route.path)
		}
	}
}

func TestRouterServeHTTP(t *testing.T) {
	r := New().Router
	r.Add(GET, "/users", func(*Context) {}, nil)

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
	fmt.Printf("%s%s, %p: parent=%p, handler=%v, echo=%v\n", p, n.prefix, n, n.parent, n.handler, n.echo)

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

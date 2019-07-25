package echo

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	staticRoutes = []*Route{
		{"GET", "/", ""},
		{"GET", "/cmd.html", ""},
		{"GET", "/code.html", ""},
		{"GET", "/contrib.html", ""},
		{"GET", "/contribute.html", ""},
		{"GET", "/debugging_with_gdb.html", ""},
		{"GET", "/docs.html", ""},
		{"GET", "/effective_go.html", ""},
		{"GET", "/files.log", ""},
		{"GET", "/gccgo_contribute.html", ""},
		{"GET", "/gccgo_install.html", ""},
		{"GET", "/go-logo-black.png", ""},
		{"GET", "/go-logo-blue.png", ""},
		{"GET", "/go-logo-white.png", ""},
		{"GET", "/go1.1.html", ""},
		{"GET", "/go1.2.html", ""},
		{"GET", "/go1.html", ""},
		{"GET", "/go1compat.html", ""},
		{"GET", "/go_faq.html", ""},
		{"GET", "/go_mem.html", ""},
		{"GET", "/go_spec.html", ""},
		{"GET", "/help.html", ""},
		{"GET", "/ie.css", ""},
		{"GET", "/install-source.html", ""},
		{"GET", "/install.html", ""},
		{"GET", "/logo-153x55.png", ""},
		{"GET", "/Makefile", ""},
		{"GET", "/root.html", ""},
		{"GET", "/share.png", ""},
		{"GET", "/sieve.gif", ""},
		{"GET", "/tos.html", ""},
		{"GET", "/articles/", ""},
		{"GET", "/articles/go_command.html", ""},
		{"GET", "/articles/index.html", ""},
		{"GET", "/articles/wiki/", ""},
		{"GET", "/articles/wiki/edit.html", ""},
		{"GET", "/articles/wiki/final-noclosure.go", ""},
		{"GET", "/articles/wiki/final-noerror.go", ""},
		{"GET", "/articles/wiki/final-parsetemplate.go", ""},
		{"GET", "/articles/wiki/final-template.go", ""},
		{"GET", "/articles/wiki/final.go", ""},
		{"GET", "/articles/wiki/get.go", ""},
		{"GET", "/articles/wiki/http-sample.go", ""},
		{"GET", "/articles/wiki/index.html", ""},
		{"GET", "/articles/wiki/Makefile", ""},
		{"GET", "/articles/wiki/notemplate.go", ""},
		{"GET", "/articles/wiki/part1-noerror.go", ""},
		{"GET", "/articles/wiki/part1.go", ""},
		{"GET", "/articles/wiki/part2.go", ""},
		{"GET", "/articles/wiki/part3-errorhandling.go", ""},
		{"GET", "/articles/wiki/part3.go", ""},
		{"GET", "/articles/wiki/test.bash", ""},
		{"GET", "/articles/wiki/test_edit.good", ""},
		{"GET", "/articles/wiki/test_Test.txt.good", ""},
		{"GET", "/articles/wiki/test_view.good", ""},
		{"GET", "/articles/wiki/view.html", ""},
		{"GET", "/codewalk/", ""},
		{"GET", "/codewalk/codewalk.css", ""},
		{"GET", "/codewalk/codewalk.js", ""},
		{"GET", "/codewalk/codewalk.xml", ""},
		{"GET", "/codewalk/functions.xml", ""},
		{"GET", "/codewalk/markov.go", ""},
		{"GET", "/codewalk/markov.xml", ""},
		{"GET", "/codewalk/pig.go", ""},
		{"GET", "/codewalk/popout.png", ""},
		{"GET", "/codewalk/run", ""},
		{"GET", "/codewalk/sharemem.xml", ""},
		{"GET", "/codewalk/urlpoll.go", ""},
		{"GET", "/devel/", ""},
		{"GET", "/devel/release.html", ""},
		{"GET", "/devel/weekly.html", ""},
		{"GET", "/gopher/", ""},
		{"GET", "/gopher/appenginegopher.jpg", ""},
		{"GET", "/gopher/appenginegophercolor.jpg", ""},
		{"GET", "/gopher/appenginelogo.gif", ""},
		{"GET", "/gopher/bumper.png", ""},
		{"GET", "/gopher/bumper192x108.png", ""},
		{"GET", "/gopher/bumper320x180.png", ""},
		{"GET", "/gopher/bumper480x270.png", ""},
		{"GET", "/gopher/bumper640x360.png", ""},
		{"GET", "/gopher/doc.png", ""},
		{"GET", "/gopher/frontpage.png", ""},
		{"GET", "/gopher/gopherbw.png", ""},
		{"GET", "/gopher/gophercolor.png", ""},
		{"GET", "/gopher/gophercolor16x16.png", ""},
		{"GET", "/gopher/help.png", ""},
		{"GET", "/gopher/pkg.png", ""},
		{"GET", "/gopher/project.png", ""},
		{"GET", "/gopher/ref.png", ""},
		{"GET", "/gopher/run.png", ""},
		{"GET", "/gopher/talks.png", ""},
		{"GET", "/gopher/pencil/", ""},
		{"GET", "/gopher/pencil/gopherhat.jpg", ""},
		{"GET", "/gopher/pencil/gopherhelmet.jpg", ""},
		{"GET", "/gopher/pencil/gophermega.jpg", ""},
		{"GET", "/gopher/pencil/gopherrunning.jpg", ""},
		{"GET", "/gopher/pencil/gopherswim.jpg", ""},
		{"GET", "/gopher/pencil/gopherswrench.jpg", ""},
		{"GET", "/play/", ""},
		{"GET", "/play/fib.go", ""},
		{"GET", "/play/hello.go", ""},
		{"GET", "/play/life.go", ""},
		{"GET", "/play/peano.go", ""},
		{"GET", "/play/pi.go", ""},
		{"GET", "/play/sieve.go", ""},
		{"GET", "/play/solitaire.go", ""},
		{"GET", "/play/tree.go", ""},
		{"GET", "/progs/", ""},
		{"GET", "/progs/cgo1.go", ""},
		{"GET", "/progs/cgo2.go", ""},
		{"GET", "/progs/cgo3.go", ""},
		{"GET", "/progs/cgo4.go", ""},
		{"GET", "/progs/defer.go", ""},
		{"GET", "/progs/defer.out", ""},
		{"GET", "/progs/defer2.go", ""},
		{"GET", "/progs/defer2.out", ""},
		{"GET", "/progs/eff_bytesize.go", ""},
		{"GET", "/progs/eff_bytesize.out", ""},
		{"GET", "/progs/eff_qr.go", ""},
		{"GET", "/progs/eff_sequence.go", ""},
		{"GET", "/progs/eff_sequence.out", ""},
		{"GET", "/progs/eff_unused1.go", ""},
		{"GET", "/progs/eff_unused2.go", ""},
		{"GET", "/progs/error.go", ""},
		{"GET", "/progs/error2.go", ""},
		{"GET", "/progs/error3.go", ""},
		{"GET", "/progs/error4.go", ""},
		{"GET", "/progs/go1.go", ""},
		{"GET", "/progs/gobs1.go", ""},
		{"GET", "/progs/gobs2.go", ""},
		{"GET", "/progs/image_draw.go", ""},
		{"GET", "/progs/image_package1.go", ""},
		{"GET", "/progs/image_package1.out", ""},
		{"GET", "/progs/image_package2.go", ""},
		{"GET", "/progs/image_package2.out", ""},
		{"GET", "/progs/image_package3.go", ""},
		{"GET", "/progs/image_package3.out", ""},
		{"GET", "/progs/image_package4.go", ""},
		{"GET", "/progs/image_package4.out", ""},
		{"GET", "/progs/image_package5.go", ""},
		{"GET", "/progs/image_package5.out", ""},
		{"GET", "/progs/image_package6.go", ""},
		{"GET", "/progs/image_package6.out", ""},
		{"GET", "/progs/interface.go", ""},
		{"GET", "/progs/interface2.go", ""},
		{"GET", "/progs/interface2.out", ""},
		{"GET", "/progs/json1.go", ""},
		{"GET", "/progs/json2.go", ""},
		{"GET", "/progs/json2.out", ""},
		{"GET", "/progs/json3.go", ""},
		{"GET", "/progs/json4.go", ""},
		{"GET", "/progs/json5.go", ""},
		{"GET", "/progs/run", ""},
		{"GET", "/progs/slices.go", ""},
		{"GET", "/progs/timeout1.go", ""},
		{"GET", "/progs/timeout2.go", ""},
		{"GET", "/progs/update.bash", ""},
	}

	gitHubAPI = []*Route{
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

	parseAPI = []*Route{
		// Objects
		{"POST", "/1/classes/:className", ""},
		{"GET", "/1/classes/:className/:objectId", ""},
		{"PUT", "/1/classes/:className/:objectId", ""},
		{"GET", "/1/classes/:className", ""},
		{"DELETE", "/1/classes/:className/:objectId", ""},

		// Users
		{"POST", "/1/users", ""},
		{"GET", "/1/login", ""},
		{"GET", "/1/users/:objectId", ""},
		{"PUT", "/1/users/:objectId", ""},
		{"GET", "/1/users", ""},
		{"DELETE", "/1/users/:objectId", ""},
		{"POST", "/1/requestPasswordReset", ""},

		// Roles
		{"POST", "/1/roles", ""},
		{"GET", "/1/roles/:objectId", ""},
		{"PUT", "/1/roles/:objectId", ""},
		{"GET", "/1/roles", ""},
		{"DELETE", "/1/roles/:objectId", ""},

		// Files
		{"POST", "/1/files/:fileName", ""},

		// Analytics
		{"POST", "/1/events/:eventName", ""},

		// Push Notifications
		{"POST", "/1/push", ""},

		// Installations
		{"POST", "/1/installations", ""},
		{"GET", "/1/installations/:objectId", ""},
		{"PUT", "/1/installations/:objectId", ""},
		{"GET", "/1/installations", ""},
		{"DELETE", "/1/installations/:objectId", ""},

		// Cloud Functions
		{"POST", "/1/functions", ""},
	}

	googlePlusAPI = []*Route{
		// People
		{"GET", "/people/:userId", ""},
		{"GET", "/people", ""},
		{"GET", "/activities/:activityId/people/:collection", ""},
		{"GET", "/people/:userId/people/:collection", ""},
		{"GET", "/people/:userId/openIdConnect", ""},

		// Activities
		{"GET", "/people/:userId/activities/:collection", ""},
		{"GET", "/activities/:activityId", ""},
		{"GET", "/activities", ""},

		// Comments
		{"GET", "/activities/:activityId/comments", ""},
		{"GET", "/comments/:commentId", ""},

		// Moments
		{"POST", "/people/:userId/moments/:collection", ""},
		{"GET", "/people/:userId/moments/:collection", ""},
		{"DELETE", "/moments/:id", ""},
	}
)

func TestRouterStatic(t *testing.T) {
	e := New()
	r := e.router
	path := "/folders/a/files/echo.gif"
	r.Add(http.MethodGet, path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, path, c)
	c.handler(c)
	assert.Equal(t, path, c.Get("path"))
}

func TestRouterParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/users/1", c)
	assert.Equal(t, "1", c.Param("id"))
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, "/users/1/files/1", c)
	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "1", c.Param("fid"))
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/a/:b/c/d/:e", func(c Context) error {
		return nil
	})

	r.Add(http.MethodGet, "/a/:b/c/:d/:f", func(c Context) error {
		return nil
	})

	c := e.NewContext(nil, nil).(*context)
	assert.NotPanics(t, func() {
		r.Find(http.MethodGet, "/a/1/c/d/2/3", c)
	})
}

func TestRouterMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/", func(Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/*", func(Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/*", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, "/", c)
	assert.Equal(t, "", c.Param("*"))

	r.Find(http.MethodGet, "/download", c)
	assert.Equal(t, "download", c.Param("*"))

	r.Find(http.MethodGet, "/users/joe", c)
	assert.Equal(t, "joe", c.Param("*"))
}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	e := New()
	r := e.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	// Routes
	r.Add(http.MethodGet, "/api/users/jack", handler)
	r.Add(http.MethodGet, "/api/users/jill", handler)
	r.Add(http.MethodGet, "/*", handler)

	c := e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/api/users/jack", c)
	c.handler(c)
	assert.Equal(t, "/api/users/jack", c.Get("path"))

	r.Find(http.MethodGet, "/api/users/jill", c)
	c.handler(c)
	assert.Equal(t, "/api/users/jill", c.Get("path"))

	r.Find(http.MethodGet, "/api/users/joe", c)
	c.handler(c)
	assert.Equal(t, "/*", c.Get("path"))
}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(http.MethodGet, "/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/1/2/3", c)
	assert.Equal(t, "1", c.Param("a"))
	assert.Equal(t, "2", c.Param("b"))
	assert.Equal(t, "3", c.Param("c"))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Route
	r.Add(http.MethodGet, "/users/:id/*", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, "/users/joe/comments", c)
	c.handler(c)
	assert.Equal(t, "joe", c.Param("id"))
}

func TestRouterMultiRoute(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/1", c)
	assert.Equal(t, "1", c.Param("id"))

	// Route > /user
	c = e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/user", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestRouterPriority(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(http.MethodGet, "/users/new", func(c Context) error {
		c.Set("b", 2)
		return nil
	})
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		c.Set("c", 3)
		return nil
	})
	r.Add(http.MethodGet, "/users/dew", func(c Context) error {
		c.Set("d", 4)
		return nil
	})
	r.Add(http.MethodGet, "/users/:id/files", func(c Context) error {
		c.Set("e", 5)
		return nil
	})
	r.Add(http.MethodGet, "/users/newsee", func(c Context) error {
		c.Set("f", 6)
		return nil
	})
	r.Add(http.MethodGet, "/users/*", func(c Context) error {
		c.Set("g", 7)
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal(t, 1, c.Get("a"))

	// Route > /users/new
	r.Find(http.MethodGet, "/users/new", c)
	c.handler(c)
	assert.Equal(t, 2, c.Get("b"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/1", c)
	c.handler(c)
	assert.Equal(t, 3, c.Get("c"))

	// Route > /users/dew
	r.Find(http.MethodGet, "/users/dew", c)
	c.handler(c)
	assert.Equal(t, 4, c.Get("d"))

	// Route > /users/:id/files
	r.Find(http.MethodGet, "/users/1/files", c)
	c.handler(c)
	assert.Equal(t, 5, c.Get("e"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/news", c)
	c.handler(c)
	assert.Equal(t, 3, c.Get("c"))

	// Route > /users/*
	r.Find(http.MethodGet, "/users/joe/books", c)
	c.handler(c)
	assert.Equal(t, 7, c.Get("g"))
	assert.Equal(t, "joe/books", c.Param("*"))
}

func TestRouterIssue1348(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/:lang/", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:lang/dupa", func(c Context) error {
		return nil
	})
}

// Issue #372
func TestRouterPriorityNotFound(t *testing.T) {
	e := New()
	r := e.router
	c := e.NewContext(nil, nil).(*context)

	// Add
	r.Add(http.MethodGet, "/a/foo", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(http.MethodGet, "/a/bar", func(c Context) error {
		c.Set("b", 2)
		return nil
	})

	// Find
	r.Find(http.MethodGet, "/a/foo", c)
	c.handler(c)
	assert.Equal(t, 1, c.Get("a"))

	r.Find(http.MethodGet, "/a/bar", c)
	c.handler(c)
	assert.Equal(t, 2, c.Get("b"))

	c = e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/abc/def", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestRouterParamNames(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	// Route > /users
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/1", c)
	assert.Equal(t, "id", c.pnames[0])
	assert.Equal(t, "1", c.Param("id"))

	// Route > /users/:uid/files/:fid
	r.Find(http.MethodGet, "/users/1/files/1", c)
	assert.Equal(t, "uid", c.pnames[0])
	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "fid", c.pnames[1])
	assert.Equal(t, "1", c.Param("fid"))
}

// Issue #623
func TestRouterStaticDynamicConflict(t *testing.T) {
	e := New()
	r := e.router
	c := e.NewContext(nil, nil)

	r.Add(http.MethodGet, "/dictionary/skills", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add(http.MethodGet, "/dictionary/:name", func(c Context) error {
		c.Set("b", 2)
		return nil
	})
	r.Add(http.MethodGet, "/server", func(c Context) error {
		c.Set("c", 3)
		return nil
	})

	r.Find(http.MethodGet, "/dictionary/skills", c)
	c.Handler()(c)
	assert.Equal(t, 1, c.Get("a"))
	c = e.NewContext(nil, nil)
	r.Find(http.MethodGet, "/dictionary/type", c)
	c.Handler()(c)
	assert.Equal(t, 2, c.Get("b"))
	c = e.NewContext(nil, nil)
	r.Find(http.MethodGet, "/server", c)
	c.Handler()(c)
	assert.Equal(t, 3, c.Get("c"))
}

// Issue #1348
func TestRouterParamBacktraceNotFound(t *testing.T) {
	e := New()
	r := e.router

	// Add
	r.Add(http.MethodGet, "/:param1", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/foo", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/bar", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/bar/:param2", func(c Context) error {
		return nil
	})

	c := e.NewContext(nil, nil).(*context)

	//Find
	r.Find(http.MethodGet, "/a", c)
	assert.Equal(t, "a", c.Param("param1"))

	r.Find(http.MethodGet, "/a/foo", c)
	assert.Equal(t, "a", c.Param("param1"))

	r.Find(http.MethodGet, "/a/bar", c)
	assert.Equal(t, "a", c.Param("param1"))

	r.Find(http.MethodGet, "/a/bar/b", c)
	assert.Equal(t, "a", c.Param("param1"))
	assert.Equal(t, "b", c.Param("param2"))

	c = e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/a/bbbbb", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func testRouterAPI(t *testing.T, api []*Route) {
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
		tokens := strings.Split(route.Path[1:], "/")
		for _, token := range tokens {
			if token[0] == ':' {
				assert.Equal(t, c.Param(token[1:]), token)
			}
		}
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

// Issue #729
func TestRouterParamAlias(t *testing.T) {
	api := []*Route{
		{http.MethodGet, "/users/:userID/following", ""},
		{http.MethodGet, "/users/:userID/followedBy", ""},
		{http.MethodGet, "/users/:userID/follow", ""},
	}
	testRouterAPI(t, api)
}

// Issue #1052
func TestRouterParamOrdering(t *testing.T) {
	api := []*Route{
		{http.MethodGet, "/:a/:b/:c/:id", ""},
		{http.MethodGet, "/:a/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{http.MethodGet, "/:a/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
		{http.MethodGet, "/:a/:b/:c/:id", ""},
	}
	testRouterAPI(t, api2)
	api3 := []*Route{
		{http.MethodGet, "/:a/:b/:c/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
		{http.MethodGet, "/:a/:id", ""},
	}
	testRouterAPI(t, api3)
}

// Issue #1139
func TestRouterMixedParams(t *testing.T) {
	api := []*Route{
		{http.MethodGet, "/teacher/:tid/room/suggestions", ""},
		{http.MethodGet, "/teacher/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{http.MethodGet, "/teacher/:id", ""},
		{http.MethodGet, "/teacher/:tid/room/suggestions", ""},
	}
	testRouterAPI(t, api2)
}

func benchmarkRouterRoutes(b *testing.B, routes []*Route) {
	e := New()
	r := e.router
	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		r.Add(route.Method, route.Path, func(c Context) error {
			return nil
		})
	}

	// Find routes
	for i := 0; i < b.N; i++ {
		for _, route := range gitHubAPI {
			c := e.pool.Get().(*context)
			r.Find(route.Method, route.Path, c)
			e.pool.Put(c)
		}
	}
}

func BenchmarkRouterStaticRoutes(b *testing.B) {
	benchmarkRouterRoutes(b, staticRoutes)
}

func BenchmarkRouterGitHubAPI(b *testing.B) {
	benchmarkRouterRoutes(b, gitHubAPI)
}

func BenchmarkRouterParseAPI(b *testing.B) {
	benchmarkRouterRoutes(b, parseAPI)
}

func BenchmarkRouterGooglePlusAPI(b *testing.B) {
	benchmarkRouterRoutes(b, googlePlusAPI)
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s, %p: type=%d, parent=%p, handler=%v, pnames=%v\n", p, n.prefix, n, n.kind, n.parent, n.methodHandler, n.pnames)

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

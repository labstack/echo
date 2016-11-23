package echo

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	staticRoutes = []Route{
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

	gitHubAPI = []Route{
		// OAuth Authorizations
		{"GET", "/authorizations", "", nil},
		{"GET", "/authorizations/:id", "", nil},
		{"POST", "/authorizations", "", nil},
		//{"PUT", "/authorizations/clients/:client_id", "", nil},
		//{"PATCH", "/authorizations/:id", "", nil},
		{"DELETE", "/authorizations/:id", "", nil},
		{"GET", "/applications/:client_id/tokens/:access_token", "", nil},
		{"DELETE", "/applications/:client_id/tokens", "", nil},
		{"DELETE", "/applications/:client_id/tokens/:access_token", "", nil},

		// Activity
		{"GET", "/events", "", nil},
		{"GET", "/repos/:owner/:repo/events", "", nil},
		{"GET", "/networks/:owner/:repo/events", "", nil},
		{"GET", "/orgs/:org/events", "", nil},
		{"GET", "/users/:user/received_events", "", nil},
		{"GET", "/users/:user/received_events/public", "", nil},
		{"GET", "/users/:user/events", "", nil},
		{"GET", "/users/:user/events/public", "", nil},
		{"GET", "/users/:user/events/orgs/:org", "", nil},
		{"GET", "/feeds", "", nil},
		{"GET", "/notifications", "", nil},
		{"GET", "/repos/:owner/:repo/notifications", "", nil},
		{"PUT", "/notifications", "", nil},
		{"PUT", "/repos/:owner/:repo/notifications", "", nil},
		{"GET", "/notifications/threads/:id", "", nil},
		//{"PATCH", "/notifications/threads/:id", "", nil},
		{"GET", "/notifications/threads/:id/subscription", "", nil},
		{"PUT", "/notifications/threads/:id/subscription", "", nil},
		{"DELETE", "/notifications/threads/:id/subscription", "", nil},
		{"GET", "/repos/:owner/:repo/stargazers", "", nil},
		{"GET", "/users/:user/starred", "", nil},
		{"GET", "/user/starred", "", nil},
		{"GET", "/user/starred/:owner/:repo", "", nil},
		{"PUT", "/user/starred/:owner/:repo", "", nil},
		{"DELETE", "/user/starred/:owner/:repo", "", nil},
		{"GET", "/repos/:owner/:repo/subscribers", "", nil},
		{"GET", "/users/:user/subscriptions", "", nil},
		{"GET", "/user/subscriptions", "", nil},
		{"GET", "/repos/:owner/:repo/subscription", "", nil},
		{"PUT", "/repos/:owner/:repo/subscription", "", nil},
		{"DELETE", "/repos/:owner/:repo/subscription", "", nil},
		{"GET", "/user/subscriptions/:owner/:repo", "", nil},
		{"PUT", "/user/subscriptions/:owner/:repo", "", nil},
		{"DELETE", "/user/subscriptions/:owner/:repo", "", nil},

		// Gists
		{"GET", "/users/:user/gists", "", nil},
		{"GET", "/gists", "", nil},
		//{"GET", "/gists/public", "", nil},
		//{"GET", "/gists/starred", "", nil},
		{"GET", "/gists/:id", "", nil},
		{"POST", "/gists", "", nil},
		//{"PATCH", "/gists/:id", "", nil},
		{"PUT", "/gists/:id/star", "", nil},
		{"DELETE", "/gists/:id/star", "", nil},
		{"GET", "/gists/:id/star", "", nil},
		{"POST", "/gists/:id/forks", "", nil},
		{"DELETE", "/gists/:id", "", nil},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha", "", nil},
		{"POST", "/repos/:owner/:repo/git/blobs", "", nil},
		{"GET", "/repos/:owner/:repo/git/commits/:sha", "", nil},
		{"POST", "/repos/:owner/:repo/git/commits", "", nil},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref", "", nil},
		{"GET", "/repos/:owner/:repo/git/refs", "", nil},
		{"POST", "/repos/:owner/:repo/git/refs", "", nil},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref", "", nil},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref", "", nil},
		{"GET", "/repos/:owner/:repo/git/tags/:sha", "", nil},
		{"POST", "/repos/:owner/:repo/git/tags", "", nil},
		{"GET", "/repos/:owner/:repo/git/trees/:sha", "", nil},
		{"POST", "/repos/:owner/:repo/git/trees", "", nil},

		// Issues
		{"GET", "/issues", "", nil},
		{"GET", "/user/issues", "", nil},
		{"GET", "/orgs/:org/issues", "", nil},
		{"GET", "/repos/:owner/:repo/issues", "", nil},
		{"GET", "/repos/:owner/:repo/issues/:number", "", nil},
		{"POST", "/repos/:owner/:repo/issues", "", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/:number", "", nil},
		{"GET", "/repos/:owner/:repo/assignees", "", nil},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", "", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", "", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments", "", nil},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id", "", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/comments", "", nil},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id", "", nil},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id", "", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/events", "", nil},
		//{"GET", "/repos/:owner/:repo/issues/events", "", nil},
		//{"GET", "/repos/:owner/:repo/issues/events/:id", "", nil},
		{"GET", "/repos/:owner/:repo/labels", "", nil},
		{"GET", "/repos/:owner/:repo/labels/:name", "", nil},
		{"POST", "/repos/:owner/:repo/labels", "", nil},
		//{"PATCH", "/repos/:owner/:repo/labels/:name", "", nil},
		{"DELETE", "/repos/:owner/:repo/labels/:name", "", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/labels", "", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/labels", "", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", "", nil},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels", "", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels", "", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels", "", nil},
		{"GET", "/repos/:owner/:repo/milestones", "", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number", "", nil},
		{"POST", "/repos/:owner/:repo/milestones", "", nil},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number", "", nil},
		{"DELETE", "/repos/:owner/:repo/milestones/:number", "", nil},

		// Miscellaneous
		{"GET", "/emojis", "", nil},
		{"GET", "/gitignore/templates", "", nil},
		{"GET", "/gitignore/templates/:name", "", nil},
		{"POST", "/markdown", "", nil},
		{"POST", "/markdown/raw", "", nil},
		{"GET", "/meta", "", nil},
		{"GET", "/rate_limit", "", nil},

		// Organizations
		{"GET", "/users/:user/orgs", "", nil},
		{"GET", "/user/orgs", "", nil},
		{"GET", "/orgs/:org", "", nil},
		//{"PATCH", "/orgs/:org", "", nil},
		{"GET", "/orgs/:org/members", "", nil},
		{"GET", "/orgs/:org/members/:user", "", nil},
		{"DELETE", "/orgs/:org/members/:user", "", nil},
		{"GET", "/orgs/:org/public_members", "", nil},
		{"GET", "/orgs/:org/public_members/:user", "", nil},
		{"PUT", "/orgs/:org/public_members/:user", "", nil},
		{"DELETE", "/orgs/:org/public_members/:user", "", nil},
		{"GET", "/orgs/:org/teams", "", nil},
		{"GET", "/teams/:id", "", nil},
		{"POST", "/orgs/:org/teams", "", nil},
		//{"PATCH", "/teams/:id", "", nil},
		{"DELETE", "/teams/:id", "", nil},
		{"GET", "/teams/:id/members", "", nil},
		{"GET", "/teams/:id/members/:user", "", nil},
		{"PUT", "/teams/:id/members/:user", "", nil},
		{"DELETE", "/teams/:id/members/:user", "", nil},
		{"GET", "/teams/:id/repos", "", nil},
		{"GET", "/teams/:id/repos/:owner/:repo", "", nil},
		{"PUT", "/teams/:id/repos/:owner/:repo", "", nil},
		{"DELETE", "/teams/:id/repos/:owner/:repo", "", nil},
		{"GET", "/user/teams", "", nil},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls", "", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number", "", nil},
		{"POST", "/repos/:owner/:repo/pulls", "", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number", "", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits", "", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", "", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", "", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", "", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", "", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments", "", nil},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number", "", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", "", nil},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", "", nil},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", "", nil},

		// Repositories
		{"GET", "/user/repos", "", nil},
		{"GET", "/users/:user/repos", "", nil},
		{"GET", "/orgs/:org/repos", "", nil},
		{"GET", "/repositories", "", nil},
		{"POST", "/user/repos", "", nil},
		{"POST", "/orgs/:org/repos", "", nil},
		{"GET", "/repos/:owner/:repo", "", nil},
		//{"PATCH", "/repos/:owner/:repo", "", nil},
		{"GET", "/repos/:owner/:repo/contributors", "", nil},
		{"GET", "/repos/:owner/:repo/languages", "", nil},
		{"GET", "/repos/:owner/:repo/teams", "", nil},
		{"GET", "/repos/:owner/:repo/tags", "", nil},
		{"GET", "/repos/:owner/:repo/branches", "", nil},
		{"GET", "/repos/:owner/:repo/branches/:branch", "", nil},
		{"DELETE", "/repos/:owner/:repo", "", nil},
		{"GET", "/repos/:owner/:repo/collaborators", "", nil},
		{"GET", "/repos/:owner/:repo/collaborators/:user", "", nil},
		{"PUT", "/repos/:owner/:repo/collaborators/:user", "", nil},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user", "", nil},
		{"GET", "/repos/:owner/:repo/comments", "", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments", "", nil},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments", "", nil},
		{"GET", "/repos/:owner/:repo/comments/:id", "", nil},
		//{"PATCH", "/repos/:owner/:repo/comments/:id", "", nil},
		{"DELETE", "/repos/:owner/:repo/comments/:id", "", nil},
		{"GET", "/repos/:owner/:repo/commits", "", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha", "", nil},
		{"GET", "/repos/:owner/:repo/readme", "", nil},
		//{"GET", "/repos/:owner/:repo/contents/*path", "", nil},
		//{"PUT", "/repos/:owner/:repo/contents/*path", "", nil},
		//{"DELETE", "/repos/:owner/:repo/contents/*path", "", nil},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref", "", nil},
		{"GET", "/repos/:owner/:repo/keys", "", nil},
		{"GET", "/repos/:owner/:repo/keys/:id", "", nil},
		{"POST", "/repos/:owner/:repo/keys", "", nil},
		//{"PATCH", "/repos/:owner/:repo/keys/:id", "", nil},
		{"DELETE", "/repos/:owner/:repo/keys/:id", "", nil},
		{"GET", "/repos/:owner/:repo/downloads", "", nil},
		{"GET", "/repos/:owner/:repo/downloads/:id", "", nil},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", "", nil},
		{"GET", "/repos/:owner/:repo/forks", "", nil},
		{"POST", "/repos/:owner/:repo/forks", "", nil},
		{"GET", "/repos/:owner/:repo/hooks", "", nil},
		{"GET", "/repos/:owner/:repo/hooks/:id", "", nil},
		{"POST", "/repos/:owner/:repo/hooks", "", nil},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id", "", nil},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests", "", nil},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", "", nil},
		{"POST", "/repos/:owner/:repo/merges", "", nil},
		{"GET", "/repos/:owner/:repo/releases", "", nil},
		{"GET", "/repos/:owner/:repo/releases/:id", "", nil},
		{"POST", "/repos/:owner/:repo/releases", "", nil},
		//{"PATCH", "/repos/:owner/:repo/releases/:id", "", nil},
		{"DELETE", "/repos/:owner/:repo/releases/:id", "", nil},
		{"GET", "/repos/:owner/:repo/releases/:id/assets", "", nil},
		{"GET", "/repos/:owner/:repo/stats/contributors", "", nil},
		{"GET", "/repos/:owner/:repo/stats/commit_activity", "", nil},
		{"GET", "/repos/:owner/:repo/stats/code_frequency", "", nil},
		{"GET", "/repos/:owner/:repo/stats/participation", "", nil},
		{"GET", "/repos/:owner/:repo/stats/punch_card", "", nil},
		{"GET", "/repos/:owner/:repo/statuses/:ref", "", nil},
		{"POST", "/repos/:owner/:repo/statuses/:ref", "", nil},

		// Search
		{"GET", "/search/repositories", "", nil},
		{"GET", "/search/code", "", nil},
		{"GET", "/search/issues", "", nil},
		{"GET", "/search/users", "", nil},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", "", nil},
		{"GET", "/legacy/repos/search/:keyword", "", nil},
		{"GET", "/legacy/user/search/:keyword", "", nil},
		{"GET", "/legacy/user/email/:email", "", nil},

		// Users
		{"GET", "/users/:user", "", nil},
		{"GET", "/user", "", nil},
		//{"PATCH", "/user", "", nil},
		{"GET", "/users", "", nil},
		{"GET", "/user/emails", "", nil},
		{"POST", "/user/emails", "", nil},
		{"DELETE", "/user/emails", "", nil},
		{"GET", "/users/:user/followers", "", nil},
		{"GET", "/user/followers", "", nil},
		{"GET", "/users/:user/following", "", nil},
		{"GET", "/user/following", "", nil},
		{"GET", "/user/following/:user", "", nil},
		{"GET", "/users/:user/following/:target_user", "", nil},
		{"PUT", "/user/following/:user", "", nil},
		{"DELETE", "/user/following/:user", "", nil},
		{"GET", "/users/:user/keys", "", nil},
		{"GET", "/user/keys", "", nil},
		{"GET", "/user/keys/:id", "", nil},
		{"POST", "/user/keys", "", nil},
		//{"PATCH", "/user/keys/:id", "", nil},
		{"DELETE", "/user/keys/:id", "", nil},
	}

	parseAPI = []Route{
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

	googlePlusAPI = []Route{
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
	assert.Equal(t, "1", c.Param("id"))
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)

	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "1", c.Param("fid"))
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
	assert.Equal(t, "", c.Param("*"))

	r.Find(GET, "/download", c)
	assert.Equal(t, "download", c.Param("*"))

	r.Find(GET, "/users/joe", c)
	assert.Equal(t, "joe", c.Param("*"))
}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(GET, "/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*context)
	r.Find(GET, "/1/2/3", c)
	assert.Equal(t, "1", c.Param("a"))
	assert.Equal(t, "2", c.Param("b"))
	assert.Equal(t, "3", c.Param("c"))
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
	assert.Equal(t, "joe", c.Param("id"))
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
	assert.Equal(t, "1", c.Param("id"))

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
	assert.Equal(t, "joe/books", c.Param("*"))
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
	assert.Equal(t, "1", c.Param("id"))

	// Route > /users/:uid/files/:fid
	r.Find(GET, "/users/1/files/1", c)
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

func testRouterAPI(t *testing.T, api []Route) {
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
	api := []Route{
		{GET, "/users/:userID/following", ""},
		{GET, "/users/:userID/followedBy", ""},
		{GET, "/users/:userID/follow", ""},
	}
	testRouterAPI(t, api)
}

func benchmarkRouterRoutes(b *testing.B, routes []Route) {
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

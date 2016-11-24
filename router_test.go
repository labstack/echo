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
		{"GET", "/", "", nil},
		{"GET", "/cmd.html", "", nil},
		{"GET", "/code.html", "", nil},
		{"GET", "/contrib.html", "", nil},
		{"GET", "/contribute.html", "", nil},
		{"GET", "/debugging_with_gdb.html", "", nil},
		{"GET", "/docs.html", "", nil},
		{"GET", "/effective_go.html", "", nil},
		{"GET", "/files.log", "", nil},
		{"GET", "/gccgo_contribute.html", "", nil},
		{"GET", "/gccgo_install.html", "", nil},
		{"GET", "/go-logo-black.png", "", nil},
		{"GET", "/go-logo-blue.png", "", nil},
		{"GET", "/go-logo-white.png", "", nil},
		{"GET", "/go1.1.html", "", nil},
		{"GET", "/go1.2.html", "", nil},
		{"GET", "/go1.html", "", nil},
		{"GET", "/go1compat.html", "", nil},
		{"GET", "/go_faq.html", "", nil},
		{"GET", "/go_mem.html", "", nil},
		{"GET", "/go_spec.html", "", nil},
		{"GET", "/help.html", "", nil},
		{"GET", "/ie.css", "", nil},
		{"GET", "/install-source.html", "", nil},
		{"GET", "/install.html", "", nil},
		{"GET", "/logo-153x55.png", "", nil},
		{"GET", "/Makefile", "", nil},
		{"GET", "/root.html", "", nil},
		{"GET", "/share.png", "", nil},
		{"GET", "/sieve.gif", "", nil},
		{"GET", "/tos.html", "", nil},
		{"GET", "/articles/", "", nil},
		{"GET", "/articles/go_command.html", "", nil},
		{"GET", "/articles/index.html", "", nil},
		{"GET", "/articles/wiki/", "", nil},
		{"GET", "/articles/wiki/edit.html", "", nil},
		{"GET", "/articles/wiki/final-noclosure.go", "", nil},
		{"GET", "/articles/wiki/final-noerror.go", "", nil},
		{"GET", "/articles/wiki/final-parsetemplate.go", "", nil},
		{"GET", "/articles/wiki/final-template.go", "", nil},
		{"GET", "/articles/wiki/final.go", "", nil},
		{"GET", "/articles/wiki/get.go", "", nil},
		{"GET", "/articles/wiki/http-sample.go", "", nil},
		{"GET", "/articles/wiki/index.html", "", nil},
		{"GET", "/articles/wiki/Makefile", "", nil},
		{"GET", "/articles/wiki/notemplate.go", "", nil},
		{"GET", "/articles/wiki/part1-noerror.go", "", nil},
		{"GET", "/articles/wiki/part1.go", "", nil},
		{"GET", "/articles/wiki/part2.go", "", nil},
		{"GET", "/articles/wiki/part3-errorhandling.go", "", nil},
		{"GET", "/articles/wiki/part3.go", "", nil},
		{"GET", "/articles/wiki/test.bash", "", nil},
		{"GET", "/articles/wiki/test_edit.good", "", nil},
		{"GET", "/articles/wiki/test_Test.txt.good", "", nil},
		{"GET", "/articles/wiki/test_view.good", "", nil},
		{"GET", "/articles/wiki/view.html", "", nil},
		{"GET", "/codewalk/", "", nil},
		{"GET", "/codewalk/codewalk.css", "", nil},
		{"GET", "/codewalk/codewalk.js", "", nil},
		{"GET", "/codewalk/codewalk.xml", "", nil},
		{"GET", "/codewalk/functions.xml", "", nil},
		{"GET", "/codewalk/markov.go", "", nil},
		{"GET", "/codewalk/markov.xml", "", nil},
		{"GET", "/codewalk/pig.go", "", nil},
		{"GET", "/codewalk/popout.png", "", nil},
		{"GET", "/codewalk/run", "", nil},
		{"GET", "/codewalk/sharemem.xml", "", nil},
		{"GET", "/codewalk/urlpoll.go", "", nil},
		{"GET", "/devel/", "", nil},
		{"GET", "/devel/release.html", "", nil},
		{"GET", "/devel/weekly.html", "", nil},
		{"GET", "/gopher/", "", nil},
		{"GET", "/gopher/appenginegopher.jpg", "", nil},
		{"GET", "/gopher/appenginegophercolor.jpg", "", nil},
		{"GET", "/gopher/appenginelogo.gif", "", nil},
		{"GET", "/gopher/bumper.png", "", nil},
		{"GET", "/gopher/bumper192x108.png", "", nil},
		{"GET", "/gopher/bumper320x180.png", "", nil},
		{"GET", "/gopher/bumper480x270.png", "", nil},
		{"GET", "/gopher/bumper640x360.png", "", nil},
		{"GET", "/gopher/doc.png", "", nil},
		{"GET", "/gopher/frontpage.png", "", nil},
		{"GET", "/gopher/gopherbw.png", "", nil},
		{"GET", "/gopher/gophercolor.png", "", nil},
		{"GET", "/gopher/gophercolor16x16.png", "", nil},
		{"GET", "/gopher/help.png", "", nil},
		{"GET", "/gopher/pkg.png", "", nil},
		{"GET", "/gopher/project.png", "", nil},
		{"GET", "/gopher/ref.png", "", nil},
		{"GET", "/gopher/run.png", "", nil},
		{"GET", "/gopher/talks.png", "", nil},
		{"GET", "/gopher/pencil/", "", nil},
		{"GET", "/gopher/pencil/gopherhat.jpg", "", nil},
		{"GET", "/gopher/pencil/gopherhelmet.jpg", "", nil},
		{"GET", "/gopher/pencil/gophermega.jpg", "", nil},
		{"GET", "/gopher/pencil/gopherrunning.jpg", "", nil},
		{"GET", "/gopher/pencil/gopherswim.jpg", "", nil},
		{"GET", "/gopher/pencil/gopherswrench.jpg", "", nil},
		{"GET", "/play/", "", nil},
		{"GET", "/play/fib.go", "", nil},
		{"GET", "/play/hello.go", "", nil},
		{"GET", "/play/life.go", "", nil},
		{"GET", "/play/peano.go", "", nil},
		{"GET", "/play/pi.go", "", nil},
		{"GET", "/play/sieve.go", "", nil},
		{"GET", "/play/solitaire.go", "", nil},
		{"GET", "/play/tree.go", "", nil},
		{"GET", "/progs/", "", nil},
		{"GET", "/progs/cgo1.go", "", nil},
		{"GET", "/progs/cgo2.go", "", nil},
		{"GET", "/progs/cgo3.go", "", nil},
		{"GET", "/progs/cgo4.go", "", nil},
		{"GET", "/progs/defer.go", "", nil},
		{"GET", "/progs/defer.out", "", nil},
		{"GET", "/progs/defer2.go", "", nil},
		{"GET", "/progs/defer2.out", "", nil},
		{"GET", "/progs/eff_bytesize.go", "", nil},
		{"GET", "/progs/eff_bytesize.out", "", nil},
		{"GET", "/progs/eff_qr.go", "", nil},
		{"GET", "/progs/eff_sequence.go", "", nil},
		{"GET", "/progs/eff_sequence.out", "", nil},
		{"GET", "/progs/eff_unused1.go", "", nil},
		{"GET", "/progs/eff_unused2.go", "", nil},
		{"GET", "/progs/error.go", "", nil},
		{"GET", "/progs/error2.go", "", nil},
		{"GET", "/progs/error3.go", "", nil},
		{"GET", "/progs/error4.go", "", nil},
		{"GET", "/progs/go1.go", "", nil},
		{"GET", "/progs/gobs1.go", "", nil},
		{"GET", "/progs/gobs2.go", "", nil},
		{"GET", "/progs/image_draw.go", "", nil},
		{"GET", "/progs/image_package1.go", "", nil},
		{"GET", "/progs/image_package1.out", "", nil},
		{"GET", "/progs/image_package2.go", "", nil},
		{"GET", "/progs/image_package2.out", "", nil},
		{"GET", "/progs/image_package3.go", "", nil},
		{"GET", "/progs/image_package3.out", "", nil},
		{"GET", "/progs/image_package4.go", "", nil},
		{"GET", "/progs/image_package4.out", "", nil},
		{"GET", "/progs/image_package5.go", "", nil},
		{"GET", "/progs/image_package5.out", "", nil},
		{"GET", "/progs/image_package6.go", "", nil},
		{"GET", "/progs/image_package6.out", "", nil},
		{"GET", "/progs/interface.go", "", nil},
		{"GET", "/progs/interface2.go", "", nil},
		{"GET", "/progs/interface2.out", "", nil},
		{"GET", "/progs/json1.go", "", nil},
		{"GET", "/progs/json2.go", "", nil},
		{"GET", "/progs/json2.out", "", nil},
		{"GET", "/progs/json3.go", "", nil},
		{"GET", "/progs/json4.go", "", nil},
		{"GET", "/progs/json5.go", "", nil},
		{"GET", "/progs/run", "", nil},
		{"GET", "/progs/slices.go", "", nil},
		{"GET", "/progs/timeout1.go", "", nil},
		{"GET", "/progs/timeout2.go", "", nil},
		{"GET", "/progs/update.bash", "", nil},
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
		{"POST", "/1/classes/:className", "", nil},
		{"GET", "/1/classes/:className/:objectId", "", nil},
		{"PUT", "/1/classes/:className/:objectId", "", nil},
		{"GET", "/1/classes/:className", "", nil},
		{"DELETE", "/1/classes/:className/:objectId", "", nil},

		// Users
		{"POST", "/1/users", "", nil},
		{"GET", "/1/login", "", nil},
		{"GET", "/1/users/:objectId", "", nil},
		{"PUT", "/1/users/:objectId", "", nil},
		{"GET", "/1/users", "", nil},
		{"DELETE", "/1/users/:objectId", "", nil},
		{"POST", "/1/requestPasswordReset", "", nil},

		// Roles
		{"POST", "/1/roles", "", nil},
		{"GET", "/1/roles/:objectId", "", nil},
		{"PUT", "/1/roles/:objectId", "", nil},
		{"GET", "/1/roles", "", nil},
		{"DELETE", "/1/roles/:objectId", "", nil},

		// Files
		{"POST", "/1/files/:fileName", "", nil},

		// Analytics
		{"POST", "/1/events/:eventName", "", nil},

		// Push Notifications
		{"POST", "/1/push", "", nil},

		// Installations
		{"POST", "/1/installations", "", nil},
		{"GET", "/1/installations/:objectId", "", nil},
		{"PUT", "/1/installations/:objectId", "", nil},
		{"GET", "/1/installations", "", nil},
		{"DELETE", "/1/installations/:objectId", "", nil},

		// Cloud Functions
		{"POST", "/1/functions", "", nil},
	}

	googlePlusAPI = []Route{
		// People
		{"GET", "/people/:userId", "", nil},
		{"GET", "/people", "", nil},
		{"GET", "/activities/:activityId/people/:collection", "", nil},
		{"GET", "/people/:userId/people/:collection", "", nil},
		{"GET", "/people/:userId/openIdConnect", "", nil},

		// Activities
		{"GET", "/people/:userId/activities/:collection", "", nil},
		{"GET", "/activities/:activityId", "", nil},
		{"GET", "/activities", "", nil},

		// Comments
		{"GET", "/activities/:activityId/comments", "", nil},
		{"GET", "/comments/:commentId", "", nil},

		// Moments
		{"POST", "/people/:userId/moments/:collection", "", nil},
		{"GET", "/people/:userId/moments/:collection", "", nil},
		{"DELETE", "/moments/:id", "", nil},
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
		{GET, "/users/:userID/following", "", nil},
		{GET, "/users/:userID/followedBy", "", nil},
		{GET, "/users/:userID/follow", "", nil},
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

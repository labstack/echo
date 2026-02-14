// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testRoute struct {
	Method  string
	Path    string
	Handler string
}

var (
	staticRoutes = []testRoute{
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

	gitHubAPI = []testRoute{
		// OAuth Authorizations
		{"GET", "/authorizations", ""},
		{"GET", "/authorizations/:id", ""},
		{"POST", "/authorizations", ""},

		{"PUT", "/authorizations/clients/:client_id", ""},
		{"PATCH", "/authorizations/:id", ""},

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

		{"PATCH", "/notifications/threads/:id", ""},

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

		{"GET", "/gists/public", ""},
		{"GET", "/gists/starred", ""},

		{"GET", "/gists/:id", ""},
		{"POST", "/gists", ""},

		{"PATCH", "/gists/:id", ""},

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

		{"GET", "/repos/:owner/:repo/git/refs/*ref", ""},

		{"GET", "/repos/:owner/:repo/git/refs", ""},
		{"POST", "/repos/:owner/:repo/git/refs", ""},

		{"PATCH", "/repos/:owner/:repo/git/refs/*ref", ""},
		{"DELETE", "/repos/:owner/:repo/git/refs/*ref", ""},

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

		{"PATCH", "/repos/:owner/:repo/issues/:number", ""},

		{"GET", "/repos/:owner/:repo/assignees", ""},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", ""},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", ""},

		{"GET", "/repos/:owner/:repo/issues/comments", ""},
		{"GET", "/repos/:owner/:repo/issues/comments/:id", ""},

		{"POST", "/repos/:owner/:repo/issues/:number/comments", ""},

		{"PATCH", "/repos/:owner/:repo/issues/comments/:id", ""},
		{"DELETE", "/repos/:owner/:repo/issues/comments/:id", ""},

		{"GET", "/repos/:owner/:repo/issues/:number/events", ""},

		{"GET", "/repos/:owner/:repo/issues/events", ""},
		{"GET", "/repos/:owner/:repo/issues/events/:id", ""},

		{"GET", "/repos/:owner/:repo/labels", ""},
		{"GET", "/repos/:owner/:repo/labels/:name", ""},
		{"POST", "/repos/:owner/:repo/labels", ""},

		{"PATCH", "/repos/:owner/:repo/labels/:name", ""},

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

		{"PATCH", "/repos/:owner/:repo/milestones/:number", ""},

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

		{"PATCH", "/orgs/:org", ""},

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

		{"PATCH", "/teams/:id", ""},

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

		{"PATCH", "/repos/:owner/:repo/pulls/:number", ""},

		{"GET", "/repos/:owner/:repo/pulls/:number/commits", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", ""},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", ""},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", ""},

		{"GET", "/repos/:owner/:repo/pulls/comments", ""},
		{"GET", "/repos/:owner/:repo/pulls/comments/:number", ""},

		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", ""},

		{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", ""},
		{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", ""},

		// Repositories
		{"GET", "/user/repos", ""},
		{"GET", "/users/:user/repos", ""},
		{"GET", "/orgs/:org/repos", ""},
		{"GET", "/repositories", ""},
		{"POST", "/user/repos", ""},
		{"POST", "/orgs/:org/repos", ""},
		{"GET", "/repos/:owner/:repo", ""},

		{"PATCH", "/repos/:owner/:repo", ""},

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

		{"PATCH", "/repos/:owner/:repo/comments/:id", ""},

		{"DELETE", "/repos/:owner/:repo/comments/:id", ""},
		{"GET", "/repos/:owner/:repo/commits", ""},
		{"GET", "/repos/:owner/:repo/commits/:sha", ""},
		{"GET", "/repos/:owner/:repo/readme", ""},

		//{"GET", "/repos/:owner/:repo/contents/*path", ""},
		//{"PUT", "/repos/:owner/:repo/contents/*path", ""},
		//{"DELETE", "/repos/:owner/:repo/contents/*path", ""},

		{"GET", "/repos/:owner/:repo/:archive_format/:ref", ""},

		{"GET", "/repos/:owner/:repo/keys", ""},
		{"GET", "/repos/:owner/:repo/keys/:id", ""},
		{"POST", "/repos/:owner/:repo/keys", ""},

		{"PATCH", "/repos/:owner/:repo/keys/:id", ""},

		{"DELETE", "/repos/:owner/:repo/keys/:id", ""},
		{"GET", "/repos/:owner/:repo/downloads", ""},
		{"GET", "/repos/:owner/:repo/downloads/:id", ""},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", ""},
		{"GET", "/repos/:owner/:repo/forks", ""},
		{"POST", "/repos/:owner/:repo/forks", ""},
		{"GET", "/repos/:owner/:repo/hooks", ""},
		{"GET", "/repos/:owner/:repo/hooks/:id", ""},
		{"POST", "/repos/:owner/:repo/hooks", ""},

		{"PATCH", "/repos/:owner/:repo/hooks/:id", ""},

		{"POST", "/repos/:owner/:repo/hooks/:id/tests", ""},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", ""},
		{"POST", "/repos/:owner/:repo/merges", ""},
		{"GET", "/repos/:owner/:repo/releases", ""},
		{"GET", "/repos/:owner/:repo/releases/:id", ""},
		{"POST", "/repos/:owner/:repo/releases", ""},

		{"PATCH", "/repos/:owner/:repo/releases/:id", ""},

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

		{"PATCH", "/user", ""},

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

		{"PATCH", "/user/keys/:id", ""},

		{"DELETE", "/user/keys/:id", ""},
	}

	parseAPI = []testRoute{
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

	googlePlusAPI = []testRoute{
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

	paramAndAnyAPI = []testRoute{
		{"GET", "/root/:first/foo/*", ""},
		{"GET", "/root/:first/:second/*", ""},
		{"GET", "/root/:first/bar/:second/*", ""},
		{"GET", "/root/:first/qux/:second/:third/:fourth", ""},
		{"GET", "/root/:first/qux/:second/:third/:fourth/*", ""},
		{"GET", "/root/*", ""},

		{"POST", "/root/:first/foo/*", ""},
		{"POST", "/root/:first/:second/*", ""},
		{"POST", "/root/:first/bar/:second/*", ""},
		{"POST", "/root/:first/qux/:second/:third/:fourth", ""},
		{"POST", "/root/:first/qux/:second/:third/:fourth/*", ""},
		{"POST", "/root/*", ""},

		{"PUT", "/root/:first/foo/*", ""},
		{"PUT", "/root/:first/:second/*", ""},
		{"PUT", "/root/:first/bar/:second/*", ""},
		{"PUT", "/root/:first/qux/:second/:third/:fourth", ""},
		{"PUT", "/root/:first/qux/:second/:third/:fourth/*", ""},
		{"PUT", "/root/*", ""},

		{"DELETE", "/root/:first/foo/*", ""},
		{"DELETE", "/root/:first/:second/*", ""},
		{"DELETE", "/root/:first/bar/:second/*", ""},
		{"DELETE", "/root/:first/qux/:second/:third/:fourth", ""},
		{"DELETE", "/root/:first/qux/:second/:third/:fourth/*", ""},
		{"DELETE", "/root/*", ""},
	}

	paramAndAnyAPIToFind = []testRoute{
		{"GET", "/root/one/foo/after/the/asterisk", ""},
		{"GET", "/root/one/foo/path/after/the/asterisk", ""},
		{"GET", "/root/one/two/path/after/the/asterisk", ""},
		{"GET", "/root/one/bar/two/after/the/asterisk", ""},
		{"GET", "/root/one/qux/two/three/four", ""},
		{"GET", "/root/one/qux/two/three/four/after/the/asterisk", ""},

		{"POST", "/root/one/foo/after/the/asterisk", ""},
		{"POST", "/root/one/foo/path/after/the/asterisk", ""},
		{"POST", "/root/one/two/path/after/the/asterisk", ""},
		{"POST", "/root/one/bar/two/after/the/asterisk", ""},
		{"POST", "/root/one/qux/two/three/four", ""},
		{"POST", "/root/one/qux/two/three/four/after/the/asterisk", ""},

		{"PUT", "/root/one/foo/after/the/asterisk", ""},
		{"PUT", "/root/one/foo/path/after/the/asterisk", ""},
		{"PUT", "/root/one/two/path/after/the/asterisk", ""},
		{"PUT", "/root/one/bar/two/after/the/asterisk", ""},
		{"PUT", "/root/one/qux/two/three/four", ""},
		{"PUT", "/root/one/qux/two/three/four/after/the/asterisk", ""},

		{"DELETE", "/root/one/foo/after/the/asterisk", ""},
		{"DELETE", "/root/one/foo/path/after/the/asterisk", ""},
		{"DELETE", "/root/one/two/path/after/the/asterisk", ""},
		{"DELETE", "/root/one/bar/two/after/the/asterisk", ""},
		{"DELETE", "/root/one/qux/two/three/four", ""},
		{"DELETE", "/root/one/qux/two/three/four/after/the/asterisk", ""},
	}

	missesAPI = []testRoute{
		{"GET", "/missOne", ""},
		{"GET", "/miss/two", ""},
		{"GET", "/miss/three/levels", ""},
		{"GET", "/miss/four/levels/nooo", ""},

		{"POST", "/missOne", ""},
		{"POST", "/miss/two", ""},
		{"POST", "/miss/three/levels", ""},
		{"POST", "/miss/four/levels/nooo", ""},

		{"PUT", "/missOne", ""},
		{"PUT", "/miss/two", ""},
		{"PUT", "/miss/three/levels", ""},
		{"PUT", "/miss/four/levels/nooo", ""},

		{"DELETE", "/missOne", ""},
		{"DELETE", "/miss/two", ""},
		{"DELETE", "/miss/three/levels", ""},
		{"DELETE", "/miss/four/levels/nooo", ""},
	}

	// handlerHelper created a function that will set a context key for assertion
	handlerHelper = func(key string, value int) func(c *Context) error {
		return func(c *Context) error {
			c.Set(key, value)
			c.Set("path", c.RouteInfo().Path)
			return nil
		}
	}
	handlerFunc = func(c *Context) error {
		c.Set("path", c.RouteInfo().Path)
		return nil
	}
)

func checkUnusedParamValues(t *testing.T, c *Context, expectParam map[string]string) {
	for _, p := range c.PathValues() {
		value := p.Value
		if value != "" {
			if expectParam == nil {
				t.Errorf("pValue '%v' is set for param name '%v' but we are not expecting it with expectParam", value, p)
			} else {
				if _, ok := expectParam[p.Name]; !ok {
					t.Errorf("pValue '%v' is set for param name '%v' but we are not expecting it with expectParam", value, p)
				}
			}
		}
	}
}

func TestRouterFillsRequestPatternField(t *testing.T) {
	path := "/folders/a/files/echo.gif"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	e := New()
	e.GET(path, handlerFunc)

	c := e.NewContext(req, rec)
	_ = e.router.Route(c)

	assert.Equal(t, path, c.Path())
	assert.Equal(t, path, c.Request().Pattern)
}

func TestRouterStatic(t *testing.T) {
	path := "/folders/a/files/echo.gif"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	e := New()
	e.GET(path, handlerFunc)

	c := e.NewContext(req, rec)
	_ = e.router.Route(c)

	assert.Equal(t, path, c.Path())
	assert.Equal(t, 0, cap(*c.pathValues))
	assert.Len(t, *c.pathValues, 0)
}

func TestRouterParam(t *testing.T) {
	e := New()

	e.GET("/users/:id", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		name        string
		whenURL     string
	}{
		{
			name:        "route /users/1 to /users/:id",
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			name:        "route /users/1/ to /users/:id",
			whenURL:     "/users/1/",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1/"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := e.NewContext(nil, nil)
			c.SetRequest(httptest.NewRequest(http.MethodGet, tc.whenURL, nil))
			_ = e.router.Route(c)

			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouter_addAndMatchAllSupportedMethods(t *testing.T) {
	var testCases = []struct {
		name            string
		whenMethod      string
		expectPath      string
		expectError     string
		givenNoAddRoute bool
	}{
		{name: "ok, CONNECT", whenMethod: http.MethodConnect},
		{name: "ok, DELETE", whenMethod: http.MethodDelete},
		{name: "ok, GET", whenMethod: http.MethodGet},
		{name: "ok, HEAD", whenMethod: http.MethodHead},
		{name: "ok, OPTIONS", whenMethod: http.MethodOptions},
		{name: "ok, PATCH", whenMethod: http.MethodPatch},
		{name: "ok, POST", whenMethod: http.MethodPost},
		{name: "ok, PROPFIND", whenMethod: PROPFIND},
		{name: "ok, PUT", whenMethod: http.MethodPut},
		{name: "ok, TRACE", whenMethod: http.MethodTrace},
		{name: "ok, REPORT", whenMethod: REPORT},
		{name: "ok, NON_TRADITIONAL_METHOD", whenMethod: "NON_TRADITIONAL_METHOD"},
		{
			name:            "ok, NOT_EXISTING_METHOD",
			whenMethod:      "NOT_EXISTING_METHOD",
			givenNoAddRoute: true,
			expectPath:      "/*",
			expectError:     "Method Not Allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/*", handlerFunc)
			if !tc.givenNoAddRoute {
				e.Add(tc.whenMethod, "/my/*", handlerFunc)
			}

			req := httptest.NewRequest(tc.whenMethod, "/my/some-url", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := e.router.Route(c)
			err := handler(c)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			expectPath := "/my/*"
			if tc.expectPath != "" {
				expectPath = tc.expectPath
			}
			assert.Equal(t, expectPath, c.Path())
		})
	}
}

func TestRouterAllowHeaderForAnyOtherMethodType(t *testing.T) {
	e := New()
	r := e.router

	_, err := r.Add(Route{Method: http.MethodGet, Path: "/users", Handler: handlerFunc})
	assert.NoError(t, err)
	_, err = r.Add(Route{Method: "COPY", Path: "/users", Handler: handlerFunc})
	assert.NoError(t, err)
	_, err = r.Add(Route{Method: "LOCK", Path: "/users", Handler: handlerFunc})
	assert.NoError(t, err)

	req := httptest.NewRequest("TEST", "/users", nil)
	rec := httptest.NewRecorder()

	//r.Find("TEST", "/users", c)
	c := e.NewContext(req, rec)

	handler := e.router.Route(c)
	err = handler(c)

	assert.EqualError(t, err, "Method Not Allowed")
	assert.ElementsMatch(t, []string{"COPY", "GET", "LOCK", "OPTIONS"}, strings.Split(c.Response().Header().Get(HeaderAllow), ", "))
}

func TestMethodNotAllowedAndNotFound(t *testing.T) {
	e := New()

	// Routes
	ri, err := e.AddRoute(Route{Method: http.MethodGet, Path: "/*", Handler: handlerFunc})
	assert.NoError(t, err)
	assert.Equal(t, "GET:/*", ri.Name)

	ri, err = e.AddRoute(Route{Method: http.MethodPost, Path: "/users/:id", Handler: handlerFunc})
	assert.NoError(t, err)
	assert.Equal(t, "POST:/users/:id", ri.Name)

	var testCases = []struct {
		name              string
		whenMethod        string
		whenURL           string
		expectRoute       string
		expectParam       map[string]string
		expectError       error
		expectAllowHeader string
	}{
		{
			name:        "exact match for route+method",
			whenMethod:  http.MethodPost,
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			name:              "matches node but not method. sends 405 from best match node",
			whenMethod:        http.MethodPut,
			whenURL:           "/users/1",
			expectRoute:       "/users/:id",
			expectError:       ErrMethodNotAllowed,
			expectAllowHeader: "OPTIONS, POST",
		},
		{
			name:        "best match is any route up in tree",
			whenMethod:  http.MethodGet,
			whenURL:     "/users/1",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "users/1"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			req := httptest.NewRequest(method, tc.whenURL, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
			assert.Equal(t, tc.expectAllowHeader, c.Response().Header().Get(HeaderAllow))
		})
	}
}

func TestRouterOptionsMethodHandler(t *testing.T) {
	e := New()

	var keyInContext any
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			err := next(c)
			keyInContext = c.Get(ContextKeyHeaderAllow)
			return err
		}
	})
	e.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "Echo!")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "OPTIONS, GET", rec.Header().Get(HeaderAllow))
	assert.Equal(t, "OPTIONS, GET", keyInContext)
}

func TestRouterHandleMethodOptions(t *testing.T) {
	e := New()
	e.contextPathParamAllocSize.Store(1)
	r := e.router

	r.Add(Route{Method: http.MethodGet, Path: "/users", Handler: handlerFunc})
	r.Add(Route{Method: http.MethodPost, Path: "/users", Handler: handlerFunc})
	r.Add(Route{Method: http.MethodPut, Path: "/users/:id", Handler: handlerFunc})
	r.Add(Route{Method: http.MethodGet, Path: "/users/:id", Handler: handlerFunc})

	var testCases = []struct {
		name              string
		whenMethod        string
		whenURL           string
		expectAllowHeader string
		expectStatus      int
	}{
		{
			name:              "allows GET and POST handlers",
			whenMethod:        http.MethodOptions,
			whenURL:           "/users",
			expectAllowHeader: "OPTIONS, GET, POST",
			expectStatus:      http.StatusNoContent,
		},
		{
			name:              "allows GET and PUT handlers",
			whenMethod:        http.MethodOptions,
			whenURL:           "/users/1",
			expectAllowHeader: "OPTIONS, GET, PUT",
			expectStatus:      http.StatusNoContent,
		},
		{
			name:              "GET does not have allows header",
			whenMethod:        http.MethodGet,
			whenURL:           "/users",
			expectAllowHeader: "",
			expectStatus:      http.StatusOK,
		},
		{
			name:              "path with no handlers does not set Allows header",
			whenMethod:        http.MethodOptions,
			whenURL:           "/notFound",
			expectAllowHeader: "",
			expectStatus:      http.StatusNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := r.Route(c)
			err := h(c)

			if tc.expectStatus >= 400 {
				assert.Error(t, err)
				he := err.(HTTPStatusCoder)
				assert.Equal(t, tc.expectStatus, he.StatusCode())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectStatus, rec.Code)
			}
			assert.Equal(t, tc.expectAllowHeader, c.Response().Header().Get(HeaderAllow))
		})
	}
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	e.GET("/users/:uid/files/:fid", handlerFunc)

	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/users/1/files/1", nil), nil)
	_ = e.router.Route(c)

	assert.Equal(t, "/users/:uid/files/:fid", c.Path())
	assert.Equal(t, "1", c.pathValues.GetOr("uid", ""))
	assert.Equal(t, "1", c.pathValues.GetOr("fid", ""))
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	e := New()

	e.GET("/a/:b/c/d/:e", handlerFunc)
	e.GET("/a/:b/c/:d/:f", handlerFunc)

	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/a/1/c/d/2/3", nil), nil)
	// `2/3` should mapped to path `/a/:b/c/d/:e` and into `:e`
	handler := e.router.Route(c)

	err := handler(c)
	assert.Equal(t, "/a/:b/c/d/:e", c.Path())
	assert.NoError(t, err)
}

// Issue #1754 - router needs to backtrack multiple levels upwards in tree to find the matching route
// route evaluation order
//
// Routes:
// 1) /a/:b/c
// 2) /a/c/d
// 3) /a/c/df
//
// 4) /a/*/f
// 5) /:e/c/f
//
// 6) /*
//
// Searching route for "/a/c/f" should match "/a/*/f"
// When route `4) /a/*/f` is not added then request for "/a/c/f" should match "/:e/c/f"
//
//	              +----------+
//	        +-----+ "/" root +--------------------+--------------------------+
//	        |     +----------+                    |                          |
//	        |                                     |                          |
//	+-------v-------+                         +---v---------+        +-------v---+
//	| "a/" (static) +---------------+         | ":" (param) |        | "*" (any) |
//	+-+----------+--+               |         +-----------+-+        +-----------+
//	  |          |                  |                     |
//
// +---------------v+  +-- ---v------+    +------v----+          +-----v-----------+
// | "c/d" (static) |  | ":" (param) |    | "*" (any) |          | "/c/f" (static) |
// +---------+------+  +--------+----+    +----------++          +-----------------+
//
//	|                  |                    |
//	|                  |                    |
//
// +---------v----+      +------v--------+    +------v--------+
// | "f" (static) |      | "/c" (static) |    | "/f" (static) |
// +--------------+      +---------------+    +---------------+
func TestRouteMultiLevelBacktracking(t *testing.T) {
	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		name        string
		whenURL     string
	}{
		{
			name:        "route /a/c/df to /a/c/df",
			whenURL:     "/a/c/df",
			expectRoute: "/a/c/df",
		},
		{
			name:        "route /a/x/df to /a/:b/c",
			whenURL:     "/a/x/c",
			expectRoute: "/a/:b/c",
			expectParam: map[string]string{"b": "x"},
		},
		{
			name:        "route /a/x/f to /a/*/f",
			whenURL:     "/a/x/f",
			expectRoute: "/a/*/f",
			expectParam: map[string]string{"*": "x/f"}, // NOTE: `x` would be probably more suitable
		},
		{
			name:        "route /b/c/f to /:e/c/f",
			whenURL:     "/b/c/f",
			expectRoute: "/:e/c/f",
			expectParam: map[string]string{"e": "b"},
		},
		{
			name:        "route /b/c/c to /*",
			whenURL:     "/b/c/c",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "b/c/c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/a/:b/c", handlerFunc)
			e.GET("/a/c/d", handlerFunc)
			e.GET("/a/c/df", handlerFunc)
			e.GET("/a/*/f", handlerFunc)
			e.GET("/:e/c/f", handlerFunc)
			e.GET("/*", handlerFunc)

			c := e.NewContext(httptest.NewRequest(http.MethodGet, tc.whenURL, nil), nil)
			_ = e.router.Route(c)

			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1754 - router needs to backtrack multiple levels upwards in tree to find the matching route
// route evaluation order
//
// Request for "/a/c/f" should match "/:e/c/f"
//
//	                       +-0,7--------+
//	                       | "/" (root) |----------------------------------+
//	                       +------------+                                  |
//	                            |      |                                   |
//	                            |      |                                   |
//	        +-1,6-----------+   |      |          +-8-----------+   +------v----+
//	        | "a/" (static) +<--+      +--------->+ ":" (param) |   | "*" (any) |
//	        +---------------+                     +-------------+   +-----------+
//	           |          |                             |
//	+-2--------v-----+   +v-3,5--------+       +-9------v--------+
//	| "c/d" (static) |   | ":" (param) |       | "/c/f" (static) |
//	+----------------+   +-------------+       +-----------------+
//	                      |
//	                 +-4--v----------+
//	                 | "/c" (static) |
//	                 +---------------+
func TestRouteMultiLevelBacktracking2(t *testing.T) {
	e := New()

	e.GET("/a/:b/c", handlerFunc)
	e.GET("/a/c/d", handlerFunc)
	e.GET("/a/c/df", handlerFunc)
	e.GET("/:e/c/f", handlerFunc)
	e.GET("/*", handlerFunc)

	var testCases = []struct {
		expectParam map[string]string
		name        string
		whenURL     string
		expectRoute string
	}{
		{
			name:        "route /a/c/df to /a/c/df",
			whenURL:     "/a/c/df",
			expectRoute: "/a/c/df",
		},
		{
			name:        "route /a/x/df to /a/:b/c",
			whenURL:     "/a/x/c",
			expectRoute: "/a/:b/c",
			expectParam: map[string]string{"b": "x"},
		},
		{
			name:        "route /a/c/f to /:e/c/f",
			whenURL:     "/a/c/f",
			expectRoute: "/:e/c/f",
			expectParam: map[string]string{"e": "a"},
		},
		{
			name:        "route /b/c/f to /:e/c/f",
			whenURL:     "/b/c/f",
			expectRoute: "/:e/c/f",
			expectParam: map[string]string{"e": "b"},
		},
		{
			name:        "route /b/c/c to /*",
			whenURL:     "/b/c/c",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "b/c/c"},
		},
		{ // this traverses `/a/:b/c` and `/:e/c/f` branches and eventually backtracks to `/*`
			name:        "route /a/c/cf to /*",
			whenURL:     "/a/c/cf",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "a/c/cf"},
		},
		{
			name:        "route /anyMatch to /*",
			whenURL:     "/anyMatch",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "anyMatch"},
		},
		{
			name:        "route /anyMatch/withSlash to /*",
			whenURL:     "/anyMatch/withSlash",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "anyMatch/withSlash"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := e.NewContext(httptest.NewRequest(http.MethodGet, tc.whenURL, nil), nil)
			_ = e.router.Route(c)

			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterBacktrackingFromMultipleParamKinds(t *testing.T) {
	e := New()

	e.GET("/*", handlerFunc) // this can match only path that does not have slash in it
	e.GET("/:1/second", handlerFunc)
	e.GET("/:1/:2", handlerFunc) // this acts as match ANY for all routes that have at least one slash
	e.GET("/:1/:2/third", handlerFunc)
	e.GET("/:1/:2/:3/fourth", handlerFunc)
	e.GET("/:1/:2/:3/:4/fifth", handlerFunc)

	var testCases = []struct {
		expectParam map[string]string
		name        string
		whenURL     string
		expectRoute string
	}{
		{
			name:        "route /first to /*",
			whenURL:     "/first",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "first"},
		},
		{
			name:        "route /first/second to /:1/second",
			whenURL:     "/first/second",
			expectRoute: "/:1/second",
			expectParam: map[string]string{"1": "first"},
		},
		{
			name:        "route /first/second-new to /:1/:2",
			whenURL:     "/first/second-new",
			expectRoute: "/:1/:2",
			expectParam: map[string]string{
				"1": "first",
				"2": "second-new",
			},
		},
		{ // FIXME: should match `/:1/:2` when backtracking in tree. this 1 level backtracking fails even with old implementation
			name:        "route /first/second/ to /:1/:2",
			whenURL:     "/first/second/",                        /// <-- slash at the end is problematic
			expectRoute: "/*",                                    // "/:1/:2",
			expectParam: map[string]string{"*": "first/second/"}, // map[string]string{"1": "first", "2": "second/"},
		},
		{ // FIXME: should match `/:1/:2`. same backtracking problem. when backtracking is at `/:1/:2` during backtracking this node should be match as it has executable handler
			name:        "route /first/second/third/fourth/fifth/nope to /:1/:2",
			whenURL:     "/first/second/third/fourth/fifth/nope",
			expectRoute: "/*",                                                           // "/:1/:2",
			expectParam: map[string]string{"*": "first/second/third/fourth/fifth/nope"}, // map[string]string{"1": "first", "2": "second/third/fourth/fifth/nope"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := e.NewContext(httptest.NewRequest(http.MethodGet, tc.whenURL, nil), nil)
			_ = e.router.Route(c)

			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1509
func TestRouterParamStaticConflict(t *testing.T) {
	e := New()

	g := e.Group("/g")
	g.GET("/skills", handlerFunc)
	g.GET("/status", handlerFunc)
	g.GET("/:name", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		whenURL     string
	}{
		{
			whenURL:     "/g/s",
			expectRoute: "/g/:name",
			expectParam: map[string]string{"name": "s"},
		},
		{
			whenURL:     "/g/status",
			expectRoute: "/g/status",
			expectParam: map[string]string{"name": ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(httptest.NewRequest(http.MethodGet, tc.whenURL, nil), nil)

			handler := e.router.Route(c)

			assert.NoError(t, handler(c))
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterParam_escapeColon(t *testing.T) {
	// to allow Google cloud API like route paths with colon in them
	// i.e. https://service.name/v1/some/resource/name:customVerb <- that `:customVerb` is not path param. It is just a string
	e := New()

	e.POST("/files/a/long/file\\:undelete", handlerFunc)
	e.POST("/multilevel\\:undelete/second\\:something", handlerFunc)
	e.POST("/mixed/:id/second\\:something", handlerFunc)
	e.POST("/v1/some/resource/name:customVerb", handlerFunc)

	var testCases = []struct {
		whenURL     string
		expectRoute string
		expectParam map[string]string
		expectError string
	}{
		{
			whenURL:     "/files/a/long/file:undelete",
			expectRoute: "/files/a/long/file\\:undelete",
			expectParam: map[string]string{},
		},
		{
			whenURL:     "/multilevel:undelete/second:something",
			expectRoute: "/multilevel\\:undelete/second\\:something",
			expectParam: map[string]string{},
		},
		{
			whenURL:     "/mixed/123/second:something",
			expectRoute: "/mixed/:id/second\\:something",
			expectParam: map[string]string{"id": "123"},
		},
		{
			whenURL:     "/files/a/long/file:notMatching",
			expectRoute: "",
			expectError: "Not Found",
			expectParam: nil,
		},
		{
			whenURL:     "/v1/some/resource/name:PATCH",
			expectRoute: "/v1/some/resource/name:customVerb",
			expectParam: map[string]string{"customVerb": ":PATCH"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)

			assert.Equal(t, tc.expectRoute, c.Path())
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterMatchAny(t *testing.T) {
	e := New()

	// Routes
	e.GET("/", handlerFunc)
	e.GET("/*", handlerFunc)
	e.GET("/users/*", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		whenURL     string
	}{
		{
			whenURL:     "/",
			expectRoute: "/",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/download",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "download"},
		},
		{
			whenURL:     "/users/joe",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": "joe"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(httptest.NewRequest(http.MethodGet, tc.whenURL, nil), nil)

			handler := e.router.Route(c)

			assert.NoError(t, handler(c))
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// NOTE: this is to document current implementation. Last added route with `*` asterisk is always the match and no
// backtracking or more precise matching is done to find more suitable match.
//
// Current behaviour might not be correct or expected.
// But this is where we are without well defined requirements/rules how (multiple) asterisks work in route
func TestRouterAnyMatchesLastAddedAnyRoute(t *testing.T) {
	e := New()

	e.GET("/users/*", handlerFunc)
	e.GET("/users/*/action*", handlerFunc)

	req := httptest.NewRequest(http.MethodGet, "/users/xxx/action/sea", nil)
	c := e.NewContext(req, nil)
	handler := e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "/users/*/action*", c.Path())
	assert.Equal(t, "xxx/action/sea", c.pathValues.GetOr("*", ""))

	// if we add another route then it is the last added and so it is matched
	e.GET("/users/*/action/search", handlerFunc)

	c2 := e.NewContext(httptest.NewRequest(http.MethodGet, "/users/xxx/action/sea", nil), nil)
	handler2 := e.router.Route(c2)

	assert.NoError(t, handler2(c2))
	assert.Equal(t, "/users/*/action/search", c2.Path())
	assert.Equal(t, "xxx/action/sea", c2.pathValues.GetOr("*", ""))
}

// Issue #1739
func TestRouterMatchAnyPrefixIssue(t *testing.T) {
	e := New()

	// Routes
	e.GET("/*", handlerFunc)
	e.GET("/users/*", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		whenURL     string
	}{
		{
			whenURL:     "/",
			expectRoute: "/*",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/users",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "users"},
		},
		{
			whenURL:     "/users/",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/users_prefix",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "users_prefix"},
		},
		{
			whenURL:     "/users_prefix/",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "users_prefix/"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			assert.NoError(t, handler(c))
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// TestRouterMatchAnySlash shall verify finding the best route
// for any routes with trailing slash requests
func TestRouterMatchAnySlash(t *testing.T) {
	e := New()

	// Routes
	e.GET("/users", handlerFunc)
	e.GET("/users/*", handlerFunc)
	e.GET("/img/*", handlerFunc)
	e.GET("/img/load", handlerFunc)
	e.GET("/img/load/*", handlerFunc)
	e.GET("/assets/*", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/",
			expectRoute: "",
			expectParam: map[string]string{"*": ""},
			expectError: ErrNotFound,
		},
		{ // Test trailing slash request for simple any route (see #1526)
			whenURL:     "/users/",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/users/joe",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": "joe"},
		},
		// Test trailing slash request for nested any route (see #1526)
		{
			whenURL:     "/img/load",
			expectRoute: "/img/load",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/img/load/",
			expectRoute: "/img/load/*",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/img/load/ben",
			expectRoute: "/img/load/*",
			expectParam: map[string]string{"*": "ben"},
		},
		// Test /assets/* any route
		{ // ... without trailing slash must not match
			whenURL:     "/assets",
			expectRoute: "",
			expectParam: map[string]string{"*": ""},
			expectError: ErrNotFound,
		},

		{ // ... with trailing slash must match
			whenURL:     "/assets/",
			expectRoute: "/assets/*",
			expectParam: map[string]string{"*": ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	e := New()

	// Routes
	e.GET("/api/users/jack", handlerFunc)
	e.GET("/api/users/jill", handlerFunc)
	e.GET("/api/users/*", handlerFunc)
	e.GET("/api/*", handlerFunc)
	e.GET("/other/*", handlerFunc)
	e.GET("/*", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectError error
		expectParam map[string]string
		whenURL     string
	}{
		{
			whenURL:     "/api/users/jack",
			expectRoute: "/api/users/jack",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/api/users/jill",
			expectRoute: "/api/users/jill",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/api/users/joe",
			expectRoute: "/api/users/*",
			expectParam: map[string]string{"*": "joe"},
		},
		{
			whenURL:     "/api/nousers/joe",
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "nousers/joe"},
		},
		{
			whenURL:     "/api/none",
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "none"},
		},
		{
			whenURL:     "/api/none",
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "none"},
		},
		{
			whenURL:     "/noapi/users/jim",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "noapi/users/jim"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}
func TestRouterMatchAnyMultiLevelWithPost(t *testing.T) {
	e := New()

	// Routes
	e.POST("/api/auth/login", handlerFunc)
	e.POST("/api/auth/forgotPassword", handlerFunc)
	e.Any("/api/*", handlerFunc)
	e.Any("/*", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
	}{
		{ // POST /api/auth/login shall choose login method
			whenURL:     "/api/auth/login",
			whenMethod:  http.MethodPost,
			expectRoute: "/api/auth/login",
			expectParam: map[string]string{"*": ""},
		},
		{ // POST /api/auth/logout shall choose nearest any route
			whenURL:     "/api/auth/logout",
			whenMethod:  http.MethodPost,
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "auth/logout"},
		},
		{ // POST to /api/other/test shall choose nearest any route
			whenURL:     "/api/other/test",
			whenMethod:  http.MethodPost,
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "other/test"},
		},
		{ // GET to /api/other/test shall choose nearest any route
			whenURL:     "/api/other/test",
			whenMethod:  http.MethodGet,
			expectRoute: "/api/*",
			expectParam: map[string]string{"*": "other/test"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	e.GET("/:a/:b/:c", handlerFunc)

	req := httptest.NewRequest(http.MethodGet, "/1/2/3", nil)
	c := e.NewContext(req, nil)
	handler := e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "1", c.pathValues.GetOr("a", "---none---"))
	assert.Equal(t, "2", c.pathValues.GetOr("b", "---none---"))
	assert.Equal(t, "3", c.pathValues.GetOr("c", "---none---"))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	e := New()

	// Route
	e.GET("/users/:id/*", handlerFunc)

	req := httptest.NewRequest(http.MethodGet, "/users/joe/comments", nil)
	c := e.NewContext(req, nil)
	handler := e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "/users/:id/*", c.Path())
	assert.Equal(t, "joe", c.pathValues.GetOr("id", "---none---"))
	assert.Equal(t, "comments", c.pathValues.GetOr("*", "---none---"))
}

func TestRouterMultiRoute(t *testing.T) {
	e := New()

	// Routes
	e.GET("/users", handlerFunc)
	e.GET("/users/:id", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/users",
			expectRoute: "/users",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			whenURL:     "/user",
			expectRoute: "",
			expectParam: map[string]string{"*": ""},
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterPriority(t *testing.T) {
	e := New()

	// Routes
	e.GET("/users", handlerFunc)
	e.GET("/users/new", handlerFunc)
	e.GET("/users/:id", handlerFunc)
	e.GET("/users/dew", handlerFunc)
	e.GET("/users/:id/files", handlerFunc)
	e.GET("/users/newsee", handlerFunc)
	e.GET("/users/*", handlerFunc)
	e.GET("/users/new/*", handlerFunc)
	e.GET("/*", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
	}{
		{
			whenURL:     "/users",
			expectRoute: "/users",
		},
		{
			whenURL:     "/users/new",
			expectRoute: "/users/new",
		},
		{
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			whenURL:     "/users/dew",
			expectRoute: "/users/dew",
		},
		{
			whenURL:     "/users/1/files",
			expectRoute: "/users/:id/files",
			expectParam: map[string]string{"id": "1"},
		},
		{
			whenURL:     "/users/new",
			expectRoute: "/users/new",
		},
		{
			whenURL:     "/users/news",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "news"},
		},
		{
			whenURL:     "/users/newsee",
			expectRoute: "/users/newsee",
		},
		{
			whenURL:     "/users/joe/books",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": "joe/books"},
		},
		{
			whenURL:     "/users/new/someone",
			expectRoute: "/users/new/*",
			expectParam: map[string]string{"*": "someone"},
		},
		{
			whenURL:     "/users/dew/someone",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": "dew/someone"},
		},
		{ // Route > /users/* should be matched although /users/dew exists
			whenURL:     "/users/notexists/someone",
			expectRoute: "/users/*",
			expectParam: map[string]string{"*": "notexists/someone"},
		},
		{
			whenURL:     "/nousers",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "nousers"},
		},
		{
			whenURL:     "/nousers/new",
			expectRoute: "/*",
			expectParam: map[string]string{"*": "nousers/new"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterIssue1348(t *testing.T) {
	e := New()

	e.GET("/:lang/", handlerFunc)
	e.GET("/:lang/dupa", handlerFunc)
}

// Issue #372
func TestRouterPriorityNotFound(t *testing.T) {
	e := New()

	// Add
	e.GET("/a/foo", handlerFunc)
	e.GET("/a/bar", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/a/foo",
			expectRoute: "/a/foo",
		},
		{
			whenURL:     "/a/bar",
			expectRoute: "/a/bar",
		},
		{
			whenURL:     "/abc/def",
			expectRoute: "",
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterParamNames(t *testing.T) {
	e := New()

	// Routes
	e.GET("/users", handlerFunc)
	e.GET("/users/:id", handlerFunc)
	e.GET("/users/:uid/files/:fid", handlerFunc)

	var testCases = []struct {
		expectRoute any
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
	}{
		{
			whenURL:     "/users",
			expectRoute: "/users",
		},
		{
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			whenURL:     "/users/1/files/1",
			expectRoute: "/users/:uid/files/:fid",
			expectParam: map[string]string{
				"uid": "1",
				"fid": "1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #623 and #1406
func TestRouterStaticDynamicConflict(t *testing.T) {
	e := New()

	e.GET("/dictionary/skills", handlerFunc)
	e.GET("/dictionary/:name", handlerFunc)
	e.GET("/users/new", handlerFunc)
	e.GET("/users/:name", handlerFunc)
	e.GET("/server", handlerFunc)
	e.GET("/", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenMethod  string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/dictionary/skills",
			expectRoute: "/dictionary/skills",
			expectParam: map[string]string{"*": ""},
		},
		{
			whenURL:     "/dictionary/skillsnot",
			expectRoute: "/dictionary/:name",
			expectParam: map[string]string{"name": "skillsnot"},
		},
		{
			whenURL:     "/dictionary/type",
			expectRoute: "/dictionary/:name",
			expectParam: map[string]string{"name": "type"},
		},
		{
			whenURL:     "/server",
			expectRoute: "/server",
		},
		{
			whenURL:     "/users/new",
			expectRoute: "/users/new",
		},
		{
			whenURL:     "/users/new2",
			expectRoute: "/users/:name",
			expectParam: map[string]string{"name": "new2"},
		},
		{
			whenURL:     "/",
			expectRoute: "/",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1348
func TestRouterParamBacktraceNotFound(t *testing.T) {
	e := New()

	// Add
	e.GET("/:param1", handlerFunc)
	e.GET("/:param1/foo", handlerFunc)
	e.GET("/:param1/bar", handlerFunc)
	e.GET("/:param1/bar/:param2", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		name        string
		whenMethod  string
		whenURL     string
		expectRoute string
	}{
		{
			name:        "route /a to /:param1",
			whenURL:     "/a",
			expectRoute: "/:param1",
			expectParam: map[string]string{"param1": "a"},
		},
		{
			name:        "route /a/foo to /:param1/foo",
			whenURL:     "/a/foo",
			expectRoute: "/:param1/foo",
			expectParam: map[string]string{"param1": "a"},
		},
		{
			name:        "route /a/bar to /:param1/bar",
			whenURL:     "/a/bar",
			expectRoute: "/:param1/bar",
			expectParam: map[string]string{"param1": "a"},
		},
		{
			name:        "route /a/bar/b to /:param1/bar/:param2",
			whenURL:     "/a/bar/b",
			expectRoute: "/:param1/bar/:param2",
			expectParam: map[string]string{
				"param1": "a",
				"param2": "b",
			},
		},
		{
			name:        "route /a/bbbbb should return 404",
			whenURL:     "/a/bbbbb",
			expectRoute: "",
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}

			req := httptest.NewRequest(method, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func testRouterAPI(t *testing.T, api []testRoute) {
	e := New()

	for _, route := range api {
		ri, err := e.AddRoute(Route{
			Method: route.Method,
			Path:   route.Path,
			Handler: func(c *Context) error {
				return nil
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, ri)
	}

	c := e.NewContext(nil, nil)
	for _, route := range api {
		t.Run(route.Path, func(t *testing.T) {
			c.SetRequest(httptest.NewRequest(route.Method, route.Path, nil))
			e.router.Route(c)

			tokens := strings.Split(route.Path[1:], "/")
			for _, token := range tokens {
				if token[0] == ':' {
					assert.Equal(t, c.pathValues.GetOr(token[1:], "---none---"), token)
				}
			}
		})
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

func TestRouter_Match_DifferentParamNamesForSamePlace(t *testing.T) {
	var testCases = []struct {
		expectError error
		expectParam map[string]string
		name        string
		whenURL     string
		whenMethod  string
		expectRoute string
	}{
		{
			name:        "ok, 1=id + 2=file",
			whenURL:     "/users/123/file/payroll.csv",
			whenMethod:  http.MethodGet,
			expectRoute: "/users/:id/file/:file",
			expectParam: map[string]string{"id": "123", "file": "payroll.csv"},
		},
		{
			name:        "ok, 1=id2 + 2=file2",
			whenURL:     "/users/123/file/payroll.csv",
			whenMethod:  http.MethodPost,
			expectRoute: "/users/:id2/file/:file2",
			expectParam: map[string]string{"id2": "123", "file2": "payroll.csv"},
		},
		{
			name:        "ok, 1=uid",
			whenURL:     "/users/999/files",
			whenMethod:  http.MethodGet,
			expectRoute: "/users/:uid/files",
			expectParam: map[string]string{"uid": "999"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/users/create", handlerFunc)
			e.GET("/users/:id/file/:file", handlerFunc)
			e.POST("/users/:id2/file/:file2", handlerFunc)
			e.GET("/users/:uid/files", handlerFunc)

			req := httptest.NewRequest(tc.whenMethod, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #2164 - this test is meant to document path parameter behaviour when request url has empty value in place
// of the path parameter. As tests show the result is different depending on where parameter exists in the route path.
func TestDefaultRouter_PathValuesCanMatchEmptyValues(t *testing.T) {
	var testCases = []struct {
		expectError error
		expectParam map[string]string
		name        string
		whenURL     string
		expectRoute string
	}{
		{
			name:        "ok, route is matched with even empty param is in the middle and between slashes",
			whenURL:     "/a//b",
			expectRoute: "/a/:id/b",
			expectParam: map[string]string{"id": ""},
		},
		{
			name:        "ok, route is matched with even empty param is in the middle",
			whenURL:     "/a2/b",
			expectRoute: "/a2:id/b",
			expectParam: map[string]string{"id": ""},
		},
		{
			name:        "ok, route is NOT matched with even empty param is at the end",
			whenURL:     "/a3/",
			expectRoute: "",
			expectParam: map[string]string{},
			expectError: ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/a/:id/b", handlerFunc)
			e.GET("/a2:id/b", handlerFunc)
			e.GET("/a3/:id", handlerFunc)

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)
			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #729
func TestRouterParamAlias(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/users/:userID/following", ""},
		{http.MethodGet, "/users/:userID/followedBy", ""},
		{http.MethodGet, "/users/:userID/follow", ""},
	}
	testRouterAPI(t, api)
}

// Issue #1052
func TestRouterParamOrdering(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/:a/:b/:c/:id", ""},
		{http.MethodGet, "/:a/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []testRoute{
		{http.MethodGet, "/:a/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
		{http.MethodGet, "/:a/:b/:c/:id", ""},
	}
	testRouterAPI(t, api2)
	api3 := []testRoute{
		{http.MethodGet, "/:a/:b/:c/:id", ""},
		{http.MethodGet, "/:a/:e/:id", ""},
		{http.MethodGet, "/:a/:id", ""},
	}
	testRouterAPI(t, api3)
}

// Issue #1139
func TestRouterMixedParams(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/teacher/:tid/room/suggestions", ""},
		{http.MethodGet, "/teacher/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []testRoute{
		{http.MethodGet, "/teacher/:id", ""},
		{http.MethodGet, "/teacher/:tid/room/suggestions", ""},
	}
	testRouterAPI(t, api2)
}

// Issue #1466
func TestRouterParam1466(t *testing.T) {
	e := New()

	e.POST("/users/signup", handlerFunc)
	e.POST("/users/signup/bulk", handlerFunc)
	e.POST("/users/survey", handlerFunc)

	e.GET("/users/:username", handlerFunc)
	e.GET("/interests/:name/users", handlerFunc)
	e.GET("/skills/:name/users", handlerFunc)
	// Additional routes for Issue 1479
	e.GET("/users/:username/likes/projects/ids", handlerFunc)
	e.GET("/users/:username/profile", handlerFunc)
	e.GET("/users/:username/uploads/:type", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/users/ajitem",
			expectRoute: "/users/:username",
			expectParam: map[string]string{"username": "ajitem"},
		},
		{
			whenURL:     "/users/sharewithme",
			expectRoute: "/users/:username",
			expectParam: map[string]string{"username": "sharewithme"},
		},
		{ // route `/users/signup` is registered for POST. so param route `/users/:username` (lesser priority) is matched as it has GET handler
			whenURL:     "/users/signup",
			expectRoute: "/users/:username",
			expectParam: map[string]string{"username": "signup"},
		},
		// Additional assertions for #1479
		{
			whenURL:     "/users/sharewithme/likes/projects/ids",
			expectRoute: "/users/:username/likes/projects/ids",
			expectParam: map[string]string{"username": "sharewithme"},
		},
		{
			whenURL:     "/users/ajitem/likes/projects/ids",
			expectRoute: "/users/:username/likes/projects/ids",
			expectParam: map[string]string{"username": "ajitem"},
		},
		{
			whenURL:     "/users/sharewithme/profile",
			expectRoute: "/users/:username/profile",
			expectParam: map[string]string{"username": "sharewithme"},
		},
		{
			whenURL:     "/users/ajitem/profile",
			expectRoute: "/users/:username/profile",
			expectParam: map[string]string{"username": "ajitem"},
		},
		{
			whenURL:     "/users/sharewithme/uploads/self",
			expectRoute: "/users/:username/uploads/:type",
			expectParam: map[string]string{
				"username": "sharewithme",
				"type":     "self",
			},
		},
		{
			whenURL:     "/users/ajitem/uploads/self",
			expectRoute: "/users/:username/uploads/:type",
			expectParam: map[string]string{
				"username": "ajitem",
				"type":     "self",
			},
		},
		{
			whenURL:     "/users/tree/free",
			expectRoute: "", // not found
			expectParam: map[string]string{"id": ""},
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)

			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, ""))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestPathValuesSizeOverMultipleRequests(t *testing.T) {
	e := New()
	e.GET("/test/:id/:action", handlerFunc) // max params is 2

	req := httptest.NewRequest(http.MethodGet, "/test/1/a", nil)
	rec := httptest.NewRecorder()

	c := e.AcquireContext()
	c.Reset(req, rec)
	assert.Equal(t, 0, len(*c.pathValues)) // fresh context is empty
	assert.Equal(t, 2, cap(*c.pathValues)) // is set max path params amount

	// imitate some (pre)middleware changing/replacing pathparams to smaller size
	c.SetPathValues(PathValues{
		{Name: "id", Value: "1"},
	})
	assert.Equal(t, 1, len(*c.pathValues)) // as SetPathValues was provided
	assert.Equal(t, 2, cap(*c.pathValues)) // SetPathValues did not change that to smaller

	handler := e.router.Route(c)
	e.ReleaseContext(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, 2, len(*c.pathValues)) // matched route had 2 path params
	assert.Equal(t, 2, cap(*c.pathValues)) // was not changed
	assert.Equal(t, "1", c.Param("id"))
	assert.Equal(t, "a", c.Param("action"))
}

// Issue #1655
func TestRouterFindNotPanicOrLoopsWhenContextSetParamValuesIsCalledWithLessValuesThanEchoMaxParam(t *testing.T) {
	e := New()

	v0 := e.Group("/:version")
	v0.GET("/admin", func(c *Context) error {
		c.SetPathValues(PathValues{{
			Name:  "version",
			Value: "v1",
		}})
		return nil
	})

	v0.GET("/images/view/:id", handlerHelper("iv", 1))
	v0.GET("/images/:id", handlerHelper("i", 1))
	v0.GET("/view/*", handlerHelper("v", 1))

	//If this API is called before the next two one panic the other loops ( of course without my fix ;) )
	req := httptest.NewRequest(http.MethodGet, "/v1/admin", nil)
	c := e.NewContext(req, nil)
	handler := e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "v1", c.Param("version"))

	//panic
	req = httptest.NewRequest(http.MethodGet, "/v1/view/same-data", nil)
	c = e.NewContext(req, nil)
	handler = e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "same-data", c.Param("*"))
	assert.Equal(t, 1, c.Get("v"))

	//looping
	req = httptest.NewRequest(http.MethodGet, "/v1/images/view", nil)
	c = e.NewContext(req, nil)
	handler = e.router.Route(c)

	assert.NoError(t, handler(c))
	assert.Equal(t, "view", c.Param("id"))
	assert.Equal(t, 1, c.Get("i"))
}

// Issue #1653
func TestRouterPanicWhenParamNoRootOnlyChildsFailsFind(t *testing.T) {
	e := New()

	e.GET("/users/create", handlerFunc)
	e.GET("/users/:id/edit", handlerFunc)
	e.GET("/users/:id/active", handlerFunc)

	var testCases = []struct {
		expectError error
		expectParam map[string]string
		whenURL     string
		expectRoute string
	}{
		{
			whenURL:     "/users/alice/edit",
			expectRoute: "/users/:id/edit",
			expectParam: map[string]string{"id": "alice"},
		},
		{
			whenURL:     "/users/bob/active",
			expectRoute: "/users/:id/active",
			expectParam: map[string]string{"id": "bob"},
		},
		{
			whenURL:     "/users/create",
			expectRoute: "/users/create",
			expectParam: nil,
		},
		//This panic before the fix for Issue #1653
		{
			whenURL:     "/users/createNotFound",
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)

			handler := e.router.Route(c)

			err := handler(c)
			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.pathValues.GetOr(param, "---none---"))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouter_addEmptyPathToSlashReverse(t *testing.T) {
	r := NewRouter(RouterConfig{})
	_, err := r.Add(Route{Method: http.MethodGet, Path: "", Handler: handlerFunc, Name: "empty"})
	assert.NoError(t, err)

	reversed, err := r.Routes().Reverse("empty") // empty path is normalized to `/` internally but stays "" in route

	assert.NoError(t, err)
	assert.Equal(t, "", reversed)
}

func TestRouter_ReverseNotFound(t *testing.T) {
	r := NewRouter(RouterConfig{})
	_, err := r.Add(Route{Method: http.MethodGet, Path: "", Handler: handlerFunc, Name: "empty"})
	assert.NoError(t, err)

	reversed, err := r.Routes().Reverse("not-existing")

	assert.EqualError(t, err, "route not found")
	assert.Equal(t, "", reversed)
}

func TestRoutes_ReverseHandlerName(t *testing.T) {
	static := func(*Context) error { return nil }
	getUser := func(*Context) error { return nil }
	getAny := func(*Context) error { return nil }
	getFile := func(*Context) error { return nil }

	var testCases = []struct {
		name      string
		when      string
		expect    string
		expectErr string
		whenArgs  []any
	}{
		{
			name:     "ok, HandlerName + args",
			when:     HandlerName(getFile),
			whenArgs: []any{"1"},
			expect:   "/group/users/1/files/:fid",
		},
		{
			name:   "ok, HandlerName",
			when:   HandlerName(getFile),
			expect: "/group/users/:uid/files/:fid",
		},
		{
			name:   "ok, unnamed fixed name",
			when:   "GET:/static/file",
			expect: "/static/file",
		},
		{
			name:   "ok, unnamed fixed name",
			when:   "GET:/users/:id",
			expect: "/users/:id",
		},
		{
			name:   "ok, unnamed any route",
			when:   RouteAny + ":/documents/*",
			expect: "/documents/*",
		},
		{
			name:     "ok, unnamed any route + args",
			when:     RouteAny + ":/documents/*",
			whenArgs: []any{"index.html"},
			expect:   "/documents/index.html",
		},
		{
			name:     "ok, named route + args",
			when:     HandlerName(getFile),
			whenArgs: []any{"1", "abc"},
			expect:   "/group/users/1/files/abc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.GET("/static/file", static)
			e.GET("/users/:id", getUser)
			e.Any("/documents/*", getAny)

			g := e.Group("/group")
			g.GET("/roles/:rid/files/:fid", getFile)
			g.AddRoute(Route{Method: http.MethodGet, Path: "/users/:uid/files/:fid", Handler: getFile, Name: HandlerName(getFile)})

			reversed, err := e.Router().Routes().Reverse(tc.when, tc.whenArgs...)

			assert.Equal(t, tc.expect, reversed)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoutes_Reverse(t *testing.T) {
	var testCases = []struct {
		name      string
		when      string
		expect    string
		expectErr string
		whenArgs  []any
	}{
		{
			name:   "ok, static",
			when:   "/static",
			expect: "/static",
		},
		{
			name:     "ok, static + args",
			when:     "/static",
			whenArgs: []any{"missing param"},
			expect:   "/static",
		},
		{
			name:   "ok, static/*",
			when:   "/static/*",
			expect: "/static/*",
		},
		{
			name:     "ok, static/*",
			when:     "/static/*",
			whenArgs: []any{"foo.txt"},
			expect:   "/static/foo.txt",
		},
		{
			name:   "ok, /params/:foo",
			when:   "/params/:foo",
			expect: "/params/:foo",
		},
		{
			name:     "ok, /params/:foo + args",
			when:     "/params/:foo",
			whenArgs: []any{"one"},
			expect:   "/params/one",
		},
		{
			name:   "ok, /params/:foo/bar/:qux",
			when:   "/params/:foo/bar/:qux",
			expect: "/params/:foo/bar/:qux",
		},
		{
			name:     "ok, /params/:foo/bar/:qux + args1",
			when:     "/params/:foo/bar/:qux",
			whenArgs: []any{"one"},
			expect:   "/params/one/bar/:qux",
		},
		{
			name:     "ok, /params/:foo/bar/:qux + args2",
			when:     "/params/:foo/bar/:qux",
			whenArgs: []any{"one", "two"},
			expect:   "/params/one/bar/two",
		},
		{
			name:     "ok, /params/:foo/bar/:qux/* + args3",
			when:     "/params/:foo/bar/:qux/*",
			whenArgs: []any{"one", "two", "three"},
			expect:   "/params/one/bar/two/three",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dummyHandler := func(*Context) error { return nil }

			router := NewRouter(RouterConfig{})
			router.Add(Route{Path: "/static", Name: "/static", Method: http.MethodGet, Handler: dummyHandler})
			router.Add(Route{Path: "/static/*", Name: "/static/*", Method: http.MethodGet, Handler: dummyHandler})
			router.Add(Route{Path: "/params/:foo", Name: "/params/:foo", Method: http.MethodGet, Handler: dummyHandler})
			router.Add(Route{Path: "/params/:foo/bar/:qux", Name: "/params/:foo/bar/:qux", Method: http.MethodGet, Handler: dummyHandler})
			router.Add(Route{Path: "/params/:foo/bar/:qux/*", Name: "/params/:foo/bar/:qux/*", Method: http.MethodGet, Handler: dummyHandler})

			reversed, err := router.Routes().Reverse(tc.when, tc.whenArgs...)
			assert.Equal(t, tc.expect, reversed)
			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRouter_Routes(t *testing.T) {
	routes := []*Route{
		{Method: http.MethodGet, Path: "/users/:user/events"},
		{Method: http.MethodGet, Path: "/users/:user/events/public"},
		{Method: http.MethodPost, Path: "/repos/:owner/:repo/git/refs"},
		{Method: http.MethodPost, Path: "/repos/:owner/:repo/git/tags"},
	}
	router := NewRouter(RouterConfig{})
	for _, r := range routes {
		_, err := router.Add(Route{
			Method: r.Method,
			Path:   r.Path,
			Handler: func(c *Context) error {
				return c.String(http.StatusOK, "OK")
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	assert.Equal(t, len(routes), len(router.Routes()))
	for _, r := range router.Routes() {
		found := false
		for _, rr := range routes {
			if r.Method == rr.Method && r.Path == rr.Path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Route %s %s not found", r.Method, r.Path)
		}
	}
}

func TestRouterNoRoutablePath(t *testing.T) {
	e := New()

	e.router.Add(Route{Path: "/static", Name: "/static", Method: http.MethodGet, Handler: func(*Context) error { return nil }})

	req := httptest.NewRequest(http.MethodGet, "/notFound", nil)
	c := e.NewContext(req, nil)

	e.router.Route(c)
	// No routable path, don't set Path.
	assert.Equal(t, "", c.Path())
}

func benchmarkRouterRoutes(b *testing.B, routes []testRoute, routesToFind []testRoute) {
	e := New()
	r := e.router
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		e.AddRoute(Route{
			Method: route.Method,
			Path:   route.Path,
			Handler: func(c *Context) error {
				return nil
			},
		})
	}

	// Routes adding are performed just once, so it doesn't make sense to see that in the benchmark
	b.ResetTimer()

	// Find routes
	for i := 0; i < b.N; i++ {
		for _, route := range routesToFind {
			c := e.contextPool.Get().(*Context)
			c.request = req

			req.Method = route.Method
			req.URL.Path = route.Path

			_ = r.Route(c)

			e.contextPool.Put(c)
		}
	}
}

func TestDefaultRouter_Remove(t *testing.T) {
	var testCases = []struct {
		name       string
		whenMethod string
		whenPath   string
		expectErr  string
		givenPaths []string
	}{
		{
			name:     "ok, static",
			whenPath: "/users",
		},
		{
			name:     "ok, static without slash",
			whenPath: "users",
		},
		{
			name:     "ok, multilevel static",
			whenPath: "/users/newsee",
		},
		{
			name:     "ok, static with path params",
			whenPath: "/users/:id/files",
		},
		{
			name:     "ok, path params",
			whenPath: "/users/:id",
		},
		{
			name:     "ok, any",
			whenPath: "/users/new/*",
		},
		{
			name:     "ok, any root",
			whenPath: "/*",
		},
		{
			name:     "ok, multilevel any",
			whenPath: "/users/dew/*",
		},
		{
			name:       "ok, single root",
			givenPaths: []string{"/users"},
			whenPath:   "/users",
		},
		{
			name:       "nok, no routes, nothing to remove",
			givenPaths: []string{},
			whenPath:   "/users/",
			expectErr:  "router has no routes to remove",
		},
		{
			name:      "nok, route not found (matches partial node)",
			whenPath:  "/users/",
			expectErr: "could not find route to remove by given path",
		},
		{
			name:      "nok, route not found (matches partial node)",
			whenPath:  "",
			expectErr: "could not find route to remove by given path",
		},
		{
			name:      "nok, route not found",
			whenPath:  "/this_is_not_existent",
			expectErr: "could not find route to remove by given path",
		},
		{
			name:       "nok, multilevel static but different method",
			whenPath:   "/users/newsee",
			whenMethod: http.MethodPost,
			expectErr:  "could not find route to remove by given path and method",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			paths := []string{
				"/users",
				"/users/new",
				"/users/:id",
				"/users/dew",
				"/users/dew/*",
				"/users/:id/files",
				"/users/newsee",
				"/users/new/new",
				"/users/new/*",
				"/*",
			}
			if tc.givenPaths != nil {
				paths = tc.givenPaths
			}
			index := -1
			for i, p := range paths {
				e.GET(p, handlerFunc)
				if p == tc.whenPath {
					index = i
				}
			}
			var toCheckPaths []string
			if index != -1 {
				toCheckPaths = append(paths[:index], paths[index+1:]...)
			}

			method := tc.whenMethod
			if method == "" {
				method = http.MethodGet
			}

			err := e.Router().Remove(method, tc.whenPath)

			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}

			for _, p := range toCheckPaths {
				req := httptest.NewRequest(http.MethodGet, p, nil)
				c := e.NewContext(req, nil)
				_ = e.Router().Route(c)
				assert.Equal(t, p, c.Path(), "after removing %v we matched wrong route. when matching: %v, got: %v", tc.whenPath, p, c.Path())
			}
		})
	}
}

func TestDefaultRouter_AddWithoutHandler(t *testing.T) {
	router := NewRouter(RouterConfig{})

	ri, err := router.Add(Route{Method: http.MethodGet, Path: "/info", Handler: nil})
	assert.EqualError(t, err, "GET /info: adding route without handler function")
	assert.Equal(t, RouteInfo{}, ri)
}

func TestDefaultRouter_AddDuplicateRouteNotAllowed(t *testing.T) {
	e := New()
	router := NewRouter(RouterConfig{})
	e.router = router

	ri, err := router.Add(Route{
		Method: http.MethodGet,
		Path:   "/info",
		Handler: func(c *Context) error {
			return c.String(http.StatusTeapot, "OLD")
		},
		Name: "old",
	})
	assert.NoError(t, err)
	assert.NotNil(t, ri)

	ri, err = router.Add(Route{Method: http.MethodGet, Path: "/info", Handler: handlerFunc, Name: "new"})
	assert.Equal(t, RouteInfo{}, ri)
	assert.EqualError(t, err, "GET /info: adding duplicate route (same method+path) is not allowed")
	assert.Len(t, router.Routes(), 1)

	status, body := request(http.MethodGet, "/info", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "OLD", body)
}

// See issue #1531, #1258 - there are cases when path parameter need to be unescaped
func TestDefaultRouter_UnescapePathParamValues(t *testing.T) {
	var testCases = []struct {
		name                         string
		whenURL                      string
		expectPath                   string
		expectPathValues             PathValues
		givenUnescapePathParamValues bool
	}{
		{
			name:                         "ok, unescape = true",
			givenUnescapePathParamValues: true,
			whenURL:                      "/first/value%20with%20space",
			expectPath:                   "/first/:raw",
			expectPathValues: PathValues{
				{
					Name:  "raw",
					Value: "value with space",
				},
			},
		},
		{
			name:                         "ok, multiple, unescape = true",
			givenUnescapePathParamValues: true,
			whenURL:                      "/second/%20/with%20space",
			expectPath:                   "/second/:id/:fileName",
			expectPathValues: PathValues{
				{Name: "id", Value: " "},
				{Name: "fileName", Value: "with space"},
			},
		},
		{
			name:                         "ok, any route, unescape = true",
			givenUnescapePathParamValues: true,
			whenURL:                      "/third/%20%2Fwith%20space",
			expectPath:                   "/third/*",
			expectPathValues: PathValues{
				{Name: "*", Value: " /with space"},
			},
		},
		{
			name:                         "ok, ending with static node, unescape = true",
			givenUnescapePathParamValues: true,
			whenURL:                      "/fourth/%20%2Fwith%20space/static",
			expectPath:                   "/fourth/:id/static",
			expectPathValues: PathValues{
				{Name: "id", Value: " /with space"},
			},
		},
		{
			name:                         "ok, unescape = false",
			givenUnescapePathParamValues: false,
			whenURL:                      "/first/value%20with%20space",
			expectPath:                   "/first/:raw",
			expectPathValues: PathValues{
				{
					Name:  "raw",
					Value: "value%20with%20space",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			router := NewRouter(RouterConfig{UnescapePathParamValues: tc.givenUnescapePathParamValues})
			e.router = router
			e.contextPathParamAllocSize.Store(2)

			_, err := router.Add(Route{Method: http.MethodGet, Path: "/first/:raw", Handler: handlerFunc})
			assert.NoError(t, err)
			_, err = router.Add(Route{Method: http.MethodGet, Path: "/second/:id/:fileName", Handler: handlerFunc})
			assert.NoError(t, err)
			_, err = router.Add(Route{Method: http.MethodGet, Path: "/third/*", Handler: handlerFunc})
			assert.NoError(t, err)
			_, err = router.Add(Route{Method: http.MethodGet, Path: "/fourth/:id/static", Handler: handlerFunc})
			assert.NoError(t, err)

			target, _ := url.Parse(tc.whenURL)
			req := httptest.NewRequest(http.MethodGet, target.String(), nil)
			req.URL.RawPath = tc.whenURL

			c := e.NewContext(req, nil)
			_ = e.Router().Route(c)

			assert.Equal(t, tc.expectPath, c.Path())
			assert.Equal(t, tc.expectPathValues, c.PathValues())
		})
	}
}

func TestDefaultRouter_AddDuplicateRouteAllowed(t *testing.T) {
	e := New()
	router := NewRouter(RouterConfig{AllowOverwritingRoute: true})
	e.router = router

	ri, err := router.Add(Route{Method: http.MethodGet, Path: "/info", Handler: handlerFunc, Name: "old"})
	assert.NoError(t, err)
	assert.Equal(t, ri, RouteInfo{
		Method: http.MethodGet,
		Path:   "/info",
		Name:   "old",
	})

	ri, err = router.Add(Route{
		Method: http.MethodGet,
		Path:   "/info",
		Handler: func(c *Context) error {
			return c.String(http.StatusTeapot, "NEW")
		},
		Name: "new",
	})
	assert.NoError(t, err)
	assert.Equal(t, ri, RouteInfo{
		Method: http.MethodGet,
		Path:   "/info",
		Name:   "new",
	})

	routes := router.Routes()
	assert.Len(t, routes, 1)
	assert.Equal(t, "new", routes[0].Name)

	status, body := request(http.MethodGet, "/info", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "NEW", body)
}

func TestDefaultRouter_UseEscapedPathForRouting(t *testing.T) {
	var testCases = []struct {
		name                               string
		whenRawPath                        string
		expectBody                         string
		expectStatus                       int
		givenDoNotUseEscapedPathForRouting bool
	}{
		{
			name:         "ok, static route",
			whenRawPath:  "/what's%20up",
			expectBody:   "/what's up",
			expectStatus: http.StatusTeapot,
		},
		{
			name:                               "nok, rawPath does not match with route path",
			givenDoNotUseEscapedPathForRouting: true,
			whenRawPath:                        "/what's%20up", // route is "/what's up"
			expectBody:                         "{\"message\":\"Not Found\"}\n",
			expectStatus:                       http.StatusNotFound,
		},
		{
			name:         "ok, match path param",
			whenRawPath:  "/test/what's%20up",
			expectBody:   "/test/:param|what's up",
			expectStatus: http.StatusTeapot,
		},
		{
			name:                               "ok, path param is unescaped as it is in rawPath",
			givenDoNotUseEscapedPathForRouting: true,
			whenRawPath:                        "/test/what's%20up",
			expectBody:                         "/test/:param|what's%20up",
			expectStatus:                       http.StatusTeapot,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			router := NewRouter(RouterConfig{UseEscapedPathForMatching: !tc.givenDoNotUseEscapedPathForRouting})
			e.router = router
			e.contextPathParamAllocSize.Store(1)

			ri, err := router.Add(Route{
				Method: http.MethodGet,
				Path:   "/what's up",
				Handler: func(c *Context) error {
					return c.String(http.StatusTeapot, c.RouteInfo().Path)
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, ri)
			ri, err = router.Add(Route{
				Method: http.MethodGet,
				Path:   "/test/:param",
				Handler: func(c *Context) error {
					return c.String(http.StatusTeapot, c.RouteInfo().Path+"|"+c.Param("param"))
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, ri)

			status, body := request(http.MethodGet, tc.whenRawPath, e)
			assert.Equal(t, tc.expectStatus, status)
			assert.Equal(t, tc.expectBody, body)
		})
	}
}

func TestDefaultRouter_NotFoundHandler(t *testing.T) {
	e := New()
	router := NewRouter(RouterConfig{
		NotFoundHandler: func(c *Context) error {
			return c.String(http.StatusTeapot, "404")
		},
	})
	e.router = router
	e.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	status, body := request(http.MethodGet, "/noFound", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "404", body)
}

func TestDefaultRouter_MethodNotAllowedHandler(t *testing.T) {
	e := New()
	router := NewRouter(RouterConfig{
		MethodNotAllowedHandler: func(c *Context) error {
			return c.String(http.StatusTeapot, "405")
		},
	})
	e.router = router
	e.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	status, body := request(http.MethodPost, "/test", e)
	assert.Equal(t, http.StatusTeapot, status)
	assert.Equal(t, "405", body)
}

func TestDefaultRouter_OptionsMethodHandler(t *testing.T) {
	e := New()
	router := NewRouter(RouterConfig{
		OptionsMethodHandler: func(c *Context) error {
			return c.String(http.StatusBadRequest, "not empty")
		},
	})
	e.router = router
	e.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	status, body := request(http.MethodOptions, "/test", e)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "not empty", body)
}

func TestRouter_RouteWhenNotFoundRouteWithNodeSplitting(t *testing.T) {
	e := New()
	r := e.router

	hf := func(c *Context) error {
		return c.String(http.StatusOK, c.RouteInfo().Name)
	}
	r.Add(Route{Method: http.MethodGet, Path: "/test*", Handler: hf, Name: "0"})
	r.Add(Route{Method: RouteNotFound, Path: "/test*", Handler: hf, Name: "1"})
	r.Add(Route{Method: RouteNotFound, Path: "/test", Handler: hf, Name: "2"})

	// Tree before:
	// 1    `/`
	// 1.1      `*` (any) ID=1
	// 1.2      `test` (static) ID=2
	// 1.2.1        `*` (any) ID=0

	// node with path `test` has routeNotFound handler from previous Add call. Now when we insert `/te/st*` into router tree
	// This means that node `test` is split into `te` and `st` nodes and new node `/st*` is inserted.
	// On that split `/test` routeNotFound handler must not be lost.
	r.Add(Route{Method: http.MethodGet, Path: "/te/st*", Handler: hf, Name: "3"})
	// Tree after:
	// 1    `/`
	// 1.1      `*` (any) ID=1
	// 1.2      `te` (static)
	// 1.2.1        `st` (static) ID=2
	// 1.2.1.1          `*` (any) ID=0
	// 1.2.2        `/st` (static)
	// 1.2.2.1          `*` (any) ID=3

	_, body := request(http.MethodPut, "/test", e)

	assert.Equal(t, "2", body)
}

func TestRouter_RouteWhenNotFoundRouteAnyKind(t *testing.T) {
	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		name        string
		whenURL     string
		expectID    int
	}{
		{
			name:        "route not existent /xx to not found handler /*",
			whenURL:     "/xx",
			expectRoute: "/*",
			expectID:    4,
			expectParam: map[string]string{"*": "xx"},
		},
		{
			name:        "route not existent /a/xx to not found handler /a/*",
			whenURL:     "/a/xx",
			expectRoute: "/a/*",
			expectID:    5,
			expectParam: map[string]string{"*": "xx"},
		},
		{
			name:        "route not existent /a/c/dxxx to not found handler /a/c/d*",
			whenURL:     "/a/c/dxxx",
			expectRoute: "/a/c/d*",
			expectID:    6,
			expectParam: map[string]string{"*": "xxx"},
		},
		{
			name:        "route /a/c/df to /a/c/df",
			whenURL:     "/a/c/df",
			expectRoute: "/a/c/df",
			expectID:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.contextPathParamAllocSize.Store(1)
			r := e.router

			r.Add(Route{Method: http.MethodGet, Path: "/", Handler: handlerHelper("ID", 0), Name: "0"})
			r.Add(Route{Method: http.MethodGet, Path: "/a/c/df", Handler: handlerHelper("ID", 1), Name: "1"})
			r.Add(Route{Method: http.MethodGet, Path: "/a/b*", Handler: handlerHelper("ID", 2), Name: "2"})
			r.Add(Route{Method: http.MethodPut, Path: "/*", Handler: handlerHelper("ID", 3), Name: "3"})

			r.Add(Route{Method: RouteNotFound, Path: "/a/c/d*", Handler: handlerHelper("ID", 6), Name: "6"})
			r.Add(Route{Method: RouteNotFound, Path: "/a/*", Handler: handlerHelper("ID", 5), Name: "5"})
			r.Add(Route{Method: RouteNotFound, Path: "/*", Handler: handlerHelper("ID", 4), Name: "4"})

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)

			handler := r.Route(c)
			handler(c)

			testValue, _ := c.Get("ID").(int)
			assert.Equal(t, tc.expectID, testValue)
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouter_RouteWhenNotFoundRouteParamKind(t *testing.T) {
	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		name        string
		whenURL     string
		expectID    int
	}{
		{
			name:        "route not existent /xx to not found handler /:file",
			whenURL:     "/xx",
			expectRoute: "/:file",
			expectID:    4,
			expectParam: map[string]string{"file": "xx"},
		},
		{
			name:        "route not existent /a/xx to not found handler /a/:file",
			whenURL:     "/a/xx",
			expectRoute: "/a/:file",
			expectID:    5,
			expectParam: map[string]string{"file": "xx"},
		},
		{
			name:        "route not existent /a/c/dxxx to not found handler /a/c/d:file",
			whenURL:     "/a/c/dxxx",
			expectRoute: "/a/c/d:file",
			expectID:    6,
			expectParam: map[string]string{"file": "xxx"},
		},
		{
			name:        "route /a/c/df to /a/c/df",
			whenURL:     "/a/c/df",
			expectRoute: "/a/c/df",
			expectID:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.contextPathParamAllocSize.Store(1)
			r := e.router

			r.Add(Route{Method: http.MethodGet, Path: "/", Handler: handlerHelper("ID", 0), Name: "0"})
			r.Add(Route{Method: http.MethodGet, Path: "/a/c/df", Handler: handlerHelper("ID", 1), Name: "1"})
			r.Add(Route{Method: http.MethodGet, Path: "/a/b*", Handler: handlerHelper("ID", 2), Name: "2"})
			r.Add(Route{Method: http.MethodPut, Path: "/*", Handler: handlerHelper("ID", 3), Name: "3"})

			r.Add(Route{Method: RouteNotFound, Path: "/a/c/d:file", Handler: handlerHelper("ID", 6), Name: "6"})
			r.Add(Route{Method: RouteNotFound, Path: "/a/:file", Handler: handlerHelper("ID", 5), Name: "5"})
			r.Add(Route{Method: RouteNotFound, Path: "/:file", Handler: handlerHelper("ID", 4), Name: "4"})

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)

			handler := r.Route(c)
			handler(c)

			testValue, _ := c.Get("ID").(int)
			assert.Equal(t, tc.expectID, testValue)
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouter_RouteWhenNotFoundRouteStaticKind(t *testing.T) {
	// note: static not found handler is quite silly thing to have but we still support it
	var testCases = []struct {
		expectRoute any
		expectParam map[string]string
		name        string
		whenURL     string
		expectID    int
	}{
		{
			name:        "route not existent / to not found handler /",
			whenURL:     "/",
			expectRoute: "/",
			expectID:    3,
			expectParam: map[string]string{},
		},
		{
			name:        "route /a to /a",
			whenURL:     "/a",
			expectRoute: "/a",
			expectID:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.contextPathParamAllocSize.Store(1)
			r := e.router

			r.Add(Route{Method: http.MethodPut, Path: "/", Handler: handlerHelper("ID", 0), Name: "0"})
			r.Add(Route{Method: http.MethodGet, Path: "/a", Handler: handlerHelper("ID", 1), Name: "1"})
			r.Add(Route{Method: http.MethodPut, Path: "/*", Handler: handlerHelper("ID", 2), Name: "2"})

			r.Add(Route{Method: RouteNotFound, Path: "/", Handler: handlerHelper("ID", 3), Name: "3"})

			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)
			c := e.NewContext(req, nil)

			handler := r.Route(c)
			handler(c)

			testValue, _ := c.Get("ID").(int)
			assert.Equal(t, tc.expectID, testValue)
			assert.Equal(t, tc.expectRoute, c.Path())
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestPathValues_Get(t *testing.T) {
	var testCases = []struct {
		name     string
		whenKey  string
		expect   string
		expectOK bool
	}{
		{
			name:     "ok",
			whenKey:  `tag`,
			expect:   `latest`,
			expectOK: true,
		},
		{
			name:     "nok",
			whenKey:  `not existent key`,
			expect:   "",
			expectOK: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pv := PathValues{
				{Name: "image", Value: "docker"},
				{Name: "tag", Value: "latest"},
			}

			result, ok := pv.Get(tc.whenKey)
			assert.Equal(t, tc.expectOK, ok)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestPathValues_GetOr(t *testing.T) {
	var testCases = []struct {
		name             string
		whenKey          string
		whenDefaultValue string
		expect           string
	}{
		{
			name:             "ok, existing key",
			whenKey:          `tag`,
			whenDefaultValue: `tag`,
			expect:           `latest`,
		},
		{
			name:             "nok, no existing key return default value",
			whenKey:          `not existent key`,
			whenDefaultValue: `ok`,
			expect:           "ok",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pv := PathValues{
				{Name: "image", Value: "docker"},
				{Name: "tag", Value: "latest"},
			}

			result := pv.GetOr(tc.whenKey, tc.whenDefaultValue)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func BenchmarkRouterStaticRoutes(b *testing.B) {
	benchmarkRouterRoutes(b, staticRoutes, staticRoutes)
}

func BenchmarkRouterStaticRoutesMisses(b *testing.B) {
	benchmarkRouterRoutes(b, staticRoutes, missesAPI)
}

func BenchmarkRouterGitHubAPI(b *testing.B) {
	benchmarkRouterRoutes(b, gitHubAPI, gitHubAPI)
}

func BenchmarkRouterGitHubAPIMisses(b *testing.B) {
	benchmarkRouterRoutes(b, gitHubAPI, missesAPI)
}

func BenchmarkRouterParseAPI(b *testing.B) {
	benchmarkRouterRoutes(b, parseAPI, parseAPI)
}

func BenchmarkRouterParseAPIMisses(b *testing.B) {
	benchmarkRouterRoutes(b, parseAPI, missesAPI)
}

func BenchmarkRouterGooglePlusAPI(b *testing.B) {
	benchmarkRouterRoutes(b, googlePlusAPI, googlePlusAPI)
}

func BenchmarkRouterGooglePlusAPIMisses(b *testing.B) {
	benchmarkRouterRoutes(b, googlePlusAPI, missesAPI)
}

func BenchmarkRouterParamsAndAnyAPI(b *testing.B) {
	benchmarkRouterRoutes(b, paramAndAnyAPI, paramAndAnyAPIToFind)
}

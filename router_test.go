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
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/cmd.html"},
		{Method: "GET", Path: "/code.html"},
		{Method: "GET", Path: "/contrib.html"},
		{Method: "GET", Path: "/contribute.html"},
		{Method: "GET", Path: "/debugging_with_gdb.html"},
		{Method: "GET", Path: "/docs.html"},
		{Method: "GET", Path: "/effective_go.html"},
		{Method: "GET", Path: "/files.log"},
		{Method: "GET", Path: "/gccgo_contribute.html"},
		{Method: "GET", Path: "/gccgo_install.html"},
		{Method: "GET", Path: "/go-logo-black.png"},
		{Method: "GET", Path: "/go-logo-blue.png"},
		{Method: "GET", Path: "/go-logo-white.png"},
		{Method: "GET", Path: "/go1.1.html"},
		{Method: "GET", Path: "/go1.2.html"},
		{Method: "GET", Path: "/go1.html"},
		{Method: "GET", Path: "/go1compat.html"},
		{Method: "GET", Path: "/go_faq.html"},
		{Method: "GET", Path: "/go_mem.html"},
		{Method: "GET", Path: "/go_spec.html"},
		{Method: "GET", Path: "/help.html"},
		{Method: "GET", Path: "/ie.css"},
		{Method: "GET", Path: "/install-source.html"},
		{Method: "GET", Path: "/install.html"},
		{Method: "GET", Path: "/logo-153x55.png"},
		{Method: "GET", Path: "/Makefile"},
		{Method: "GET", Path: "/root.html"},
		{Method: "GET", Path: "/share.png"},
		{Method: "GET", Path: "/sieve.gif"},
		{Method: "GET", Path: "/tos.html"},
		{Method: "GET", Path: "/articles/"},
		{Method: "GET", Path: "/articles/go_command.html"},
		{Method: "GET", Path: "/articles/index.html"},
		{Method: "GET", Path: "/articles/wiki/"},
		{Method: "GET", Path: "/articles/wiki/edit.html"},
		{Method: "GET", Path: "/articles/wiki/final-noclosure.go"},
		{Method: "GET", Path: "/articles/wiki/final-noerror.go"},
		{Method: "GET", Path: "/articles/wiki/final-parsetemplate.go"},
		{Method: "GET", Path: "/articles/wiki/final-template.go"},
		{Method: "GET", Path: "/articles/wiki/final.go"},
		{Method: "GET", Path: "/articles/wiki/get.go"},
		{Method: "GET", Path: "/articles/wiki/http-sample.go"},
		{Method: "GET", Path: "/articles/wiki/index.html"},
		{Method: "GET", Path: "/articles/wiki/Makefile"},
		{Method: "GET", Path: "/articles/wiki/notemplate.go"},
		{Method: "GET", Path: "/articles/wiki/part1-noerror.go"},
		{Method: "GET", Path: "/articles/wiki/part1.go"},
		{Method: "GET", Path: "/articles/wiki/part2.go"},
		{Method: "GET", Path: "/articles/wiki/part3-errorhandling.go"},
		{Method: "GET", Path: "/articles/wiki/part3.go"},
		{Method: "GET", Path: "/articles/wiki/test.bash"},
		{Method: "GET", Path: "/articles/wiki/test_edit.good"},
		{Method: "GET", Path: "/articles/wiki/test_Test.txt.good"},
		{Method: "GET", Path: "/articles/wiki/test_view.good"},
		{Method: "GET", Path: "/articles/wiki/view.html"},
		{Method: "GET", Path: "/codewalk/"},
		{Method: "GET", Path: "/codewalk/codewalk.css"},
		{Method: "GET", Path: "/codewalk/codewalk.js"},
		{Method: "GET", Path: "/codewalk/codewalk.xml"},
		{Method: "GET", Path: "/codewalk/functions.xml"},
		{Method: "GET", Path: "/codewalk/markov.go"},
		{Method: "GET", Path: "/codewalk/markov.xml"},
		{Method: "GET", Path: "/codewalk/pig.go"},
		{Method: "GET", Path: "/codewalk/popout.png"},
		{Method: "GET", Path: "/codewalk/run"},
		{Method: "GET", Path: "/codewalk/sharemem.xml"},
		{Method: "GET", Path: "/codewalk/urlpoll.go"},
		{Method: "GET", Path: "/devel/"},
		{Method: "GET", Path: "/devel/release.html"},
		{Method: "GET", Path: "/devel/weekly.html"},
		{Method: "GET", Path: "/gopher/"},
		{Method: "GET", Path: "/gopher/appenginegopher.jpg"},
		{Method: "GET", Path: "/gopher/appenginegophercolor.jpg"},
		{Method: "GET", Path: "/gopher/appenginelogo.gif"},
		{Method: "GET", Path: "/gopher/bumper.png"},
		{Method: "GET", Path: "/gopher/bumper192x108.png"},
		{Method: "GET", Path: "/gopher/bumper320x180.png"},
		{Method: "GET", Path: "/gopher/bumper480x270.png"},
		{Method: "GET", Path: "/gopher/bumper640x360.png"},
		{Method: "GET", Path: "/gopher/doc.png"},
		{Method: "GET", Path: "/gopher/frontpage.png"},
		{Method: "GET", Path: "/gopher/gopherbw.png"},
		{Method: "GET", Path: "/gopher/gophercolor.png"},
		{Method: "GET", Path: "/gopher/gophercolor16x16.png"},
		{Method: "GET", Path: "/gopher/help.png"},
		{Method: "GET", Path: "/gopher/pkg.png"},
		{Method: "GET", Path: "/gopher/project.png"},
		{Method: "GET", Path: "/gopher/ref.png"},
		{Method: "GET", Path: "/gopher/run.png"},
		{Method: "GET", Path: "/gopher/talks.png"},
		{Method: "GET", Path: "/gopher/pencil/"},
		{Method: "GET", Path: "/gopher/pencil/gopherhat.jpg"},
		{Method: "GET", Path: "/gopher/pencil/gopherhelmet.jpg"},
		{Method: "GET", Path: "/gopher/pencil/gophermega.jpg"},
		{Method: "GET", Path: "/gopher/pencil/gopherrunning.jpg"},
		{Method: "GET", Path: "/gopher/pencil/gopherswim.jpg"},
		{Method: "GET", Path: "/gopher/pencil/gopherswrench.jpg"},
		{Method: "GET", Path: "/play/"},
		{Method: "GET", Path: "/play/fib.go"},
		{Method: "GET", Path: "/play/hello.go"},
		{Method: "GET", Path: "/play/life.go"},
		{Method: "GET", Path: "/play/peano.go"},
		{Method: "GET", Path: "/play/pi.go"},
		{Method: "GET", Path: "/play/sieve.go"},
		{Method: "GET", Path: "/play/solitaire.go"},
		{Method: "GET", Path: "/play/tree.go"},
		{Method: "GET", Path: "/progs/"},
		{Method: "GET", Path: "/progs/cgo1.go"},
		{Method: "GET", Path: "/progs/cgo2.go"},
		{Method: "GET", Path: "/progs/cgo3.go"},
		{Method: "GET", Path: "/progs/cgo4.go"},
		{Method: "GET", Path: "/progs/defer.go"},
		{Method: "GET", Path: "/progs/defer.out"},
		{Method: "GET", Path: "/progs/defer2.go"},
		{Method: "GET", Path: "/progs/defer2.out"},
		{Method: "GET", Path: "/progs/eff_bytesize.go"},
		{Method: "GET", Path: "/progs/eff_bytesize.out"},
		{Method: "GET", Path: "/progs/eff_qr.go"},
		{Method: "GET", Path: "/progs/eff_sequence.go"},
		{Method: "GET", Path: "/progs/eff_sequence.out"},
		{Method: "GET", Path: "/progs/eff_unused1.go"},
		{Method: "GET", Path: "/progs/eff_unused2.go"},
		{Method: "GET", Path: "/progs/error.go"},
		{Method: "GET", Path: "/progs/error2.go"},
		{Method: "GET", Path: "/progs/error3.go"},
		{Method: "GET", Path: "/progs/error4.go"},
		{Method: "GET", Path: "/progs/go1.go"},
		{Method: "GET", Path: "/progs/gobs1.go"},
		{Method: "GET", Path: "/progs/gobs2.go"},
		{Method: "GET", Path: "/progs/image_draw.go"},
		{Method: "GET", Path: "/progs/image_package1.go"},
		{Method: "GET", Path: "/progs/image_package1.out"},
		{Method: "GET", Path: "/progs/image_package2.go"},
		{Method: "GET", Path: "/progs/image_package2.out"},
		{Method: "GET", Path: "/progs/image_package3.go"},
		{Method: "GET", Path: "/progs/image_package3.out"},
		{Method: "GET", Path: "/progs/image_package4.go"},
		{Method: "GET", Path: "/progs/image_package4.out"},
		{Method: "GET", Path: "/progs/image_package5.go"},
		{Method: "GET", Path: "/progs/image_package5.out"},
		{Method: "GET", Path: "/progs/image_package6.go"},
		{Method: "GET", Path: "/progs/image_package6.out"},
		{Method: "GET", Path: "/progs/interface.go"},
		{Method: "GET", Path: "/progs/interface2.go"},
		{Method: "GET", Path: "/progs/interface2.out"},
		{Method: "GET", Path: "/progs/json1.go"},
		{Method: "GET", Path: "/progs/json2.go"},
		{Method: "GET", Path: "/progs/json2.out"},
		{Method: "GET", Path: "/progs/json3.go"},
		{Method: "GET", Path: "/progs/json4.go"},
		{Method: "GET", Path: "/progs/json5.go"},
		{Method: "GET", Path: "/progs/run"},
		{Method: "GET", Path: "/progs/slices.go"},
		{Method: "GET", Path: "/progs/timeout1.go"},
		{Method: "GET", Path: "/progs/timeout2.go"},
		{Method: "GET", Path: "/progs/update.bash"},
	}

	gitHubAPI = []*Route{
		// OAuth Authorizations
		{Method: "GET", Path: "/authorizations"},
		{Method: "GET", Path: "/authorizations/:id"},
		{Method: "POST", Path: "/authorizations"},

		{Method: "PUT", Path: "/authorizations/clients/:client_id"},
		{Method: "PATCH", Path: "/authorizations/:id"},

		{Method: "DELETE", Path: "/authorizations/:id"},
		{Method: "GET", Path: "/applications/:client_id/tokens/:access_token"},
		{Method: "DELETE", Path: "/applications/:client_id/tokens"},
		{Method: "DELETE", Path: "/applications/:client_id/tokens/:access_token"},

		// Activity
		{Method: "GET", Path: "/events"},
		{Method: "GET", Path: "/repos/:owner/:repo/events"},
		{Method: "GET", Path: "/networks/:owner/:repo/events"},
		{Method: "GET", Path: "/orgs/:org/events"},
		{Method: "GET", Path: "/users/:user/received_events"},
		{Method: "GET", Path: "/users/:user/received_events/public"},
		{Method: "GET", Path: "/users/:user/events"},
		{Method: "GET", Path: "/users/:user/events/public"},
		{Method: "GET", Path: "/users/:user/events/orgs/:org"},
		{Method: "GET", Path: "/feeds"},
		{Method: "GET", Path: "/notifications"},
		{Method: "GET", Path: "/repos/:owner/:repo/notifications"},
		{Method: "PUT", Path: "/notifications"},
		{Method: "PUT", Path: "/repos/:owner/:repo/notifications"},
		{Method: "GET", Path: "/notifications/threads/:id"},

		{Method: "PATCH", Path: "/notifications/threads/:id"},

		{Method: "GET", Path: "/notifications/threads/:id/subscription"},
		{Method: "PUT", Path: "/notifications/threads/:id/subscription"},
		{Method: "DELETE", Path: "/notifications/threads/:id/subscription"},
		{Method: "GET", Path: "/repos/:owner/:repo/stargazers"},
		{Method: "GET", Path: "/users/:user/starred"},
		{Method: "GET", Path: "/user/starred"},
		{Method: "GET", Path: "/user/starred/:owner/:repo"},
		{Method: "PUT", Path: "/user/starred/:owner/:repo"},
		{Method: "DELETE", Path: "/user/starred/:owner/:repo"},
		{Method: "GET", Path: "/repos/:owner/:repo/subscribers"},
		{Method: "GET", Path: "/users/:user/subscriptions"},
		{Method: "GET", Path: "/user/subscriptions"},
		{Method: "GET", Path: "/repos/:owner/:repo/subscription"},
		{Method: "PUT", Path: "/repos/:owner/:repo/subscription"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/subscription"},
		{Method: "GET", Path: "/user/subscriptions/:owner/:repo"},
		{Method: "PUT", Path: "/user/subscriptions/:owner/:repo"},
		{Method: "DELETE", Path: "/user/subscriptions/:owner/:repo"},

		// Gists
		{Method: "GET", Path: "/users/:user/gists"},
		{Method: "GET", Path: "/gists"},

		{Method: "GET", Path: "/gists/public"},
		{Method: "GET", Path: "/gists/starred"},

		{Method: "GET", Path: "/gists/:id"},
		{Method: "POST", Path: "/gists"},

		{Method: "PATCH", Path: "/gists/:id"},

		{Method: "PUT", Path: "/gists/:id/star"},
		{Method: "DELETE", Path: "/gists/:id/star"},
		{Method: "GET", Path: "/gists/:id/star"},
		{Method: "POST", Path: "/gists/:id/forks"},
		{Method: "DELETE", Path: "/gists/:id"},

		// Git Data
		{Method: "GET", Path: "/repos/:owner/:repo/git/blobs/:sha"},
		{Method: "POST", Path: "/repos/:owner/:repo/git/blobs"},
		{Method: "GET", Path: "/repos/:owner/:repo/git/commits/:sha"},
		{Method: "POST", Path: "/repos/:owner/:repo/git/commits"},

		{Method: "GET", Path: "/repos/:owner/:repo/git/refs/*ref"},

		{Method: "GET", Path: "/repos/:owner/:repo/git/refs"},
		{Method: "POST", Path: "/repos/:owner/:repo/git/refs"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/git/refs/*ref"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/git/refs/*ref"},

		{Method: "GET", Path: "/repos/:owner/:repo/git/tags/:sha"},
		{Method: "POST", Path: "/repos/:owner/:repo/git/tags"},
		{Method: "GET", Path: "/repos/:owner/:repo/git/trees/:sha"},
		{Method: "POST", Path: "/repos/:owner/:repo/git/trees"},

		// Issues
		{Method: "GET", Path: "/issues"},
		{Method: "GET", Path: "/user/issues"},
		{Method: "GET", Path: "/orgs/:org/issues"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues/:number"},
		{Method: "POST", Path: "/repos/:owner/:repo/issues"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/issues/:number"},

		{Method: "GET", Path: "/repos/:owner/:repo/assignees"},
		{Method: "GET", Path: "/repos/:owner/:repo/assignees/:assignee"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues/:number/comments"},

		{Method: "GET", Path: "/repos/:owner/:repo/issues/comments"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues/comments/:id"},

		{Method: "POST", Path: "/repos/:owner/:repo/issues/:number/comments"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/issues/comments/:id"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/issues/comments/:id"},

		{Method: "GET", Path: "/repos/:owner/:repo/issues/:number/events"},

		{Method: "GET", Path: "/repos/:owner/:repo/issues/events"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues/events/:id"},

		{Method: "GET", Path: "/repos/:owner/:repo/labels"},
		{Method: "GET", Path: "/repos/:owner/:repo/labels/:name"},
		{Method: "POST", Path: "/repos/:owner/:repo/labels"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/labels/:name"},

		{Method: "DELETE", Path: "/repos/:owner/:repo/labels/:name"},
		{Method: "GET", Path: "/repos/:owner/:repo/issues/:number/labels"},
		{Method: "POST", Path: "/repos/:owner/:repo/issues/:number/labels"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/issues/:number/labels/:name"},
		{Method: "PUT", Path: "/repos/:owner/:repo/issues/:number/labels"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/issues/:number/labels"},
		{Method: "GET", Path: "/repos/:owner/:repo/milestones/:number/labels"},
		{Method: "GET", Path: "/repos/:owner/:repo/milestones"},
		{Method: "GET", Path: "/repos/:owner/:repo/milestones/:number"},
		{Method: "POST", Path: "/repos/:owner/:repo/milestones"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/milestones/:number"},

		{Method: "DELETE", Path: "/repos/:owner/:repo/milestones/:number"},

		// Miscellaneous
		{Method: "GET", Path: "/emojis"},
		{Method: "GET", Path: "/gitignore/templates"},
		{Method: "GET", Path: "/gitignore/templates/:name"},
		{Method: "POST", Path: "/markdown"},
		{Method: "POST", Path: "/markdown/raw"},
		{Method: "GET", Path: "/meta"},
		{Method: "GET", Path: "/rate_limit"},

		// Organizations
		{Method: "GET", Path: "/users/:user/orgs"},
		{Method: "GET", Path: "/user/orgs"},
		{Method: "GET", Path: "/orgs/:org"},

		{Method: "PATCH", Path: "/orgs/:org"},

		{Method: "GET", Path: "/orgs/:org/members"},
		{Method: "GET", Path: "/orgs/:org/members/:user"},
		{Method: "DELETE", Path: "/orgs/:org/members/:user"},
		{Method: "GET", Path: "/orgs/:org/public_members"},
		{Method: "GET", Path: "/orgs/:org/public_members/:user"},
		{Method: "PUT", Path: "/orgs/:org/public_members/:user"},
		{Method: "DELETE", Path: "/orgs/:org/public_members/:user"},
		{Method: "GET", Path: "/orgs/:org/teams"},
		{Method: "GET", Path: "/teams/:id"},
		{Method: "POST", Path: "/orgs/:org/teams"},

		{Method: "PATCH", Path: "/teams/:id"},

		{Method: "DELETE", Path: "/teams/:id"},
		{Method: "GET", Path: "/teams/:id/members"},
		{Method: "GET", Path: "/teams/:id/members/:user"},
		{Method: "PUT", Path: "/teams/:id/members/:user"},
		{Method: "DELETE", Path: "/teams/:id/members/:user"},
		{Method: "GET", Path: "/teams/:id/repos"},
		{Method: "GET", Path: "/teams/:id/repos/:owner/:repo"},
		{Method: "PUT", Path: "/teams/:id/repos/:owner/:repo"},
		{Method: "DELETE", Path: "/teams/:id/repos/:owner/:repo"},
		{Method: "GET", Path: "/user/teams"},

		// Pull Requests
		{Method: "GET", Path: "/repos/:owner/:repo/pulls"},
		{Method: "GET", Path: "/repos/:owner/:repo/pulls/:number"},
		{Method: "POST", Path: "/repos/:owner/:repo/pulls"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/pulls/:number"},

		{Method: "GET", Path: "/repos/:owner/:repo/pulls/:number/commits"},
		{Method: "GET", Path: "/repos/:owner/:repo/pulls/:number/files"},
		{Method: "GET", Path: "/repos/:owner/:repo/pulls/:number/merge"},
		{Method: "PUT", Path: "/repos/:owner/:repo/pulls/:number/merge"},
		{Method: "GET", Path: "/repos/:owner/:repo/pulls/:number/comments"},

		{Method: "GET", Path: "/repos/:owner/:repo/pulls/comments"},
		{Method: "GET", Path: "/repos/:owner/:repo/pulls/comments/:number"},

		{Method: "PUT", Path: "/repos/:owner/:repo/pulls/:number/comments"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/pulls/comments/:number"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/pulls/comments/:number"},

		// Repositories
		{Method: "GET", Path: "/user/repos"},
		{Method: "GET", Path: "/users/:user/repos"},
		{Method: "GET", Path: "/orgs/:org/repos"},
		{Method: "GET", Path: "/repositories"},
		{Method: "POST", Path: "/user/repos"},
		{Method: "POST", Path: "/orgs/:org/repos"},
		{Method: "GET", Path: "/repos/:owner/:repo"},

		{Method: "PATCH", Path: "/repos/:owner/:repo"},

		{Method: "GET", Path: "/repos/:owner/:repo/contributors"},
		{Method: "GET", Path: "/repos/:owner/:repo/languages"},
		{Method: "GET", Path: "/repos/:owner/:repo/teams"},
		{Method: "GET", Path: "/repos/:owner/:repo/tags"},
		{Method: "GET", Path: "/repos/:owner/:repo/branches"},
		{Method: "GET", Path: "/repos/:owner/:repo/branches/:branch"},
		{Method: "DELETE", Path: "/repos/:owner/:repo"},
		{Method: "GET", Path: "/repos/:owner/:repo/collaborators"},
		{Method: "GET", Path: "/repos/:owner/:repo/collaborators/:user"},
		{Method: "PUT", Path: "/repos/:owner/:repo/collaborators/:user"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/collaborators/:user"},
		{Method: "GET", Path: "/repos/:owner/:repo/comments"},
		{Method: "GET", Path: "/repos/:owner/:repo/commits/:sha/comments"},
		{Method: "POST", Path: "/repos/:owner/:repo/commits/:sha/comments"},
		{Method: "GET", Path: "/repos/:owner/:repo/comments/:id"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/comments/:id"},

		{Method: "DELETE", Path: "/repos/:owner/:repo/comments/:id"},
		{Method: "GET", Path: "/repos/:owner/:repo/commits"},
		{Method: "GET", Path: "/repos/:owner/:repo/commits/:sha"},
		{Method: "GET", Path: "/repos/:owner/:repo/readme"},

		//{Method: "GET", Path: "/repos/:owner/:repo/contents/*path"},
		//{Method: "PUT", Path: "/repos/:owner/:repo/contents/*path"},
		//{Method: "DELETE", Path: "/repos/:owner/:repo/contents/*path"},

		{Method: "GET", Path: "/repos/:owner/:repo/:archive_format/:ref"},

		{Method: "GET", Path: "/repos/:owner/:repo/keys"},
		{Method: "GET", Path: "/repos/:owner/:repo/keys/:id"},
		{Method: "POST", Path: "/repos/:owner/:repo/keys"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/keys/:id"},

		{Method: "DELETE", Path: "/repos/:owner/:repo/keys/:id"},
		{Method: "GET", Path: "/repos/:owner/:repo/downloads"},
		{Method: "GET", Path: "/repos/:owner/:repo/downloads/:id"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/downloads/:id"},
		{Method: "GET", Path: "/repos/:owner/:repo/forks"},
		{Method: "POST", Path: "/repos/:owner/:repo/forks"},
		{Method: "GET", Path: "/repos/:owner/:repo/hooks"},
		{Method: "GET", Path: "/repos/:owner/:repo/hooks/:id"},
		{Method: "POST", Path: "/repos/:owner/:repo/hooks"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/hooks/:id"},

		{Method: "POST", Path: "/repos/:owner/:repo/hooks/:id/tests"},
		{Method: "DELETE", Path: "/repos/:owner/:repo/hooks/:id"},
		{Method: "POST", Path: "/repos/:owner/:repo/merges"},
		{Method: "GET", Path: "/repos/:owner/:repo/releases"},
		{Method: "GET", Path: "/repos/:owner/:repo/releases/:id"},
		{Method: "POST", Path: "/repos/:owner/:repo/releases"},

		{Method: "PATCH", Path: "/repos/:owner/:repo/releases/:id"},

		{Method: "DELETE", Path: "/repos/:owner/:repo/releases/:id"},
		{Method: "GET", Path: "/repos/:owner/:repo/releases/:id/assets"},
		{Method: "GET", Path: "/repos/:owner/:repo/stats/contributors"},
		{Method: "GET", Path: "/repos/:owner/:repo/stats/commit_activity"},
		{Method: "GET", Path: "/repos/:owner/:repo/stats/code_frequency"},
		{Method: "GET", Path: "/repos/:owner/:repo/stats/participation"},
		{Method: "GET", Path: "/repos/:owner/:repo/stats/punch_card"},
		{Method: "GET", Path: "/repos/:owner/:repo/statuses/:ref"},
		{Method: "POST", Path: "/repos/:owner/:repo/statuses/:ref"},

		// Search
		{Method: "GET", Path: "/search/repositories"},
		{Method: "GET", Path: "/search/code"},
		{Method: "GET", Path: "/search/issues"},
		{Method: "GET", Path: "/search/users"},
		{Method: "GET", Path: "/legacy/issues/search/:owner/:repository/:state/:keyword"},
		{Method: "GET", Path: "/legacy/repos/search/:keyword"},
		{Method: "GET", Path: "/legacy/user/search/:keyword"},
		{Method: "GET", Path: "/legacy/user/email/:email"},

		// Users
		{Method: "GET", Path: "/users/:user"},
		{Method: "GET", Path: "/user"},

		{Method: "PATCH", Path: "/user"},

		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/user/emails"},
		{Method: "POST", Path: "/user/emails"},
		{Method: "DELETE", Path: "/user/emails"},
		{Method: "GET", Path: "/users/:user/followers"},
		{Method: "GET", Path: "/user/followers"},
		{Method: "GET", Path: "/users/:user/following"},
		{Method: "GET", Path: "/user/following"},
		{Method: "GET", Path: "/user/following/:user"},
		{Method: "GET", Path: "/users/:user/following/:target_user"},
		{Method: "PUT", Path: "/user/following/:user"},
		{Method: "DELETE", Path: "/user/following/:user"},
		{Method: "GET", Path: "/users/:user/keys"},
		{Method: "GET", Path: "/user/keys"},
		{Method: "GET", Path: "/user/keys/:id"},
		{Method: "POST", Path: "/user/keys"},

		{Method: "PATCH", Path: "/user/keys/:id"},

		{Method: "DELETE", Path: "/user/keys/:id"},
	}

	parseAPI = []*Route{
		// Objects
		{Method: "POST", Path: "/1/classes/:className"},
		{Method: "GET", Path: "/1/classes/:className/:objectId"},
		{Method: "PUT", Path: "/1/classes/:className/:objectId"},
		{Method: "GET", Path: "/1/classes/:className"},
		{Method: "DELETE", Path: "/1/classes/:className/:objectId"},

		// Users
		{Method: "POST", Path: "/1/users"},
		{Method: "GET", Path: "/1/login"},
		{Method: "GET", Path: "/1/users/:objectId"},
		{Method: "PUT", Path: "/1/users/:objectId"},
		{Method: "GET", Path: "/1/users"},
		{Method: "DELETE", Path: "/1/users/:objectId"},
		{Method: "POST", Path: "/1/requestPasswordReset"},

		// Roles
		{Method: "POST", Path: "/1/roles"},
		{Method: "GET", Path: "/1/roles/:objectId"},
		{Method: "PUT", Path: "/1/roles/:objectId"},
		{Method: "GET", Path: "/1/roles"},
		{Method: "DELETE", Path: "/1/roles/:objectId"},

		// Files
		{Method: "POST", Path: "/1/files/:fileName"},

		// Analytics
		{Method: "POST", Path: "/1/events/:eventName"},

		// Push Notifications
		{Method: "POST", Path: "/1/push"},

		// Installations
		{Method: "POST", Path: "/1/installations"},
		{Method: "GET", Path: "/1/installations/:objectId"},
		{Method: "PUT", Path: "/1/installations/:objectId"},
		{Method: "GET", Path: "/1/installations"},
		{Method: "DELETE", Path: "/1/installations/:objectId"},

		// Cloud Functions
		{Method: "POST", Path: "/1/functions"},
	}

	googlePlusAPI = []*Route{
		// People
		{Method: "GET", Path: "/people/:userId"},
		{Method: "GET", Path: "/people"},
		{Method: "GET", Path: "/activities/:activityId/people/:collection"},
		{Method: "GET", Path: "/people/:userId/people/:collection"},
		{Method: "GET", Path: "/people/:userId/openIdConnect"},

		// Activities
		{Method: "GET", Path: "/people/:userId/activities/:collection"},
		{Method: "GET", Path: "/activities/:activityId"},
		{Method: "GET", Path: "/activities"},

		// Comments
		{Method: "GET", Path: "/activities/:activityId/comments"},
		{Method: "GET", Path: "/comments/:commentId"},

		// Moments
		{Method: "POST", Path: "/people/:userId/moments/:collection"},
		{Method: "GET", Path: "/people/:userId/moments/:collection"},
		{Method: "DELETE", Path: "/moments/:id"},
	}

	paramAndAnyAPI = []*Route{
		{Method: "GET", Path: "/root/:first/foo/*"},
		{Method: "GET", Path: "/root/:first/:second/*"},
		{Method: "GET", Path: "/root/:first/bar/:second/*"},
		{Method: "GET", Path: "/root/:first/qux/:second/:third/:fourth"},
		{Method: "GET", Path: "/root/:first/qux/:second/:third/:fourth/*"},
		{Method: "GET", Path: "/root/*"},

		{Method: "POST", Path: "/root/:first/foo/*"},
		{Method: "POST", Path: "/root/:first/:second/*"},
		{Method: "POST", Path: "/root/:first/bar/:second/*"},
		{Method: "POST", Path: "/root/:first/qux/:second/:third/:fourth"},
		{Method: "POST", Path: "/root/:first/qux/:second/:third/:fourth/*"},
		{Method: "POST", Path: "/root/*"},

		{Method: "PUT", Path: "/root/:first/foo/*"},
		{Method: "PUT", Path: "/root/:first/:second/*"},
		{Method: "PUT", Path: "/root/:first/bar/:second/*"},
		{Method: "PUT", Path: "/root/:first/qux/:second/:third/:fourth"},
		{Method: "PUT", Path: "/root/:first/qux/:second/:third/:fourth/*"},
		{Method: "PUT", Path: "/root/*"},

		{Method: "DELETE", Path: "/root/:first/foo/*"},
		{Method: "DELETE", Path: "/root/:first/:second/*"},
		{Method: "DELETE", Path: "/root/:first/bar/:second/*"},
		{Method: "DELETE", Path: "/root/:first/qux/:second/:third/:fourth"},
		{Method: "DELETE", Path: "/root/:first/qux/:second/:third/:fourth/*"},
		{Method: "DELETE", Path: "/root/*"},
	}

	paramAndAnyAPIToFind = []*Route{
		{Method: "GET", Path: "/root/one/foo/after/the/asterisk"},
		{Method: "GET", Path: "/root/one/foo/path/after/the/asterisk"},
		{Method: "GET", Path: "/root/one/two/path/after/the/asterisk"},
		{Method: "GET", Path: "/root/one/bar/two/after/the/asterisk"},
		{Method: "GET", Path: "/root/one/qux/two/three/four"},
		{Method: "GET", Path: "/root/one/qux/two/three/four/after/the/asterisk"},

		{Method: "POST", Path: "/root/one/foo/after/the/asterisk"},
		{Method: "POST", Path: "/root/one/foo/path/after/the/asterisk"},
		{Method: "POST", Path: "/root/one/two/path/after/the/asterisk"},
		{Method: "POST", Path: "/root/one/bar/two/after/the/asterisk"},
		{Method: "POST", Path: "/root/one/qux/two/three/four"},
		{Method: "POST", Path: "/root/one/qux/two/three/four/after/the/asterisk"},

		{Method: "PUT", Path: "/root/one/foo/after/the/asterisk"},
		{Method: "PUT", Path: "/root/one/foo/path/after/the/asterisk"},
		{Method: "PUT", Path: "/root/one/two/path/after/the/asterisk"},
		{Method: "PUT", Path: "/root/one/bar/two/after/the/asterisk"},
		{Method: "PUT", Path: "/root/one/qux/two/three/four"},
		{Method: "PUT", Path: "/root/one/qux/two/three/four/after/the/asterisk"},

		{Method: "DELETE", Path: "/root/one/foo/after/the/asterisk"},
		{Method: "DELETE", Path: "/root/one/foo/path/after/the/asterisk"},
		{Method: "DELETE", Path: "/root/one/two/path/after/the/asterisk"},
		{Method: "DELETE", Path: "/root/one/bar/two/after/the/asterisk"},
		{Method: "DELETE", Path: "/root/one/qux/two/three/four"},
		{Method: "DELETE", Path: "/root/one/qux/two/three/four/after/the/asterisk"},
	}

	missesAPI = []*Route{
		{Method: "GET", Path: "/missOne"},
		{Method: "GET", Path: "/miss/two"},
		{Method: "GET", Path: "/miss/three/levels"},
		{Method: "GET", Path: "/miss/four/levels/nooo"},

		{Method: "POST", Path: "/missOne"},
		{Method: "POST", Path: "/miss/two"},
		{Method: "POST", Path: "/miss/three/levels"},
		{Method: "POST", Path: "/miss/four/levels/nooo"},

		{Method: "PUT", Path: "/missOne"},
		{Method: "PUT", Path: "/miss/two"},
		{Method: "PUT", Path: "/miss/three/levels"},
		{Method: "PUT", Path: "/miss/four/levels/nooo"},

		{Method: "DELETE", Path: "/missOne"},
		{Method: "DELETE", Path: "/miss/two"},
		{Method: "DELETE", Path: "/miss/three/levels"},
		{Method: "DELETE", Path: "/miss/four/levels/nooo"},
	}

	// handlerHelper created a function that will set a context key for assertion
	handlerHelper = func(key string, value int) func(c Context) error {
		return func(c Context) error {
			c.Set(key, value)
			c.Set("path", c.Path())
			return nil
		}
	}
	handlerFunc = func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}
)

func checkUnusedParamValues(t *testing.T, c *context, expectParam map[string]string) {
	for i, p := range c.pnames {
		value := c.pvalues[i]
		if value != "" {
			if expectParam == nil {
				t.Errorf("pValue '%v' is set for param name '%v' but we are not expecting it with expectParam", value, p)
			} else {
				if _, ok := expectParam[p]; !ok {
					t.Errorf("pValue '%v' is set for param name '%v' but we are not expecting it with expectParam", value, p)
				}
			}
		}
	}
}

func TestRouterStatic(t *testing.T) {
	e := New()
	r := e.router
	path := "/folders/a/files/echo.gif"
	r.Add(http.MethodGet, path, handlerFunc)
	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, path, c)
	c.handler(c)

	assert.Equal(t, path, c.Get("path"))
}

func TestRouterParam(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/users/:id", handlerFunc)

	var testCases = []struct {
		name        string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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

			c := e.NewContext(nil, nil).(*context)
			r.Find(http.MethodGet, tc.whenURL, c)

			c.handler(c)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestMethodNotAllowedAndNotFound(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/*", handlerFunc)
	r.Add(http.MethodPost, "/users/:id", handlerFunc)

	var testCases = []struct {
		name        string
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
	}{
		{
			name:        "exact match for route+method",
			whenMethod:  http.MethodPost,
			whenURL:     "/users/1",
			expectRoute: "/users/:id",
			expectParam: map[string]string{"id": "1"},
		},
		{
			name:        "matches node but not method. sends 405 from best match node",
			whenMethod:  http.MethodPut,
			whenURL:     "/users/1",
			expectRoute: nil,
			expectError: ErrMethodNotAllowed,
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
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add(http.MethodGet, "/users/:uid/files/:fid", handlerFunc)
	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, "/users/1/files/1", c)

	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "1", c.Param("fid"))
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/a/:b/c/d/:e", handlerFunc)
	r.Add(http.MethodGet, "/a/:b/c/:d/:f", handlerFunc)

	c := e.NewContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/a/1/c/d/2/3", c) // `2/3` should mapped to path `/a/:b/c/d/:e` and into `:e`

	err := c.handler(c)
	assert.Equal(t, "/a/:b/c/d/:e", c.Get("path"))
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
//                             +----------+
//                       +-----+ "/" root +--------------------+--------------------------+
//                       |     +----------+                    |                          |
//                       |                                     |                          |
//               +-------v-------+                         +---v---------+        +-------v---+
//               | "a/" (static) +---------------+         | ":" (param) |        | "*" (any) |
//               +-+----------+--+               |         +-----------+-+        +-----------+
//                 |          |                  |                     |
// +---------------v+  +-- ---v------+    +------v----+          +-----v-----------+
// | "c/d" (static) |  | ":" (param) |    | "*" (any) |          | "/c/f" (static) |
// +---------+------+  +--------+----+    +----------++          +-----------------+
//           |                  |                    |
//           |                  |                    |
// +---------v----+      +------v--------+    +------v--------+
// | "f" (static) |      | "/c" (static) |    | "/f" (static) |
// +--------------+      +---------------+    +---------------+
func TestRouteMultiLevelBacktracking(t *testing.T) {
	var testCases = []struct {
		name        string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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
			r := e.router

			r.Add(http.MethodGet, "/a/:b/c", handlerHelper("case", 1))
			r.Add(http.MethodGet, "/a/c/d", handlerHelper("case", 2))
			r.Add(http.MethodGet, "/a/c/df", handlerHelper("case", 3))
			r.Add(http.MethodGet, "/a/*/f", handlerHelper("case", 4))
			r.Add(http.MethodGet, "/:e/c/f", handlerHelper("case", 5))
			r.Add(http.MethodGet, "/*", handlerHelper("case", 6))

			c := e.NewContext(nil, nil).(*context)
			r.Find(http.MethodGet, tc.whenURL, c)

			c.handler(c)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
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
//                            +-0,7--------+
//                            | "/" (root) |----------------------------------+
//                            +------------+                                  |
//                                 |      |                                   |
//                                 |      |                                   |
//             +-1,6-----------+   |      |          +-8-----------+   +------v----+
//             | "a/" (static) +<--+      +--------->+ ":" (param) |   | "*" (any) |
//             +---------------+                     +-------------+   +-----------+
//                |          |                             |
//     +-2--------v-----+   +v-3,5--------+       +-9------v--------+
//     | "c/d" (static) |   | ":" (param) |       | "/c/f" (static) |
//     +----------------+   +-------------+       +-----------------+
//                           |
//                      +-4--v----------+
//                      | "/c" (static) |
//                      +---------------+
func TestRouteMultiLevelBacktracking2(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/a/:b/c", handlerFunc)
	r.Add(http.MethodGet, "/a/c/d", handlerFunc)
	r.Add(http.MethodGet, "/a/c/df", handlerFunc)
	r.Add(http.MethodGet, "/:e/c/f", handlerFunc)
	r.Add(http.MethodGet, "/*", handlerFunc)

	var testCases = []struct {
		name        string
		whenURL     string
		expectRoute string
		expectParam map[string]string
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)

			c.handler(c)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterBacktrackingFromMultipleParamKinds(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/*", handlerFunc) // this can match only path that does not have slash in it
	r.Add(http.MethodGet, "/:1/second", handlerFunc)
	r.Add(http.MethodGet, "/:1/:2", handlerFunc) // this acts as match ANY for all routes that have at least one slash
	r.Add(http.MethodGet, "/:1/:2/third", handlerFunc)
	r.Add(http.MethodGet, "/:1/:2/:3/fourth", handlerFunc)
	r.Add(http.MethodGet, "/:1/:2/:3/:4/fifth", handlerFunc)

	c := e.NewContext(nil, nil).(*context)
	var testCases = []struct {
		name        string
		whenURL     string
		expectRoute string
		expectParam map[string]string
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
			whenURL:     "/first/second/",
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
			r.Find(http.MethodGet, tc.whenURL, c)

			c.handler(c)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1509
func TestRouterParamStaticConflict(t *testing.T) {
	e := New()
	r := e.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	g := e.Group("/g")
	g.GET("/skills", handler)
	g.GET("/status", handler)
	g.GET("/:name", handler)

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/", handlerFunc)
	r.Add(http.MethodGet, "/*", handlerFunc)
	r.Add(http.MethodGet, "/users/*", handlerFunc)

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
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
	r := e.router

	r.Add(http.MethodGet, "/users/*", handlerHelper("case", 1))
	r.Add(http.MethodGet, "/users/*/action*", handlerHelper("case", 2))

	c := e.NewContext(nil, nil).(*context)

	r.Find(http.MethodGet, "/users/xxx/action/sea", c)
	c.handler(c)
	assert.Equal(t, "/users/*/action*", c.Get("path"))
	assert.Equal(t, "xxx/action/sea", c.Param("*"))

	// if we add another route then it is the last added and so it is matched
	r.Add(http.MethodGet, "/users/*/action/search", handlerHelper("case", 3))

	r.Find(http.MethodGet, "/users/xxx/action/sea", c)
	c.handler(c)
	assert.Equal(t, "/users/*/action/search", c.Get("path"))
	assert.Equal(t, "xxx/action/sea", c.Param("*"))
}

// Issue #1739
func TestRouterMatchAnyPrefixIssue(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/*", func(c Context) error {
		c.Set("path", c.Path())
		return nil
	})
	r.Add(http.MethodGet, "/users/*", func(c Context) error {
		c.Set("path", c.Path())
		return nil
	})

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// TestRouterMatchAnySlash shall verify finding the best route
// for any routes with trailing slash requests
func TestRouterMatchAnySlash(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", handlerFunc)
	r.Add(http.MethodGet, "/users/*", handlerFunc)
	r.Add(http.MethodGet, "/img/*", handlerFunc)
	r.Add(http.MethodGet, "/img/load", handlerFunc)
	r.Add(http.MethodGet, "/img/load/*", handlerFunc)
	r.Add(http.MethodGet, "/assets/*", handlerFunc)

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
	}{
		{
			whenURL:     "/",
			expectRoute: nil,
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
			expectRoute: nil,
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/api/users/jack", handlerFunc)
	r.Add(http.MethodGet, "/api/users/jill", handlerFunc)
	r.Add(http.MethodGet, "/api/users/*", handlerFunc)
	r.Add(http.MethodGet, "/api/*", handlerFunc)
	r.Add(http.MethodGet, "/other/*", handlerFunc)
	r.Add(http.MethodGet, "/*", handlerFunc)

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}
func TestRouterMatchAnyMultiLevelWithPost(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	e.POST("/api/auth/login", handlerFunc)
	e.POST("/api/auth/forgotPassword", handlerFunc)
	e.Any("/api/*", handlerFunc)
	e.Any("/*", handlerFunc)

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
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
	r.Add(http.MethodGet, "/users", handlerFunc)
	r.Add(http.MethodGet, "/users/:id", handlerFunc)

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			expectRoute: nil,
			expectParam: map[string]string{"*": ""},
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterPriority(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", handlerFunc)
	r.Add(http.MethodGet, "/users/new", handlerFunc)
	r.Add(http.MethodGet, "/users/:id", handlerFunc)
	r.Add(http.MethodGet, "/users/dew", handlerFunc)
	r.Add(http.MethodGet, "/users/:id/files", handlerFunc)
	r.Add(http.MethodGet, "/users/newsee", handlerFunc)
	r.Add(http.MethodGet, "/users/*", handlerFunc)
	r.Add(http.MethodGet, "/users/new/*", handlerFunc)
	r.Add(http.MethodGet, "/*", handlerFunc)

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
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

	// Add
	r.Add(http.MethodGet, "/a/foo", handlerFunc)
	r.Add(http.MethodGet, "/a/bar", handlerFunc)

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			expectRoute: nil,
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func TestRouterParamNames(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add(http.MethodGet, "/users", handlerFunc)
	r.Add(http.MethodGet, "/users/:id", handlerFunc)
	r.Add(http.MethodGet, "/users/:uid/files/:fid", handlerFunc)

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #623 and #1406
func TestRouterStaticDynamicConflict(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/dictionary/skills", handlerHelper("a", 1))
	r.Add(http.MethodGet, "/dictionary/:name", handlerHelper("b", 2))
	r.Add(http.MethodGet, "/users/new", handlerHelper("d", 4))
	r.Add(http.MethodGet, "/users/:name", handlerHelper("e", 5))
	r.Add(http.MethodGet, "/server", handlerHelper("c", 3))
	r.Add(http.MethodGet, "/", handlerHelper("f", 6))

	var testCases = []struct {
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1348
func TestRouterParamBacktraceNotFound(t *testing.T) {
	e := New()
	r := e.router

	// Add
	r.Add(http.MethodGet, "/:param1", handlerFunc)
	r.Add(http.MethodGet, "/:param1/foo", handlerFunc)
	r.Add(http.MethodGet, "/:param1/bar", handlerFunc)
	r.Add(http.MethodGet, "/:param1/bar/:param2", handlerFunc)

	var testCases = []struct {
		name        string
		whenMethod  string
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
		expectError error
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
			expectRoute: nil,
			expectError: ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := e.NewContext(nil, nil).(*context)

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			r.Find(method, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectError != nil {
				assert.Equal(t, tc.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
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
		t.Run(route.Path, func(t *testing.T) {
			r.Find(route.Method, route.Path, c)
			tokens := strings.Split(route.Path[1:], "/")
			for _, token := range tokens {
				if token[0] == ':' {
					assert.Equal(t, c.Param(token[1:]), token)
				}
			}
		})
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

// Issue #729
func TestRouterParamAlias(t *testing.T) {
	api := []*Route{
		{Method: http.MethodGet, Path: "/users/:userID/following"},
		{Method: http.MethodGet, Path: "/users/:userID/followedBy"},
		{Method: http.MethodGet, Path: "/users/:userID/follow"},
	}
	testRouterAPI(t, api)
}

// Issue #1052
func TestRouterParamOrdering(t *testing.T) {
	api := []*Route{
		{Method: http.MethodGet, Path: "/:a/:b/:c/:id"},
		{Method: http.MethodGet, Path: "/:a/:id"},
		{Method: http.MethodGet, Path: "/:a/:e/:id"},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{Method: http.MethodGet, Path: "/:a/:id"},
		{Method: http.MethodGet, Path: "/:a/:e/:id"},
		{Method: http.MethodGet, Path: "/:a/:b/:c/:id"},
	}
	testRouterAPI(t, api2)
	api3 := []*Route{
		{Method: http.MethodGet, Path: "/:a/:b/:c/:id"},
		{Method: http.MethodGet, Path: "/:a/:e/:id"},
		{Method: http.MethodGet, Path: "/:a/:id"},
	}
	testRouterAPI(t, api3)
}

// Issue #1139
func TestRouterMixedParams(t *testing.T) {
	api := []*Route{
		{Method: http.MethodGet, Path: "/teacher/:tid/room/suggestions"},
		{Method: http.MethodGet, Path: "/teacher/:id"},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{Method: http.MethodGet, Path: "/teacher/:id"},
		{Method: http.MethodGet, Path: "/teacher/:tid/room/suggestions"},
	}
	testRouterAPI(t, api2)
}

// Issue #1466
func TestRouterParam1466(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodPost, "/users/signup", handlerFunc)
	r.Add(http.MethodPost, "/users/signup/bulk", handlerFunc)
	r.Add(http.MethodPost, "/users/survey", handlerFunc)
	r.Add(http.MethodGet, "/users/:username", handlerFunc)
	r.Add(http.MethodGet, "/interests/:name/users", handlerFunc)
	r.Add(http.MethodGet, "/skills/:name/users", handlerFunc)
	// Additional routes for Issue 1479
	r.Add(http.MethodGet, "/users/:username/likes/projects/ids", handlerFunc)
	r.Add(http.MethodGet, "/users/:username/profile", handlerFunc)
	r.Add(http.MethodGet, "/users/:username/uploads/:type", handlerFunc)

	var testCases = []struct {
		whenURL     string
		expectRoute interface{}
		expectParam map[string]string
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
			expectRoute: nil, // not found
			expectParam: map[string]string{"id": ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			c.handler(c)
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

// Issue #1655
func TestRouterFindNotPanicOrLoopsWhenContextSetParamValuesIsCalledWithLessValuesThanEchoMaxParam(t *testing.T) {
	e := New()
	r := e.router

	v0 := e.Group("/:version")
	v0.GET("/admin", func(c Context) error {
		c.SetParamNames("version")
		c.SetParamValues("v1")
		return nil
	})

	v0.GET("/images/view/:id", handlerHelper("iv", 1))
	v0.GET("/images/:id", handlerHelper("i", 1))
	v0.GET("/view/*", handlerHelper("v", 1))

	//If this API is called before the next two one panic the other loops ( of course without my fix ;) )
	c := e.NewContext(nil, nil)
	r.Find(http.MethodGet, "/v1/admin", c)
	c.Handler()(c)
	assert.Equal(t, "v1", c.Param("version"))

	//panic
	c = e.NewContext(nil, nil)
	r.Find(http.MethodGet, "/v1/view/same-data", c)
	c.Handler()(c)
	assert.Equal(t, "same-data", c.Param("*"))
	assert.Equal(t, 1, c.Get("v"))

	//looping
	c = e.NewContext(nil, nil)
	r.Find(http.MethodGet, "/v1/images/view", c)
	c.Handler()(c)
	assert.Equal(t, "view", c.Param("id"))
	assert.Equal(t, 1, c.Get("i"))
}

// Issue #1653
func TestRouterPanicWhenParamNoRootOnlyChildsFailsFind(t *testing.T) {
	e := New()
	r := e.router

	r.Add(http.MethodGet, "/users/create", handlerFunc)
	r.Add(http.MethodGet, "/users/:id/edit", handlerFunc)
	r.Add(http.MethodGet, "/users/:id/active", handlerFunc)

	var testCases = []struct {
		whenURL      string
		expectRoute  interface{}
		expectParam  map[string]string
		expectStatus int
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
			whenURL:      "/users/createNotFound",
			expectStatus: http.StatusNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			c := e.NewContext(nil, nil).(*context)

			r.Find(http.MethodGet, tc.whenURL, c)
			err := c.handler(c)

			if tc.expectStatus != 0 {
				assert.Error(t, err)
				he := err.(*HTTPError)
				assert.Equal(t, tc.expectStatus, he.Code)
			}
			assert.Equal(t, tc.expectRoute, c.Get("path"))
			for param, expectedValue := range tc.expectParam {
				assert.Equal(t, expectedValue, c.Param(param))
			}
			checkUnusedParamValues(t, c, tc.expectParam)
		})
	}
}

func benchmarkRouterRoutes(b *testing.B, routes []*Route, routesToFind []*Route) {
	e := New()
	r := e.router
	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		r.Add(route.Method, route.Path, func(c Context) error {
			return nil
		})
	}

	// Routes adding are performed just once, so it doesn't make sense to see that in the benchmark
	b.ResetTimer()

	// Find routes
	for i := 0; i < b.N; i++ {
		for _, route := range routesToFind {
			c := e.pool.Get().(*context)
			r.Find(route.Method, route.Path, c)
			e.pool.Put(c)
		}
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

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s, %p: type=%d, parent=%p, handler=%v, pnames=%v\n", p, n.prefix, n, n.kind, n.parent, n.methodHandler, n.pnames)

	p = prefix(tail, pfx, "    ", "│   ")

	children := n.staticChildren
	l := len(children)

	if n.paramChild != nil {
		n.paramChild.printTree(p, n.anyChild == nil && l == 0)
	}
	if n.anyChild != nil {
		n.anyChild.printTree(p, l == 0)
	}
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

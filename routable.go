package echo

type (
	// Routeable is an interface which allows dependency injection in place of using concrete types Echo, and Group while creating routes.
	Routeable interface {
		Use(middleware ...MiddlewareFunc)
		CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
		Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
		Group(prefix string, middleware ...MiddlewareFunc) (sg *Group)
		Static(prefix, root string)
		File(path, file string)
		Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route
	}
)

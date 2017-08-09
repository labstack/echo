package echo

type (
	// Register is the interface for Echo and Group.
	Register interface {
		StaticRegister
		DynamicRegister
		MethodsRegister

		Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
		Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
	}

	//MethodsRegister is the interface providing methods to register routes for Echo and Group.
	MethodsRegister interface {
		CONNECT(string, HandlerFunc, ...MiddlewareFunc) *Route
		GET(string, HandlerFunc, ...MiddlewareFunc) *Route
		HEAD(string, HandlerFunc, ...MiddlewareFunc) *Route
		OPTIONS(string, HandlerFunc, ...MiddlewareFunc) *Route
		PATCH(string, HandlerFunc, ...MiddlewareFunc) *Route
		POST(string, HandlerFunc, ...MiddlewareFunc) *Route
		PUT(string, HandlerFunc, ...MiddlewareFunc) *Route
		TRACE(string, HandlerFunc, ...MiddlewareFunc) *Route
		DELETE(string, HandlerFunc, ...MiddlewareFunc) *Route
	}

	//DynamicRegister is the interface providing methods to register dynamic routes for Echo and Group.
	DynamicRegister interface {
		Add(string, string, HandlerFunc, ...MiddlewareFunc) *Route
	}

	//StaticRegister is the interface providing methods to register static files routes for Echo and Group.
	StaticRegister interface {
		Static(prefix, root string) *Route
		File(path, file string) *Route
	}
)

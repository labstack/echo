package echo

// Mountpoint defines an interface that captures the common routing methods
// between Echo and Group, allowing an App to be more flexible in its routing.
type Mountpoint interface {
	// Use adds middleware to the router
	Use(middleware ...MiddlewareFunc)

	// HTTP routing methods
	CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
	TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// RouteNotFound sets the handler to be called when a route is not found
	RouteNotFound(path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// Any handles all HTTP methods
	Any(path string, h HandlerFunc, m ...MiddlewareFunc) []*Route

	// Match handles multiple HTTP methods
	Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route

	// Add registers a new route with a matcher for the URL path and method
	Add(method, path string, h HandlerFunc, m ...MiddlewareFunc) *Route

	// Group creates a new router group with prefix and optional middleware
	Group(prefix string, m ...MiddlewareFunc) *Group
}

// Ensure that both Echo and Group implement the Mountpoint interface
var (
	_ Mountpoint = (*Echo)(nil)
	_ Mountpoint = (*Group)(nil)
)

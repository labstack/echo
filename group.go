package echo

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middlware or functionality that should be separate
	// from the parent echo instance while still inheriting from it.
	Group struct {
		prefix     string
		middleware []MiddlewareFunc
		echo       *Echo
	}
)

// Use implements `Echo#Use()` for sub-routes within the Group.
func (g *Group) Use(m ...MiddlewareFunc) {
	g.middleware = append(g.middleware, m...)
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
	g.echo.Any(g.prefix+"*", func(c Context) error {
		return ErrNotFound
	}, g.middleware...)
}

// CONNECT implements `Echo#CONNECT()` for sub-routes within the Group.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(CONNECT, path, h, m...)
}

// Connect is deprecated, use `CONNECT()` instead.
func (g *Group) Connect(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(CONNECT, path, h, m...)
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group.
func (g *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(DELETE, path, h, m...)
}

// Delete is deprecated, use `DELETE()` instead.
func (g *Group) Delete(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(DELETE, path, h, m...)
}

// GET implements `Echo#GET()` for sub-routes within the Group.
func (g *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(GET, path, h, m...)
}

// Get is deprecated, use `GET()` instead.
func (g *Group) Get(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(GET, path, h, m...)
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group.
func (g *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(HEAD, path, h, m...)
}

// Head is deprecated, use `HEAD()` instead.
func (g *Group) Head(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(HEAD, path, h, m...)
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(OPTIONS, path, h, m...)
}

// Options is deprecated, use `OPTIONS()` instead.
func (g *Group) Options(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(OPTIONS, path, h, m...)
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group.
func (g *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(PATCH, path, h, m...)
}

// Patch is deprecated, use `PATCH()` instead.
func (g *Group) Patch(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(PATCH, path, h, m...)
}

// POST implements `Echo#POST()` for sub-routes within the Group.
func (g *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(POST, path, h, m...)
}

// Post is deprecated, use `POST()` instead.
func (g *Group) Post(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(POST, path, h, m...)
}

// PUT implements `Echo#PUT()` for sub-routes within the Group.
func (g *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(PUT, path, h, m...)
}

// Put is deprecated, use `PUT()` instead.
func (g *Group) Put(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(PUT, path, h, m...)
}

// TRACE implements `Echo#TRACE()` for sub-routes within the Group.
func (g *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(TRACE, path, h, m...)
}

// Trace is deprecated, use `TRACE()` instead.
func (g *Group) Trace(path string, h HandlerFunc, m ...MiddlewareFunc) {
	g.add(TRACE, path, h, m...)
}

// Any implements `Echo#Any()` for sub-routes within the Group.
func (g *Group) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

// Match implements `Echo#Match()` for sub-routes within the Group.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	m := []MiddlewareFunc{}
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.echo.Group(g.prefix+prefix, m...)
}

// Static implements `Echo#Static()` for sub-routes within the Group.
func (g *Group) Static(prefix, root string) {
	g.echo.Static(g.prefix+prefix, root)
}

// File implements `Echo#File()` for sub-routes within the Group.
func (g *Group) File(path, file string) {
	g.echo.File(g.prefix+path, file)
}

func (g *Group) add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	// Combine into a new slice, to avoid accidentally passing the same
	// slice for multiple routes, which would lead to later add() calls overwriting
	// the middleware from earlier calls
	m := []MiddlewareFunc{}
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	g.echo.add(method, g.prefix+path, handler, m...)
}

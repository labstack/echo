package echo

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middlware or functionality that should be separate
	// from the parent echo instance while still inheriting from it.
	Group struct {
		prefix     string
		middleware []Middleware
		echo       *Echo
	}
)

// Use implements `Echo#Use()` for sub-routes within the Group.
func (g *Group) Use(m ...Middleware) {
	g.middleware = append(g.middleware, m...)
}

// Connect implements `Echo#Connect()` for sub-routes within the Group.
func (g *Group) Connect(path string, h Handler, m ...Middleware) {
	g.add(CONNECT, path, h, m...)
}

// Delete implements `Echo#Delete()` for sub-routes within the Group.
func (g *Group) Delete(path string, h Handler, m ...Middleware) {
	g.add(DELETE, path, h, m...)
}

// Get implements `Echo#Get()` for sub-routes within the Group.
func (g *Group) Get(path string, h Handler, m ...Middleware) {
	g.add(GET, path, h, m...)
}

// Head implements `Echo#Head()` for sub-routes within the Group.
func (g *Group) Head(path string, h Handler, m ...Middleware) {
	g.add(HEAD, path, h, m...)
}

// Options implements `Echo#Options()` for sub-routes within the Group.
func (g *Group) Options(path string, h Handler, m ...Middleware) {
	g.add(OPTIONS, path, h, m...)
}

// Patch implements `Echo#Patch()` for sub-routes within the Group.
func (g *Group) Patch(path string, h Handler, m ...Middleware) {
	g.add(PATCH, path, h, m...)
}

// Post implements `Echo#Post()` for sub-routes within the Group.
func (g *Group) Post(path string, h Handler, m ...Middleware) {
	g.add(POST, path, h, m...)
}

// Put implements `Echo#Put()` for sub-routes within the Group.
func (g *Group) Put(path string, h Handler, m ...Middleware) {
	g.add(PUT, path, h, m...)
}

// Trace implements `Echo#Trace()` for sub-routes within the Group.
func (g *Group) Trace(path string, h Handler, m ...Middleware) {
	g.add(TRACE, path, h, m...)
}

// Any implements `Echo#Any()` for sub-routes within the Group.
func (g *Group) Any(path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

// Match implements `Echo#Match()` for sub-routes within the Group.
func (g *Group) Match(methods []string, path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, m ...Middleware) *Group {
	m = append(g.middleware, m...)
	return g.echo.Group(g.prefix+prefix, m...)
}

func (g *Group) add(method, path string, handler Handler, middleware ...Middleware) {
	path = g.prefix + path
	name := handlerName(handler)
	middleware = append(g.middleware, middleware...)

	g.echo.router.Add(method, path, HandlerFunc(func(c Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i].Handle(h)
		}
		return h.Handle(c)
	}), g.echo)
	r := Route{
		Method:  method,
		Path:    path,
		Handler: name,
	}
	g.echo.router.routes = append(g.echo.router.routes, r)
}

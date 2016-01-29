package echo

// Group is a set of subroutes for a specified route. It can be used for inner
// routes that share a common middlware or functionality that should be separate
// from the parent echo instance while still inheriting from it.
type Group struct {
	echo Echo
}

// Use implements the echo.Use interface for subroutes within the Group.
func (g *Group) Use(m ...Middleware) {
	for _, h := range m {
		g.echo.middleware = append(g.echo.middleware, wrapMiddleware(h))
	}
}

// Connect implements the echo.Connect interface for subroutes within the Group.
func (g *Group) Connect(path string, h Handler) {
	g.echo.Connect(path, h)
}

// Delete implements the echo.Delete interface for subroutes within the Group.
func (g *Group) Delete(path string, h Handler) {
	g.echo.Delete(path, h)
}

// Get implements the echo.Get interface for subroutes within the Group.
func (g *Group) Get(path string, h Handler) {
	g.echo.Get(path, h)
}

// Head implements the echo.Head interface for subroutes within the Group.
func (g *Group) Head(path string, h Handler) {
	g.echo.Head(path, h)
}

// Options implements the echo.Options interface for subroutes within the Group.
func (g *Group) Options(path string, h Handler) {
	g.echo.Options(path, h)
}

// Patch implements the echo.Patch interface for subroutes within the Group.
func (g *Group) Patch(path string, h Handler) {
	g.echo.Patch(path, h)
}

// Post implements the echo.Post interface for subroutes within the Group.
func (g *Group) Post(path string, h Handler) {
	g.echo.Post(path, h)
}

// Put implements the echo.Put interface for subroutes within the Group.
func (g *Group) Put(path string, h Handler) {
	g.echo.Put(path, h)
}

// Trace implements the echo.Trace interface for subroutes within the Group.
func (g *Group) Trace(path string, h Handler) {
	g.echo.Trace(path, h)
}

// Any implements the echo.Any interface for subroutes within the Group.
func (g *Group) Any(path string, h Handler) {
	for _, m := range methods {
		g.echo.add(m, path, h)
	}
}

// Match implements the echo.Match interface for subroutes within the Group.
func (g *Group) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		g.echo.add(m, path, h)
	}
}

// WebSocket implements the echo.WebSocket interface for subroutes within the
// Group.
func (g *Group) WebSocket(path string, h HandlerFunc) {
	g.echo.WebSocket(path, h)
}

// Static implements the echo.Static interface for subroutes within the Group.
func (g *Group) Static(path, root string) {
	g.echo.Static(path, root)
}

// ServeDir implements the echo.ServeDir interface for subroutes within the
// Group.
func (g *Group) ServeDir(path, root string) {
	g.echo.ServeDir(path, root)
}

// ServeFile implements the echo.ServeFile interface for subroutes within the
// Group.
func (g *Group) ServeFile(path, file string) {
	g.echo.ServeFile(path, file)
}

// Group implements the echo.Group interface for subroutes within the Group.
func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.echo.Group(prefix, m...)
}

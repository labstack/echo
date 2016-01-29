package echo

type (
	Group struct {
		echo Echo
	}
)

func (g *Group) Use(m ...Middleware) {
	for _, h := range m {
		g.echo.middleware = append(g.echo.middleware, wrapMiddleware(h))
	}
}

func (g *Group) Connect(path string, h Handler) {
	g.echo.Connect(path, h)
}

func (g *Group) Delete(path string, h Handler) {
	g.echo.Delete(path, h)
}

func (g *Group) Get(path string, h Handler) {
	g.echo.Get(path, h)
}

func (g *Group) Head(path string, h Handler) {
	g.echo.Head(path, h)
}

func (g *Group) Options(path string, h Handler) {
	g.echo.Options(path, h)
}

func (g *Group) Patch(path string, h Handler) {
	g.echo.Patch(path, h)
}

func (g *Group) Post(path string, h Handler) {
	g.echo.Post(path, h)
}

func (g *Group) Put(path string, h Handler) {
	g.echo.Put(path, h)
}

func (g *Group) Trace(path string, h Handler) {
	g.echo.Trace(path, h)
}

func (g *Group) Any(path string, h Handler) {
	for _, m := range methods {
		g.echo.add(m, path, h)
	}
}

func (g *Group) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		g.echo.add(m, path, h)
	}
}

func (g *Group) WebSocket(path string, h HandlerFunc) {
	g.echo.WebSocket(path, h)
}

func (g *Group) Static(path, root string) {
	g.echo.Static(path, root)
}

func (g *Group) ServeDir(path, root string) {
	g.echo.ServeDir(path, root)
}

func (g *Group) ServeFile(path, file string) {
	g.echo.ServeFile(path, file)
}

func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.echo.Group(prefix, m...)
}

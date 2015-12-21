package echo

type (
	Group struct {
		echo Echo
	}
)

func (g *Group) Use(m ...MiddlewareFunc) {
	for _, h := range m {
		g.echo.middleware = append(g.echo.middleware, h)
	}
}

func (g *Group) Connect(path string, h HandlerFunc) {
	g.echo.Connect(path, h)
}

func (g *Group) Delete(path string, h HandlerFunc) {
	g.echo.Delete(path, h)
}

func (g *Group) Get(path string, h HandlerFunc) {
	g.echo.Get(path, h)
}

func (g *Group) Head(path string, h HandlerFunc) {
	g.echo.Head(path, h)
}

func (g *Group) Options(path string, h HandlerFunc) {
	g.echo.Options(path, h)
}

func (g *Group) Patch(path string, h HandlerFunc) {
	g.echo.Patch(path, h)
}

func (g *Group) Post(path string, h HandlerFunc) {
	g.echo.Post(path, h)
}

func (g *Group) Put(path string, h HandlerFunc) {
	g.echo.Put(path, h)
}

func (g *Group) Trace(path string, h HandlerFunc) {
	g.echo.Trace(path, h)
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

func (g *Group) Group(prefix string, m ...MiddlewareFunc) *Group {
	return g.echo.Group(prefix, m...)
}

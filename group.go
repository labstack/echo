package echo

type (
	Group struct {
		prefix     string
		middleware []Middleware
		echo       *Echo
	}
)

func (g *Group) Use(m ...Middleware) {
	g.middleware = append(g.middleware, m...)
}

func (g *Group) Connect(path string, h Handler) {
	g.add(CONNECT, path, h)
}

func (g *Group) Delete(path string, h Handler) {
	g.add(DELETE, path, h)
}

func (g *Group) Get(path string, h Handler) {
	g.add(GET, path, h)
}

func (g *Group) Head(path string, h Handler) {
	g.add(HEAD, path, h)
}

func (g *Group) Options(path string, h Handler) {
	g.add(OPTIONS, path, h)
}

func (g *Group) Patch(path string, h Handler) {
	g.add(PATCH, path, h)
}

func (g *Group) Post(path string, h Handler) {
	g.add(POST, path, h)
}

func (g *Group) Put(path string, h Handler) {
	g.add(PUT, path, h)
}

func (g *Group) Trace(path string, h Handler) {
	g.add(TRACE, path, h)
}

func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.echo.Group(prefix, m...)
}

func (g *Group) add(method, path string, h Handler) {
	path = g.prefix + path
	name := handlerName(h)
	g.echo.router.Add(method, path, HandlerFunc(func(c Context) error {
		for i := len(g.middleware) - 1; i >= 0; i-- {
			h = g.middleware[i].Handle(h)
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

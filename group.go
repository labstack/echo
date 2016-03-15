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

func (g *Group) Connect(path string, h Handler, m ...Middleware) {
	g.add(CONNECT, path, h, m...)
}

func (g *Group) Delete(path string, h Handler, m ...Middleware) {
	g.add(DELETE, path, h, m...)
}

func (g *Group) Get(path string, h Handler, m ...Middleware) {
	g.add(GET, path, h, m...)
}

func (g *Group) Head(path string, h Handler, m ...Middleware) {
	g.add(HEAD, path, h, m...)
}

func (g *Group) Options(path string, h Handler, m ...Middleware) {
	g.add(OPTIONS, path, h, m...)
}

func (g *Group) Patch(path string, h Handler, m ...Middleware) {
	g.add(PATCH, path, h, m...)
}

func (g *Group) Post(path string, h Handler, m ...Middleware) {
	g.add(POST, path, h, m...)
}

func (g *Group) Put(path string, h Handler, m ...Middleware) {
	g.add(PUT, path, h, m...)
}

func (g *Group) Trace(path string, h Handler, m ...Middleware) {
	g.add(TRACE, path, h, m...)
}

func (g *Group) Any(path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

func (g *Group) Match(methods []string, path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		g.add(m, path, handler, middleware...)
	}
}

func (g *Group) Group(prefix string, m ...Middleware) *Group {
	return g.echo.Group(g.prefix+prefix, m...)
}

func (g *Group) add(method, path string, handler Handler, middleware ...Middleware) {
	path = g.prefix + path
	name := handlerName(handler)
	middleware = append(g.middleware, middleware...)

	g.echo.router.Add(method, path, HandlerFunc(func(c Context) error {
		h := handler
		for _, m := range middleware {
			h = m.Handle(h)
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

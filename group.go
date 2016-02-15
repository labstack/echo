package echo

type (
	Group struct {
		prefix     string
		middleware []Middleware
		echo       *Echo
	}
)

func (g *Group) Use(middleware ...interface{}) {
	for _, m := range middleware {
		g.middleware = append(g.middleware, wrapMiddleware(m))
	}
}

func (g *Group) Connect(path string, handler interface{}) {
	g.add(CONNECT, path, handler)
}

func (g *Group) Delete(path string, handler interface{}) {
	g.add(DELETE, path, handler)
}

func (g *Group) Get(path string, handler interface{}) {
	g.add(GET, path, handler)
}

func (g *Group) Head(path string, handler interface{}) {
	g.add(HEAD, path, handler)
}

func (g *Group) Options(path string, handler interface{}) {
	g.add(OPTIONS, path, handler)
}

func (g *Group) Patch(path string, handler interface{}) {
	g.add(PATCH, path, handler)
}

func (g *Group) Post(path string, handler interface{}) {
	g.add(POST, path, handler)
}

func (g *Group) Put(path string, handler interface{}) {
	g.add(PUT, path, handler)
}

func (g *Group) Trace(path string, handler interface{}) {
	g.add(TRACE, path, handler)
}

func (g *Group) Group(prefix string, middleware ...interface{}) *Group {
	return g.echo.Group(prefix, middleware...)
}

func (g *Group) add(method, path string, handler interface{}) {
	path = g.prefix + path
	h := wrapHandler(handler)
	name := handlerName(handler)
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

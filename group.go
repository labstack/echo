package echo

import (
	"io/fs"
	"net/http"
)

// Group is a set of sub-routes for a specified route. It can be used for inner
// routes that share a common middleware or functionality that should be separate
// from the parent echo instance while still inheriting from it.
type Group struct {
	host       string
	prefix     string
	middleware []MiddlewareFunc
	echo       *Echo
}

// Use implements `Echo#Use()` for sub-routes within the Group.
// Group middlewares are not executed on request when there is no matching route found.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// CONNECT implements `Echo#CONNECT()` for sub-routes within the Group. Panics on error.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodConnect, path, h, m...)
}

// DELETE implements `Echo#DELETE()` for sub-routes within the Group. Panics on error.
func (g *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodDelete, path, h, m...)
}

// GET implements `Echo#GET()` for sub-routes within the Group. Panics on error.
func (g *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodGet, path, h, m...)
}

// HEAD implements `Echo#HEAD()` for sub-routes within the Group. Panics on error.
func (g *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodHead, path, h, m...)
}

// OPTIONS implements `Echo#OPTIONS()` for sub-routes within the Group. Panics on error.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodOptions, path, h, m...)
}

// PATCH implements `Echo#PATCH()` for sub-routes within the Group. Panics on error.
func (g *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodPatch, path, h, m...)
}

// POST implements `Echo#POST()` for sub-routes within the Group. Panics on error.
func (g *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodPost, path, h, m...)
}

// PUT implements `Echo#PUT()` for sub-routes within the Group. Panics on error.
func (g *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodPut, path, h, m...)
}

// TRACE implements `Echo#TRACE()` for sub-routes within the Group. Panics on error.
func (g *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(http.MethodTrace, path, h, m...)
}

// Any implements `Echo#Any()` for sub-routes within the Group. Panics on error.
func (g *Group) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) Routes {
	errs := make([]error, 0)
	ris := make(Routes, 0)
	for _, m := range methods {
		ri, err := g.AddRoute(Route{
			Method:      m,
			Path:        path,
			Handler:     handler,
			Middlewares: middleware,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ris = append(ris, ri)
	}
	if len(errs) > 0 {
		panic(errs) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ris
}

// Match implements `Echo#Match()` for sub-routes within the Group. Panics on error.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) Routes {
	errs := make([]error, 0)
	ris := make(Routes, 0)
	for _, m := range methods {
		ri, err := g.AddRoute(Route{
			Method:      m,
			Path:        path,
			Handler:     handler,
			Middlewares: middleware,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ris = append(ris, ri)
	}
	if len(errs) > 0 {
		panic(errs) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ris
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
// Important! Group middlewares are only executed in case there was exact route match and not
// for 404 (not found) or 405 (method not allowed) cases. If this kind of behaviour is needed then add
// a catch-all route `/*` for the group which handler returns always 404
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) (sg *Group) {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	sg = g.echo.Group(g.prefix+prefix, m...)
	sg.host = g.host
	return
}

// Static implements `Echo#Static()` for sub-routes within the Group.
func (g *Group) Static(pathPrefix, fsRoot string) RouteInfo {
	subFs := MustSubFS(g.echo.Filesystem, fsRoot)
	return g.StaticFS(pathPrefix, subFs)
}

// StaticFS implements `Echo#StaticFS()` for sub-routes within the Group.
//
// When dealing with `embed.FS` use `fs := echo.MustSubFS(fs, "rootDirectory") to create sub fs which uses necessary
// prefix for directory path. This is necessary as `//go:embed assets/images` embeds files with paths
// including `assets/images` as their prefix.
func (g *Group) StaticFS(pathPrefix string, filesystem fs.FS) RouteInfo {
	return g.Add(
		http.MethodGet,
		pathPrefix+"*",
		StaticDirectoryHandler(filesystem, false),
	)
}

// FileFS implements `Echo#FileFS()` for sub-routes within the Group.
func (g *Group) FileFS(path, file string, filesystem fs.FS, m ...MiddlewareFunc) RouteInfo {
	return g.GET(path, StaticFileHandler(file, filesystem), m...)
}

// File implements `Echo#File()` for sub-routes within the Group. Panics on error.
func (g *Group) File(path, file string, middleware ...MiddlewareFunc) RouteInfo {
	handler := func(c Context) error {
		return c.File(file)
	}
	return g.Add(http.MethodGet, path, handler, middleware...)
}

// RouteNotFound implements `Echo#RouteNotFound()` for sub-routes within the Group.
//
// Example: `g.RouteNotFound("/*", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })`
func (g *Group) RouteNotFound(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return g.Add(RouteNotFound, path, h, m...)
}

// Add implements `Echo#Add()` for sub-routes within the Group. Panics on error.
func (g *Group) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouteInfo {
	ri, err := g.AddRoute(Route{
		Method:      method,
		Path:        path,
		Handler:     handler,
		Middlewares: middleware,
	})
	if err != nil {
		panic(err) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ri
}

// AddRoute registers a new Routable with Router
func (g *Group) AddRoute(route Routable) (RouteInfo, error) {
	// Combine middleware into a new slice to avoid accidentally passing the same slice for
	// multiple routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	groupRoute := route.ForGroup(g.prefix, append([]MiddlewareFunc{}, g.middleware...))
	return g.echo.add(g.host, groupRoute)
}

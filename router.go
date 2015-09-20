package echo

import "net/http"

type (
	Router struct {
		tree   *node
		routes []Route
		echo   *Echo
	}
	node struct {
		typ      ntype
		label    byte
		prefix   string
		parent   *node
		children children
		//handler  map[string]HandlerFunc
		handler *handler
		pnames  []string
		echo    *Echo
	}
	ntype    uint8
	children []*node
)

type handler struct {
	Connect        HandlerFunc
	Delete         HandlerFunc
	Get            HandlerFunc
	Head           HandlerFunc
	Options        HandlerFunc
	Patch          HandlerFunc
	Post           HandlerFunc
	Put            HandlerFunc
	Trace          HandlerFunc
	allowedMethods string
}

func (h *handler) CopyTo(v *handler) {
	v.Get = h.Get
	v.Connect = h.Connect
	v.Delete = h.Delete
	v.Get = h.Get
	v.Head = h.Head
	v.Options = h.Options
	v.Patch = h.Patch
	v.Post = h.Post
	v.Put = h.Put
	v.Trace = h.Trace
	v.allowedMethods = h.allowedMethods
}

func (h *handler) addToAllowedMethods(method string) {
	if h.allowedMethods == "" {
		h.allowedMethods = method
	} else {
		h.allowedMethods = h.allowedMethods + ", " + method
	}
}

func (h *handler) AddMethodHandler(method string, handler HandlerFunc) {
	if h != nil {
		if method == GET {
			h.addToAllowedMethods(method)
			h.Get = handler
		}
		if method == HEAD {
			h.addToAllowedMethods(method)
			h.Head = handler
		}
		if method == POST {
			h.addToAllowedMethods(method)
			h.Post = handler
		}
		if method == OPTIONS {
			h.addToAllowedMethods(method)
			h.Options = handler
		}
		if method == PUT {
			h.addToAllowedMethods(method)
			h.Put = handler
		}
		if method == DELETE {
			h.addToAllowedMethods(method)
			h.Delete = handler
		}
		if method == PATCH {
			h.addToAllowedMethods(method)
			h.Patch = handler
		}
		if method == CONNECT {
			h.addToAllowedMethods(method)
			h.Connect = handler
		}
		if method == TRACE {
			h.addToAllowedMethods(method)
			h.Trace = handler
		}
	}
}

func (h *handler) GetMethodHandler(method string) (HandlerFunc, string) {
	l := len(method)
	firstChar := method[0]
	secondChar := method[1]
	if l == 3 {
		if uint16(firstChar)<<8|uint16(secondChar) == 0x4745 {
			return h.Get, ""
		}
		if uint16(firstChar)<<8|uint16(secondChar) == 0x5055 {
			return h.Put, ""
		}
	} else if l == 4 {
		if uint16(firstChar)<<8|uint16(secondChar) == 0x504f {
			return h.Post, ""
		}
		if uint16(firstChar)<<8|uint16(secondChar) == 0x4845 {
			return h.Head, ""
		}
	} else if l == 5 {
		if uint16(firstChar)<<8|uint16(secondChar) == 0x5452 {
			return h.Trace, ""
		}
		if uint16(firstChar)<<8|uint16(secondChar) == 0x5041 {
			return h.Patch, ""
		}
	} else if l == 6 {
		if uint16(firstChar)<<8|uint16(secondChar) == 0x4445 {
			return h.Delete, ""
		}
	} else if l == 7 {
		if uint16(firstChar)<<8|uint16(secondChar) == 0x4f50 {
			return h.Options, ""
		}
		if uint16(firstChar)<<8|uint16(secondChar) == 0x434f {
			return h.Connect, ""
		}
	}
	return nil, h.allowedMethods
}

const (
	stype ntype = iota
	ptype
	mtype
)

func NewRouter(e *Echo) *Router {
	return &Router{
		// tree is base node for the search tree for all routes, each node
		// therein contains a handler string->HandlerFunc map.  This allows
		// us to include the method applicable within the tree, allowing us to
		// detect if routes should not allow particular methods, and making the
		// router more clear
		tree: &node{
			//handler: make(map[string]HandlerFunc),
			handler: new(handler),
		},
		routes: []Route{},
		echo:   e,
	}
}

func (r *Router) Add(method, path string, h HandlerFunc, e *Echo) {
	pnames := []string{} // Param names

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, stype, nil, e)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, ptype, pnames, e)
				return
			}
			r.insert(method, path[:i], nil, ptype, pnames, e)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, stype, nil, e)
			pnames = append(pnames, "_name")
			r.insert(method, path[:i+1], h, mtype, pnames, e)
			return
		}
	}

	r.insert(method, path, h, stype, pnames, e)
}

func (r *Router) insert(method, path string, h HandlerFunc, t ntype, pnames []string, e *Echo) {
	// Adjust max param
	l := len(pnames)
	if *e.maxParam < l {
		*e.maxParam = l
	}

	cn := r.tree
	if !validMethod(method) {
		panic("echo => invalid method")
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.typ = t
				// handler is a map of methods to applicable handlers, map the inserted method to the
				// handler
				//cn.handler = map[string]HandlerFunc{method: h}
				cn.handler = new(handler)
				cn.handler.AddMethodHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			}
		} else if l < pl {
			// Split node
			//newHandler := map[string]HandlerFunc{}
			newHandler := new(handler)
			//for k, v := range cn.handler {
			//newHandler[k] = v
			//}
			cn.handler.CopyTo(newHandler)
			n := newNode(cn.typ, cn.prefix[l:], cn, cn.children, newHandler, cn.pnames, cn.echo)

			// Reset parent node
			cn.typ = stype
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			//cn.handler = map[string]HandlerFunc{}
			cn.handler = new(handler)
			cn.pnames = nil
			cn.echo = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.typ = t
				// add the handler to the node's map of methods to handlers
				//cn.handler[method] = h
				cn.handler.AddMethodHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			} else {
				// Create child node
				newHandler := new(handler)
				newHandler.AddMethodHandler(method, h)
				n = newNode(t, search[l:], cn, nil, newHandler, pnames, e)
				cn.addChild(n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			newHandler := new(handler)
			newHandler.AddMethodHandler(method, h)
			n := newNode(t, search, cn, nil, newHandler, pnames, e)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				// add the handler to the node's map of methods to handlers
				//cn.handler[method] = h
				cn.handler.AddMethodHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			}
		}
		return
	}
}

// newNode - create a new router tree node
func newNode(t ntype, pre string, p *node, c children, h *handler, pnames []string, e *Echo) *node {
	return &node{
		typ:      t,
		label:    pre[0],
		prefix:   pre,
		parent:   p,
		children: c,
		// create a handler method to handler map for this node
		handler: h,
		pnames:  pnames,
		echo:    e,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t ntype) *node {
	for _, c := range n.children {
		if c.label == l && c.typ == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithType(t ntype) *node {
	for _, c := range n.children {
		if c.typ == t {
			return c
		}
	}
	return nil
}

//validMethod - validate that the http method is valid.
func validMethod(method string) bool {
	var ok = false
	for _, v := range methods {
		if v == method {
			ok = true
			break
		}
	}
	return ok
}

func (r *Router) Find(method, path string, ctx *Context) (h HandlerFunc, e *Echo) {
	// get tree base node from the router
	cn := r.tree

	e = cn.echo
	h = notFoundHandler

	if !validMethod(method) {
		// if the method is completely invalid
		h = methodNotAllowedHandler(ctx, cn.handler.allowedMethods)
		return
	}

	// Strip trailing slash
	if r.echo.stripTrailingSlash {
		l := len(path)
		if path[l-1] == '/' {
			path = path[:l-1]
		}
	}

	var (
		search = path
		c      *node  // Child node
		n      int    // Param counter
		nt     ntype  // Next type
		nn     *node  // Next node
		ns     string // Next search
	)

	// TODO: Check empty path???

	// Search order static > param > match-any
	for {
		if search == "" {
			if cn.handler != nil {
				// Found route, check if method is applicable
				//var ok = false
				//h, ok = cn.handler[method]
				e = cn.echo
				//if !ok {
				theHandler, allowedMethods := cn.handler.GetMethodHandler(method)
				if theHandler == nil {
					// route is valid, but method is not allowed, 405
					h = methodNotAllowedHandler(ctx, allowedMethods)
					return
				}
				ctx.pnames = cn.pnames
				h = theHandler
			}
			return
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			cn = nn
			search = ns
			if nt == ptype {
				goto Param
			} else if nt == mtype {
				goto MatchAny
			} else {
				// Not found
				return
			}
		}

		if search == "" {
			// TODO: Needs improvement
			if cn.findChildWithType(mtype) == nil {
				continue
			}
			// Empty value
			goto MatchAny
		}

		// Static node
		c = cn.findChild(search[0], stype)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nt = ptype
				nn = cn
				ns = search
			}
			cn = c
			continue
		}

		// Param node
	Param:
		c = cn.findChildWithType(ptype)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nt = mtype
				nn = cn
				ns = search
			}
			cn = c
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			ctx.pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Match-any node
	MatchAny:
		//		c = cn.getChild()
		c = cn.findChildWithType(mtype)
		if c != nil {
			cn = c
			ctx.pvalues[len(ctx.pvalues)-1] = search
			search = "" // End search
			continue
		}

		// Not found
		return
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := r.echo.pool.Get().(*Context)
	h, _ := r.Find(req.Method, req.URL.Path, c)
	c.reset(req, w, r.echo)
	if err := h(c); err != nil {
		r.echo.httpErrorHandler(err, c)
	}
	r.echo.pool.Put(c)
}

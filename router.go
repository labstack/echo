package echo

import "net/http"

type (
	Router struct {
		tree   *node
		routes []Route
		echo   *Echo
	}
	node struct {
		typ           ntype
		label         byte
		prefix        string
		parent        *node
		children      children
		methodHandler *methodHandler
		pnames        []string
		echo          *Echo
	}
	ntype         uint8
	children      []*node
	methodHandler struct {
		connect HandlerFunc
		delete  HandlerFunc
		get     HandlerFunc
		head    HandlerFunc
		options HandlerFunc
		patch   HandlerFunc
		post    HandlerFunc
		put     HandlerFunc
		trace   HandlerFunc
	}
)

const (
	stype ntype = iota
	ptype
	mtype
)

func NewRouter(e *Echo) *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
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
			pnames = append(pnames, "_*")
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

	cn := r.tree // Current node as root
	if cn == nil {
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
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.typ, cn.prefix[l:], cn, cn.children, cn.methodHandler, cn.pnames, cn.echo)

			// Reset parent node
			cn.typ = stype
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.pnames = nil
			cn.echo = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.typ = t
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodHandler), pnames, e)
				n.addHandler(method, h)
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
			n := newNode(t, search, cn, nil, new(methodHandler), pnames, e)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			}
		}
		return
	}
}

func newNode(t ntype, pre string, p *node, c children, mh *methodHandler, pnames []string, e *Echo) *node {
	return &node{
		typ:           t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		methodHandler: mh,
		pnames:        pnames,
		echo:          e,
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

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case GET:
		n.methodHandler.get = h
	case POST:
		n.methodHandler.post = h
	case PUT:
		n.methodHandler.put = h
	case DELETE:
		n.methodHandler.delete = h
	case PATCH:
		n.methodHandler.patch = h
	case OPTIONS:
		n.methodHandler.delete = h
	case HEAD:
		n.methodHandler.head = h
	case CONNECT:
		n.methodHandler.connect = h
	case TRACE:
		n.methodHandler.trace = h
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	case GET:
		return n.methodHandler.get
	case POST:
		return n.methodHandler.post
	case PUT:
		return n.methodHandler.put
	case DELETE:
		return n.methodHandler.delete
	case PATCH:
		return n.methodHandler.patch
	case OPTIONS:
		return n.methodHandler.delete
	case HEAD:
		return n.methodHandler.head
	case CONNECT:
		return n.methodHandler.connect
	case TRACE:
		return n.methodHandler.trace
	default:
		return nil
	}
}

func (r *Router) Find(method, path string, ctx *Context) (h HandlerFunc, e *Echo) {
	h = notFoundHandler
	e = r.echo
	cn := r.tree // Current node as root

	// Strip trailing slash
	if r.echo.stripTrailingSlash {
		l := len(path) - 1
		if path != "/" && path[l] == '/' { // Issue #218
			path = path[:l]
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

	// Search order static > param > match-any
	for {
		if search == "" {
			goto Found
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
			goto Found
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
		// c = cn.getChild()
		if cn = cn.findChildWithType(mtype); cn == nil {
			// Not found
			return
		}
		ctx.pvalues[len(cn.pnames)-1] = search
		goto Found
	}

Found:
	ctx.pnames = cn.pnames
	h = cn.findHandler(method)
	if cn.echo != nil {
		e = cn.echo
	}
	if h == nil {
		h = methodNotAllowedHandler
		// Dig further for match-any, might have an empty value for *, e.g.
		// serving a directory. Issue #207
		if cn = cn.findChildWithType(mtype); cn == nil {
			return
		}
//		println("here...")
//		if cn.echo != nil {
//			e = cn.echo
//		}
		h = cn.findHandler(method)
		ctx.pvalues[len(cn.pnames)-1] = ""
	}
	return
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

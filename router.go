package echo

import "net/http"

type (
	Router struct {
		tree   *node
		routes []Route
		echo   *Echo
	}
	node struct {
		kind          kind
		label         byte
		prefix        string
		parent        *node
		children      children
		methodHandler *methodHandler
		pnames        []string
		echo          *Echo
		fullPath      string
	}
	kind          uint8
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
	skind kind = iota
	pkind
	mkind
)

func NewRouter(e *Echo) *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
			fullPath:      "/",
		},
		routes: []Route{},
		echo:   e,
	}
}

func (r *Router) Add(method, path string, h HandlerFunc, e *Echo) {
	pnames := []string{} // Param names

	// pristine path without interpolated params for logging
	pristPath := path[:]

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, nil, e, pristPath)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, pnames, e, pristPath)
				return
			}
			r.insert(method, path[:i], nil, pkind, pnames, e, pristPath)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, nil, e, pristPath)
			pnames = append(pnames, "_*")
			r.insert(method, path[:i+1], h, mkind, pnames, e, pristPath)
			return
		}
	}

	r.insert(method, path, h, skind, pnames, e, pristPath)
}

func (r *Router) insert(method, path string, h HandlerFunc, t kind, pnames []string, e *Echo, pristPath string) {
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
			cn.fullPath = pristPath
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children, cn.methodHandler, cn.pnames, cn.echo, cn.fullPath)

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.pnames = nil
			cn.echo = nil
			cn.fullPath = pristPath

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodHandler), pnames, e, pristPath)
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
			n := newNode(t, search, cn, nil, new(methodHandler), pnames, e, pristPath)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.pnames = pnames
				cn.echo = e
				cn.fullPath = pristPath
			}
		}
		return
	}
}

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, pnames []string, e *Echo, fp string) *node {
	return &node{
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		methodHandler: mh,
		pnames:        pnames,
		echo:          e,
		fullPath:      fp,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
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

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
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
		n.methodHandler.options = h
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
		return n.methodHandler.options
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

func (r *Router) Find(method, path string, ctx *Context) (h HandlerFunc, e *Echo, routerPath string) {
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
		nk     kind   // Next kind
		nn     *node  // Next node
		ns     string // Next search
	)

	// Search order static > param > match-any
	for {
		if search == "" {
			goto End
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
			if nk == pkind {
				goto Param
			} else if nk == mkind {
				goto MatchAny
			} else {
				// Not found
				return
			}
		}

		if search == "" {
			goto End
		}

		// Static node
		c = cn.findChild(search[0], skind)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nk = pkind
				nn = cn
				ns = search
			}
			cn = c
			continue
		}

		// Param node
	Param:
		c = cn.findChildByKind(pkind)
		if c != nil {
			// Save next
			if cn.label == '/' {
				nk = mkind
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
		if cn = cn.findChildByKind(mkind); cn == nil {
			// Not found
			return
		}
		ctx.pvalues[len(cn.pnames)-1] = search
		goto End
	}

End:
	ctx.pnames = cn.pnames
	h = cn.findHandler(method)
	if cn.echo != nil {
		e = cn.echo
	}
	if h == nil {
		h = methodNotAllowedHandler
		// Dig further for match-any, might have an empty value for *, e.g.
		// serving a directory. Issue #207.
		if cn = cn.findChildByKind(mkind); cn == nil {
			return
		}
		ctx.pvalues[len(cn.pnames)-1] = ""
		if h = cn.findHandler(method); h == nil {
			h = methodNotAllowedHandler
		}
	}
	routerPath = cn.fullPath
	return
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := r.echo.pool.Get().(*Context)
	h, _, rp := r.Find(req.Method, req.URL.Path, c)
	c.reset(req, w, r.echo, rp)
	if err := h(c); err != nil {
		r.echo.httpErrorHandler(err, c)
	}
	r.echo.pool.Put(c)
}

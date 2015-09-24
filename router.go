package echo

import "net/http"

type (
	Router struct {
		connectTree *node
		deleteTree  *node
		getTree     *node
		headTree    *node
		optionsTree *node
		patchTree   *node
		postTree    *node
		putTree     *node
		traceTree   *node
		routes      []Route
		echo        *Echo
	}
	node struct {
		typ      ntype
		label    byte
		prefix   string
		parent   *node
		children children
		handler  HandlerFunc
		pnames   []string
		echo     *Echo
	}
	ntype    uint8
	children []*node
)

const (
	stype ntype = iota
	ptype
	mtype
)

func NewRouter(e *Echo) *Router {
	return &Router{
		connectTree: new(node),
		deleteTree:  new(node),
		getTree:     new(node),
		headTree:    new(node),
		optionsTree: new(node),
		patchTree:   new(node),
		postTree:    new(node),
		putTree:     new(node),
		traceTree:   new(node),
		routes:      []Route{},
		echo:        e,
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

	cn := r.findTree(method) // Current node as root
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
				cn.handler = h
				cn.pnames = pnames
				cn.echo = e
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.typ, cn.prefix[l:], cn, cn.children, cn.handler, cn.pnames, cn.echo)

			// Reset parent node
			cn.typ = stype
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.handler = nil
			cn.pnames = nil
			cn.echo = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.typ = t
				cn.handler = h
				cn.pnames = pnames
				cn.echo = e
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, h, pnames, e)
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
			n := newNode(t, search, cn, nil, h, pnames, e)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.handler = h
				cn.pnames = pnames
				cn.echo = e
			}
		}
		return
	}
}

func newNode(t ntype, pre string, p *node, c children, h HandlerFunc, pnames []string, e *Echo) *node {
	return &node{
		typ:      t,
		label:    pre[0],
		prefix:   pre,
		parent:   p,
		children: c,
		handler:  h,
		pnames:   pnames,
		echo:     e,
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

func (r *Router) findTree(method string) (n *node) {
	switch method[0] {
	case 'G': // GET
		m := uint32(method[2])<<8 | uint32(method[1])<<16 | uint32(method[0])<<24
		if m == 0x47455400 {
			n = r.getTree
		}
	case 'P': // POST, PUT or PATCH
		switch method[1] {
		case 'O': // POST
			m := uint32(method[3]) | uint32(method[2])<<8 | uint32(method[1])<<16 |
				uint32(method[0])<<24
			if m == 0x504f5354 {
				n = r.postTree
			}
		case 'U': // PUT
			m := uint32(method[2])<<8 | uint32(method[1])<<16 | uint32(method[0])<<24
			if m == 0x50555400 {
				n = r.putTree
			}
		case 'A': // PATCH
			m := uint64(method[4])<<24 | uint64(method[3])<<32 | uint64(method[2])<<40 |
				uint64(method[1])<<48 | uint64(method[0])<<56
			if m == 0x5041544348000000 {
				n = r.patchTree
			}
		}
	case 'D': // DELETE
		m := uint64(method[5])<<16 | uint64(method[4])<<24 | uint64(method[3])<<32 |
			uint64(method[2])<<40 | uint64(method[1])<<48 | uint64(method[0])<<56
		if m == 0x44454c4554450000 {
			n = r.deleteTree
		}
	case 'C': // CONNECT
		m := uint64(method[6])<<8 | uint64(method[5])<<16 | uint64(method[4])<<24 |
			uint64(method[3])<<32 | uint64(method[2])<<40 | uint64(method[1])<<48 |
			uint64(method[0])<<56
		if m == 0x434f4e4e45435400 {
			n = r.connectTree
		}
	case 'H': // HEAD
		m := uint32(method[3]) | uint32(method[2])<<8 | uint32(method[1])<<16 |
			uint32(method[0])<<24
		if m == 0x48454144 {
			n = r.headTree
		}
	case 'O': // OPTIONS
		m := uint64(method[6])<<8 | uint64(method[5])<<16 | uint64(method[4])<<24 |
			uint64(method[3])<<32 | uint64(method[2])<<40 | uint64(method[1])<<48 |
			uint64(method[0])<<56
		if m == 0x4f5054494f4e5300 {
			n = r.optionsTree
		}
	case 'T': // TRACE
		m := uint64(method[4])<<24 | uint64(method[3])<<32 | uint64(method[2])<<40 |
			uint64(method[1])<<48 | uint64(method[0])<<56
		if m == 0x5452414345000000 {
			n = r.traceTree
		}
	}
	return
}

func (r *Router) Find(method, path string, ctx *Context) (h HandlerFunc, e *Echo) {
	h = notFoundHandler
	cn := r.findTree(method) // Current node as root
	if cn == nil {
		h = badRequestHandler
		return
	}

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
			if cn.handler == nil {
				// Look up for match-any, might have an empty value for *, e.g.
				// serving a directory. Issue #207
				if cn = cn.findChildWithType(mtype); cn == nil {
					return
				}
				ctx.pvalues[len(cn.pnames)-1] = ""
			}
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
	h = cn.handler
	e = cn.echo
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

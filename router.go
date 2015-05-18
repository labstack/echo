package echo

import "net/http"

type (
	router struct {
		trees map[string]*node
		echo  *Echo
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

func NewRouter(e *Echo) (r *router) {
	r = &router{
		trees: make(map[string]*node),
		echo:  e,
	}
	for _, m := range methods {
		r.trees[m] = &node{
			prefix:   "",
			children: children{},
		}
	}
	return
}

func (r *router) Add(method, path string, h HandlerFunc, echo *Echo) {
	var pnames []string // Param names

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, stype, nil, echo)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, ptype, pnames, echo)
				return
			}
			r.insert(method, path[:i], nil, ptype, pnames, echo)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, stype, nil, echo)
			pnames = append(pnames, "_name")
			r.insert(method, path[:i+1], h, mtype, pnames, echo)
			return
		}
	}
	r.insert(method, path, h, stype, pnames, echo)
}

func (r *router) insert(method, path string, h HandlerFunc, t ntype, pnames []string, echo *Echo) {
	cn := r.trees[method] // Current node as root
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := lcp(search, cn.prefix)

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.typ = t
				cn.handler = h
				cn.pnames = pnames
				cn.echo = echo
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.typ, cn.prefix[l:], cn, cn.children, cn.handler, cn.pnames, cn.echo)
			cn.children = children{n} // Add to parent

			// Reset parent node
			cn.typ = stype
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.handler = nil
			cn.pnames = nil
			cn.echo = nil

			if l == sl {
				// At parent node
				cn.typ = t
				cn.handler = h
				cn.pnames = pnames
				cn.echo = echo
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, h, pnames, echo)
				cn.children = append(cn.children, n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChild(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(t, search, cn, nil, h, pnames, echo)
			cn.children = append(cn.children, n)
		} else {
			// Node already exists
			if h != nil {
				cn.handler = h
				cn.pnames = pnames
				cn.echo = echo
			}
		}
		return
	}
}

func newNode(t ntype, pre string, p *node, c children, h HandlerFunc, pnames []string, echo *Echo) *node {
	return &node{
		typ:      t,
		label:    pre[0],
		prefix:   pre,
		parent:   p,
		children: c,
		handler:  h,
		pnames:   pnames,
		echo:     echo,
	}
}

func (n *node) addChild(c *node) {
}

func (n *node) findChild(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findSchild(l byte) *node {
	for _, c := range n.children {
		if c.label == l && c.typ == stype {
			return c
		}
	}
	return nil
}

func (n *node) findPchild() *node {
	for _, c := range n.children {
		if c.typ == ptype {
			return c
		}
	}
	return nil
}

func (n *node) findMchild() *node {
	for _, c := range n.children {
		if c.typ == mtype {
			return c
		}
	}
	return nil
}

// Length of longest common prefix
func lcp(a, b string) (i int) {
	max := len(a)
	l := len(b)
	if l < max {
		max = l
	}
	for ; i < max && a[i] == b[i]; i++ {
	}
	return
}

func (r *router) Find(method, path string, ctx *Context) (h HandlerFunc, echo *Echo) {
	cn := r.trees[method] // Current node as root
	search := path

	var (
		c  *node  // Child node
		n  int    // Param counter
		nt ntype  // Next type
		nn *node  // Next node
		ns string // Next search
	)

	// TODO: Check empty path???

	// Search order static > param > match-any
	for {
		if search == "" {
			// Found
			ctx.pnames = cn.pnames
			h = cn.handler
			echo = cn.echo
			return
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			pl = len(cn.prefix)
			l = lcp(search, cn.prefix)
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
			if cn.findMchild() == nil {
				continue
			}
			// Empty value
			goto MatchAny
		}

		// Static node
		c = cn.findSchild(search[0])
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
		c = cn.findPchild()
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
		c = cn.findMchild()
		if c != nil {
			cn = c
			ctx.pvalues[n] = search
			search = "" // End search
			continue
		}
		// Not found
		return
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := r.echo.pool.Get().(*Context)
	h, _ := r.Find(req.Method, req.URL.Path, c)
	c.reset(w, req, r.echo)
	if h == nil {
		h = r.echo.notFoundHandler
	}
	h(c)
	r.echo.pool.Put(c)
}

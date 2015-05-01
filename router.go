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
		// pchild   *node // Param child
		// mchild   *node // Match-any child
		handler HandlerFunc
		pnames  []string
		echo    *Echo
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
			pnames = append(pnames, "_name")
			r.insert(method, path[:i], h, mtype, pnames, echo)
			return
		}
	}
	r.insert(method, path, h, stype, nil, echo)
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
			// if n.typ == ptype {
			// cn.pchild = n
			// } else if n.typ == ctype {
			// cn.cchild = n
			// }

			// Reset parent node
			cn.typ = stype
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			// cn.pchild = nil
			// cn.cchild = nil
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
				n = newNode(t, search[l:], cn, children{}, h, pnames, echo)
				cn.children = append(cn.children, n)
				// if n.typ == ptype {
				// cn.pchild = n
				// } else if n.typ == ctype {
				// cn.cchild = n
				// }
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
			n := newNode(t, search, cn, children{}, h, pnames, echo)
			cn.children = append(cn.children, n)
			// if n.typ == ptype {
			// cn.pchild = n
			// } else if n.typ == ctype {
			// cn.cchild = n
			// }
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

func newNode(t ntype, pfx string, p *node, c children, h HandlerFunc, pnames []string, echo *Echo) *node {
	return &node{
		typ:      t,
		label:    pfx[0],
		prefix:   pfx,
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
	c := new(node) // Child node
	n := 0         // Param counter

	// Search order static > param > match-any
	for {
		// TODO flip condition???
		if search == "" || search == cn.prefix || cn.typ == mtype {
			// Found
			h = cn.handler
			echo = cn.echo
			ctx.pnames = cn.pnames

			// Match-any node
			if cn.typ == mtype {
				ctx.pvalues[0] = search[len(cn.prefix):]
			}

			return
		}

		pl := len(cn.prefix)
		l := lcp(search, cn.prefix)

		if l == pl {
			search = search[l:]
		} else if l < pl && cn.label != ':' {
			goto Up
		}

		// Static node
		c = cn.findSchild(search[0])
		if c != nil {
			cn = c
			continue
		}

		// Param node
	Param:
		c = cn.findPchild()
		// c = cn.pchild
		if c != nil {
			cn = c
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			ctx.pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Match-any
		c = cn.findMchild()
		if c != nil {
			cn = c
			continue
		}

	Up:
		tn := cn // Save current node
		cn = cn.parent
		if cn == nil {
			// Not found
			return
		}
		// Search upwards
		if l == pl {
			// Reset search
			search = tn.prefix + search
		}
		goto Param
	}
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := r.echo.pool.Get().(*Context)
	h, _ := r.Find(req.Method, req.URL.Path, c)
	c.reset(w, req, nil)
	if h != nil {
		h(c)
	} else {
		r.echo.notFoundHandler(c)
	}
	r.echo.pool.Put(c)
}

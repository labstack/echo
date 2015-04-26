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
		// pchild   *node    // Param child
		// cchild   *node    // Catch-all child
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
	ctype
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
			r.insert(method, path[:l], h, ctype, pnames, echo)
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
			n := newNode(t, cn.prefix[l:], cn, cn.children, cn.handler, cn.pnames, cn.echo)
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
				n = newNode(t, search[l:], cn, children{}, h, pnames, echo)
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
			n := newNode(t, search, cn, children{}, h, pnames, echo)
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

func newNode(t ntype, pfx string, p *node, c children, h HandlerFunc, pnames []string, echo *Echo) (n *node) {
	n = &node{
		typ:      t,
		label:    pfx[0],
		prefix:   pfx,
		parent:   p,
		children: c,
		handler:  h,
		pnames:   pnames,
		echo:     echo,
	}
	return
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

func (n *node) findCchild() *node {
	for _, c := range n.children {
		if c.typ == ctype {
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

func (r *router) Find(method, path string, c *Context) (h HandlerFunc, echo *Echo) {
	cn := r.trees[method] // Current node as root
	search := path
	chn := new(node) // Child node
	n := 0           // Param counter

	// Search order static > param > catch-all
	for {
		if search == "" || search == cn.prefix {
			// Found
			h = cn.handler
			c.pnames = cn.pnames
			echo = cn.echo
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
		chn = cn.findSchild(search[0])
		if chn != nil {
			cn = chn
			continue
		}

		// Param node
	Param:
		chn = cn.findPchild()
		if chn != nil {
			cn = chn
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			c.pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Catch-all node
		chn = cn.findCchild()
		if chn != nil {
			cn = chn
			c.pvalues[n] = search
			search = "" // End search
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

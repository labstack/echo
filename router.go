package echo

import "net/http"

type (
	router struct {
		trees map[string]*node
		echo  *Echo
	}
	node struct {
		label    byte
		prefix   string
		parent   *node
		children children
		handler  HandlerFunc
		echo     *Echo
	}
	children []*node
	param    struct {
		Name  string
		Value string
	}
	Params []param
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
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			r.insert(method, path[:i], nil, echo)
			for ; i < l && path[i] != '/'; i++ {
			}
			if i == l {
				r.insert(method, path[:i], h, echo)
				return
			}
			r.insert(method, path[:i], nil, echo)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, echo)
			r.insert(method, path[:l], h, echo)
		}
	}
	r.insert(method, path, h, echo)
}

func (r *router) insert(method, path string, h HandlerFunc, echo *Echo) {
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
				cn.handler = h
				cn.echo = echo
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.prefix[l:], cn, cn.children, cn.handler, cn.echo)
			cn.children = children{n} // Add to parent

			// Reset parent node
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.handler = nil
			cn.echo = nil

			if l == sl {
				// At parent node
				cn.handler = h
				cn.echo = echo
			} else {
				// Create child node
				n = newNode(search[l:], cn, children{}, h, echo)
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
			n := newNode(search, cn, children{}, h, echo)
			cn.children = append(cn.children, n)
		} else {
			// Node already exists
			if h != nil {
				cn.handler = h
				cn.echo = echo
			}
		}
		return
	}
}

func newNode(pfx string, p *node, c children, h HandlerFunc, echo *Echo) (n *node) {
	n = &node{
		label:    pfx[0],
		prefix:   pfx,
		parent:   p,
		children: c,
		handler:  h,
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

func (r *router) Find(method, path string, params Params) (h HandlerFunc, echo *Echo) {
	cn := r.trees[method] // Current node as root
	search := path
	n := 0         // Param count
	c := new(node) // Child node

	// Search order static > param > catch-all
	for {
		if search == "" || search == cn.prefix {
			// Found
			h = cn.handler
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
		c = cn.findChild(search[0])
		if c != nil {
			cn = c
			continue
		}

		// Param node
	Param:
		c = cn.findChild(':')
		if c != nil {
			cn = c
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			params[n].Name = cn.prefix[1:]
			params[n].Value = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Catch-all node
		c = cn.findChild('*')
		if c != nil {
			cn = c
			p := params[:n+1]
			p[n].Name = "_name"
			p[n].Value = search
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
	h, _ := r.Find(req.Method, req.URL.Path, c.params)
	c.Response.Writer = w
	if h != nil {
		h(c)
	} else {
		r.echo.notFoundHandler(c)
	}
	r.echo.pool.Put(c)
}

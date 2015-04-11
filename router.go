package echo

import "net/http"

type (
	router struct {
		trees map[string]*node
		echo  *Echo
	}
	node struct {
		label   byte
		prefix  string
		has     ntype // Type of node it contains
		handler HandlerFunc
		edges   edges
		echo    *Echo
	}
	edges []*node
	ntype byte
	param struct {
		Name  string
		Value string
	}
	Params []param
)

const (
	snode ntype = 1 + iota // Static node
	pnode                  // Param node
	cnode                  // Catch-all node
)

func NewRouter(e *Echo) (r *router) {
	r = &router{
		trees: make(map[string]*node),
		echo:  e,
	}
	for _, m := range methods {
		r.trees[m] = &node{
			prefix: "",
			edges:  edges{},
		}
	}
	return
}

func (r *router) Add(method, path string, h HandlerFunc, echo *Echo) {
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			r.insert(method, path[:i], nil, pnode, echo)
			for ; i < l && path[i] != '/'; i++ {
			}
			if i == l {
				r.insert(method, path[:i], h, 0, echo)
				return
			}
			r.insert(method, path[:i], nil, 0, echo)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, cnode, echo)
			r.insert(method, path[:l], h, 0, echo)
		}
	}
	r.insert(method, path, h, snode, echo)
}

func (r *router) insert(method, path string, h HandlerFunc, has ntype, echo *Echo) {
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
			cn.has = has
			if h != nil {
				cn.handler = h
				cn.echo = echo
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.prefix[l:], cn.has, cn.handler, cn.edges, cn.echo)
			cn.edges = edges{n} // Add to parent

			// Reset parent node
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.has = 0
			cn.handler = nil
			cn.echo = nil

			if l == sl {
				// At parent node
				cn.handler = h
				cn.echo = echo
			} else {
				// Create child node
				n = newNode(search[l:], has, h, edges{}, echo)
				cn.edges = append(cn.edges, n)
			}
		} else if l < sl {
			search = search[l:]
			e := cn.findEdge(search[0])
			if e != nil {
				// Go deeper
				cn = e
				continue
			}
			// Create child node
			n := newNode(search, has, h, edges{}, echo)
			cn.edges = append(cn.edges, n)
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

func newNode(pfx string, has ntype, h HandlerFunc, e edges, echo *Echo) (n *node) {
	n = &node{
		label:   pfx[0],
		prefix:  pfx,
		has:     has,
		handler: h,
		edges:   e,
		echo:    echo,
	}
	return
}

func (n *node) findEdge(l byte) *node {
	for _, e := range n.edges {
		if e.label == l {
			return e
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
	n := 0 // Param count

	// Search order static > param > catch-all
	for {
		if search == "" || search == cn.prefix {
			// Found
			h = cn.handler
			echo = cn.echo
			return
		}

		l := lcp(search, cn.prefix)
		search = search[l:]

		// Static node
		e := cn.findEdge(search[0])
		if e != nil {
			// Go deeper
			cn = e
			continue
		}

		// Param node
		e = cn.findEdge(':')
		if e != nil {
			cn = e
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			p := params[:n+1]
			p[n].Name = cn.prefix[1:]
			p[n].Value = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Catch-all node
		e = cn.findEdge('*')
		if e != nil {
			cn = e
			p := params[:n+1]
			p[n].Name = "_name"
			p[n].Value = search
			search = "" // End search
			continue
		}

		return
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

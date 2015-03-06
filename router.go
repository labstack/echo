package bolt

import (
	"fmt"
	"net/http"
	"sync"
)

type (
	router struct {
		bolt *Bolt
		pool sync.Pool
		root *node
	}
	node struct {
		label    byte
		prefix   string
		has      ntype // Type of node(s)
		handlers []HandlerFunc
		edges    edges
	}
	edges []*node
	ntype byte
	param struct {
		Name  string
		Value string
	}
	Params []param
	Status uint16
)

const (
	snode ntype = iota // Static node
	pnode              // Param node
	wnode              // Wildcard node
)

const (
	OK Status = iota
	NotFound
	NotAllowed
)

func NewRouter(b *Bolt) (r *router) {
	r = &router{
		bolt: b,
		root: &node{
			prefix:   "",
			handlers: make([]HandlerFunc, len(MethodMap)),
			edges:    edges{},
		},
	}
	r.pool.New = func() interface{} {
		return make(Params, 0, b.maxParam)
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

func (r *router) Add(method, path string, h HandlerFunc) {
	i := 0
	l := len(path)
	for ; i < l; i++ {
		if path[i] == ':' {
			r.insert(method, path[:i], nil, pnode)
			for ; i < l && path[i] != '/'; i++ {
			}
			if i == l {
				r.insert(method, path[:i], h, snode)
				return
			} else {
				r.insert(method, path[:i], nil, snode)
			}
		}
	}
	r.insert(method, path, h, snode)
}

func (r *router) insert(method, path string, h HandlerFunc, has ntype) {
	cn := r.root // Current node
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
				cn.handlers[MethodMap[method]] = h
			}
			return
		} else if l < pl {
			// Split the node
			n := newNode(cn.prefix[l:], cn.has, cn.handlers, cn.edges)
			cn.edges = edges{n} // Add to parent

			// Reset parent node
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.has = snode
			cn.handlers = make([]HandlerFunc, len(MethodMap))

			if l == sl {
				// At parent node
				cn.handlers[MethodMap[method]] = h
			} else {
				// Need to fork a node
				n = newNode(search[l:], has, nil, nil)
				n.handlers[MethodMap[method]] = h
				cn.edges = append(cn.edges, n)
			}
			break
		} else if l < sl {
			search = search[l:]
			e := cn.findEdge(search[0])
			if e == nil {
				n := newNode(search, has, nil, nil)
				if h != nil {
					n.handlers[MethodMap[method]] = h
				}
				cn.edges = append(cn.edges, n)
				break
			} else {
				cn = e
			}
		} else {
			// Node already exists
			if h != nil {
				cn.handlers[MethodMap[method]] = h
			}
			break
		}
	}
}

func newNode(pfx string, has ntype, h []HandlerFunc, e edges) (n *node) {
	n = &node{
		label:    pfx[0],
		prefix:   pfx,
		has:      has,
		handlers: h,
		edges:    e,
	}
	if h == nil {
		n.handlers = make([]HandlerFunc, len(MethodMap))
	}
	if e == nil {
		n.edges = edges{}
	}
	return
}

func (r *router) Find(method, path string) (handler HandlerFunc, c *Context, s Status) {
	c = r.bolt.pool.Get().(*Context)
	cn := r.root // Current node
	search := path
	n := 0 // Param count

	for {
		if search == "" || search == cn.prefix {
			// Node found
			h := cn.handlers[MethodMap[method]]
			if h != nil {
				// Handler found
				handler = h
			} else {
				s = NotAllowed
			}
			return
		}

		pl := len(cn.prefix)
		l := lcp(search, cn.prefix)

		if l == pl {
			search = search[l:]
			switch cn.has {
			case pnode:
				cn = cn.edges[0]
				i := 0
				l = len(search)

				for ; i < l && search[i] != '/'; i++ {
				}
				p := c.params[:n+1]
				p[n].Name = cn.prefix[1:]
				p[n].Value = search[:i]
				n++

				search = search[i:]

				if i == l {
					// All param read
					continue
				}
			}
			e := cn.findEdge(search[0])
			if e == nil {
				// Not found
				s = NotFound
				return
			} else {
				cn = e
				continue
			}
		} else {
			// Not found
			s = NotFound
			return
		}
	}
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

func (r *router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h, c, rep := r.Find(req.Method, req.URL.Path)
	c.Response.ResponseWriter = rw
	if h != nil {
		h(c)
	} else {
		if rep == NotFound {
			r.bolt.notFoundHandler(c)
		} else if rep == NotAllowed {
			r.bolt.methodNotAllowedHandler(c)
		}
	}
	r.bolt.pool.Put(c)
}

func (ps Params) Get(n string) (v string) {
	for _, p := range ps {
		if p.Name == n {
			v = p.Value
		}
	}
	return
}

func (r *router) printTree() {
	r.root.printTree("", true)
}

func (n *node) printTree(pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	fmt.Printf("%s%s has=%d, len=%d\n", p, n.prefix, n.has, len(n.handlers))

	nodes := n.edges
	l := len(nodes)
	p = prefix(tail, pfx, "    ", "│   ")
	for i := 0; i < l-1; i++ {
		nodes[i].printTree(p, false)
	}
	if l > 0 {
		nodes[l-1].printTree(p, true)
	}
}

func prefix(tail bool, p, on, off string) string {
	if tail {
		return fmt.Sprintf("%s%s", p, on)
	} else {
		return fmt.Sprintf("%s%s", p, off)
	}
}

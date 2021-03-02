package echo

import (
	"net/http"
)

type (
	// Router is the registry of all registered routes for an `Echo` instance for
	// request matching and URL path parameter parsing.
	Router struct {
		tree   *node
		routes map[string]*Route
		echo   *Echo
	}
	node struct {
		kind            kind
		label           byte
		prefix          string
		parent          *node
		staticChildrens children
		ppath           string
		pnames          []string
		methodHandler   *methodHandler
		paramChildren   *node
		anyChildren     *node
	}
	kind          uint8
	children      []*node
	methodHandler struct {
		connect  HandlerFunc
		delete   HandlerFunc
		get      HandlerFunc
		head     HandlerFunc
		options  HandlerFunc
		patch    HandlerFunc
		post     HandlerFunc
		propfind HandlerFunc
		put      HandlerFunc
		trace    HandlerFunc
		report   HandlerFunc
	}
)

const (
	skind kind = iota
	pkind
	akind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

// NewRouter returns a new Router instance.
func NewRouter(e *Echo) *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes: map[string]*Route{},
		echo:   e,
	}
}

// Add registers a new route for method and path with matching handler.
func (r *Router) Add(method, path string, h HandlerFunc) {
	// Validate path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames)
			} else {
				r.insert(method, path[:i], nil, pkind, "", nil)
			}
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], h, akind, ppath, pnames)
		}
	}

	r.insert(method, path, h, skind, ppath, pnames)
}

func (r *Router) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string) {
	// Adjust max param
	l := len(pnames)
	if *r.echo.maxParam < l {
		*r.echo.maxParam = l
	}

	cn := r.tree // Current node as root
	if cn == nil {
		panic("echo: invalid method")
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
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.staticChildrens, cn.methodHandler, cn.ppath, cn.pnames, cn.paramChildren, cn.anyChildren)

			// Update parent path for all children to new node
			for _, child := range cn.staticChildrens {
				child.parent = n
			}
			if cn.paramChildren != nil {
				cn.paramChildren.parent = n
			}
			if cn.anyChildren != nil {
				cn.anyChildren.parent = n
			}

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.staticChildrens = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil
			cn.paramChildren = nil
			cn.anyChildren = nil

			// Only Static children could reach here
			cn.addStaticChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodHandler), ppath, pnames, nil, nil)
				n.addHandler(method, h)
				// Only Static children could reach here
				cn.addStaticChild(n)
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
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames, nil, nil)
			n.addHandler(method, h)
			switch t {
			case skind:
				cn.addStaticChild(n)
			case pkind:
				cn.paramChildren = n
			case akind:
				cn.anyChildren = n
			}
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = ppath
				if len(cn.pnames) == 0 { // Issue #729
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

func newNode(t kind, pre string, p *node, sc children, mh *methodHandler, ppath string, pnames []string, paramChildren, anyChildren *node) *node {
	return &node{
		kind:            t,
		label:           pre[0],
		prefix:          pre,
		parent:          p,
		staticChildrens: sc,
		ppath:           ppath,
		pnames:          pnames,
		methodHandler:   mh,
		paramChildren:   paramChildren,
		anyChildren:     anyChildren,
	}
}

func (n *node) addStaticChild(c *node) {
	n.staticChildrens = append(n.staticChildrens, c)
}

func (n *node) findStaticChild(l byte) *node {
	for _, c := range n.staticChildrens {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.staticChildrens {
		if c.label == l {
			return c
		}
	}
	if l == paramLabel {
		return n.paramChildren
	}
	if l == anyLabel {
		return n.anyChildren
	}
	return nil
}

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case http.MethodConnect:
		n.methodHandler.connect = h
	case http.MethodDelete:
		n.methodHandler.delete = h
	case http.MethodGet:
		n.methodHandler.get = h
	case http.MethodHead:
		n.methodHandler.head = h
	case http.MethodOptions:
		n.methodHandler.options = h
	case http.MethodPatch:
		n.methodHandler.patch = h
	case http.MethodPost:
		n.methodHandler.post = h
	case PROPFIND:
		n.methodHandler.propfind = h
	case http.MethodPut:
		n.methodHandler.put = h
	case http.MethodTrace:
		n.methodHandler.trace = h
	case REPORT:
		n.methodHandler.report = h
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	case http.MethodConnect:
		return n.methodHandler.connect
	case http.MethodDelete:
		return n.methodHandler.delete
	case http.MethodGet:
		return n.methodHandler.get
	case http.MethodHead:
		return n.methodHandler.head
	case http.MethodOptions:
		return n.methodHandler.options
	case http.MethodPatch:
		return n.methodHandler.patch
	case http.MethodPost:
		return n.methodHandler.post
	case PROPFIND:
		return n.methodHandler.propfind
	case http.MethodPut:
		return n.methodHandler.put
	case http.MethodTrace:
		return n.methodHandler.trace
	case REPORT:
		return n.methodHandler.report
	default:
		return nil
	}
}

func (n *node) checkMethodNotAllowed() HandlerFunc {
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			return MethodNotAllowedHandler
		}
	}
	return NotFoundHandler
}

// Find lookup a handler registered for method and path. It also parses URL for path
// parameters and load them into context.
//
// For performance:
//
// - Get context from `Echo#AcquireContext()`
// - Reset it `Context#Reset()`
// - Return it `Echo#ReleaseContext()`.
func (r *Router) Find(method, path string, c Context) {
	ctx := c.(*context)
	ctx.path = path
	cn := r.tree // Current node as root

	var (
		search      = path
		searchIndex = 0
		n           int           // Param counter
		pvalues     = ctx.pvalues // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	// Backtracking is needed when a dead end (leaf node) is reached in the router tree.
	// To backtrack the current node will be changed to the parent node and the next kind for the
	// router logic will be returned based on fromKind or kind of the dead end node (static > param > any).
	// For example if there is no static node match we should check parent next sibling by kind (param).
	// Backtracking itself does not check if there is a next sibling, this is done by the router logic.
	backtrackToNextNodeKind := func(fromKind kind) (nextNodeKind kind, valid bool) {
		previous := cn
		cn = previous.parent
		valid = cn != nil

		// Next node type by priority
		// NOTE: With the current implementation we never backtrack from an `any` route, so `previous.kind` is
		// always `static` or `any`
		// If this is changed then for any route next kind would be `static` and this statement should be changed
		nextNodeKind = previous.kind + 1

		if fromKind == skind {
			// when backtracking is done from static kind block we did not change search so nothing to restore
			return
		}

		// restore search to value it was before we move to current node we are backtracking from.
		if previous.kind == skind {
			searchIndex -= len(previous.prefix)
		} else {
			n--
			// for param/any node.prefix value is always `:` so we can not deduce searchIndex from that and must use pValue
			// for that index as it would also contain part of path we cut off before moving into node we are backtracking from
			searchIndex -= len(pvalues[n])
		}
		search = path[searchIndex:]
		return
	}

	// Search order static > param > any
	for {
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

		if l != pl {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision path
			nk, ok := backtrackToNextNodeKind(skind)
			if !ok {
				return // No other possibilities on the decision path
			} else if nk == pkind {
				goto Param
				// NOTE: this case (backtracking from static node to previous any node) can not happen by current any matching logic. Any node is end of search currently
				//} else if nk == akind {
				//	goto Any
			} else {
				// Not found (this should never be possible for static node we are looking currently)
				return
			}
		}

		// The full prefix has matched, remove the prefix from the remaining search
		search = search[l:]
		searchIndex = searchIndex + l

		// Finish routing if no remaining search and we are on an leaf node
		if search == "" && cn.ppath != "" {
			break
		}

		// Static node
		if search != "" {
			if child := cn.findStaticChild(search[0]); child != nil {
				cn = child
				continue
			}
		}

	Param:
		// Param node
		if child := cn.paramChildren; search != "" && child != nil {
			cn = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			pvalues[n] = search[:i]
			n++
			search = search[i:]
			searchIndex = searchIndex + i
			continue
		}

	Any:
		// Any node
		if child := cn.anyChildren; child != nil {
			// If any node is found, use remaining path for pvalues
			cn = child
			pvalues[len(cn.pnames)-1] = search
			break
		}

		// Let's backtrack to the first possible alternative node of the decision path
		nk, ok := backtrackToNextNodeKind(akind)
		if !ok {
			return // No other possibilities on the decision path
		} else if nk == pkind {
			goto Param
		} else if nk == akind {
			goto Any
		} else {
			// Not found
			return
		}
	}

	ctx.handler = cn.findHandler(method)
	ctx.path = cn.ppath
	ctx.pnames = cn.pnames

	if ctx.handler == nil {
		ctx.handler = cn.checkMethodNotAllowed()
	}
	return
}

package middleware

import (
	"bytes"
	"net/http"
	"reflect"
	"runtime"

	"github.com/siyual-park/echo-slim/v4"
)

var (
	methods = [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		echo.PROPFIND,
		http.MethodPut,
		http.MethodTrace,
		echo.REPORT,
	}
)

type (
	// Route contains a handler and information for matching against requests.
	Route struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Name   string `json:"name"`
	}

	// Router is the registry of all registered routes for an `Echo` instance for
	// request matching and URL path parameter parsing.
	Router struct {
		middleware []echo.MiddlewareFunc
		maxParam   int
		tree       *node
		routes     map[string]*Route
	}
	node struct {
		kind           kind
		label          byte
		prefix         string
		parent         *node
		staticChildren children
		ppath          string
		pnames         []string
		methodHandler  *methodHandler
		paramChild     *node
		anyChild       *node
		// isLeaf indicates that node does not have child routes
		isLeaf bool
		// isHandler indicates that node has at least one handler registered to it
		isHandler bool
	}
	kind          uint8
	children      []*node
	methodHandler struct {
		connect     echo.MiddlewareFunc
		delete      echo.MiddlewareFunc
		get         echo.MiddlewareFunc
		head        echo.MiddlewareFunc
		options     echo.MiddlewareFunc
		patch       echo.MiddlewareFunc
		post        echo.MiddlewareFunc
		propfind    echo.MiddlewareFunc
		put         echo.MiddlewareFunc
		trace       echo.MiddlewareFunc
		report      echo.MiddlewareFunc
		allowHeader string
	}
)

const (
	staticKind kind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

var (
	MethodNotAllowedMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// See RFC 7231 section 7.4.1: An origin server MUST generate an Allow field in a 405 (Method Not Allowed)
			// response and MAY do so in any other response. For disabled resources an empty Allow header may be returned
			routerAllowMethods, ok := c.Get(echo.ContextKeyHeaderAllow).(string)
			if ok && routerAllowMethods != "" {
				c.Response().Header().Set(echo.HeaderAllow, routerAllowMethods)
			}
			return echo.ErrMethodNotAllowed
		}
	}

	NotFoundMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return err
			}
			if c.Response().Status == 0 {
				return echo.NotFoundHandler(c)
			}
			return nil
		}
	}
)

func (m *methodHandler) isHandler() bool {
	return m.connect != nil ||
		m.delete != nil ||
		m.get != nil ||
		m.head != nil ||
		m.options != nil ||
		m.patch != nil ||
		m.post != nil ||
		m.propfind != nil ||
		m.put != nil ||
		m.trace != nil ||
		m.report != nil
}

func (m *methodHandler) updateAllowHeader() {
	buf := new(bytes.Buffer)
	buf.WriteString(http.MethodOptions)

	if m.connect != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodConnect)
	}
	if m.delete != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodDelete)
	}
	if m.get != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodGet)
	}
	if m.head != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodHead)
	}
	if m.patch != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPatch)
	}
	if m.post != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPost)
	}
	if m.propfind != nil {
		buf.WriteString(", PROPFIND")
	}
	if m.put != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPut)
	}
	if m.trace != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodTrace)
	}
	if m.report != nil {
		buf.WriteString(", REPORT")
	}
	m.allowHeader = buf.String()
}

// NewRouter returns a new Router instance.
func NewRouter() *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes: map[string]*Route{},
	}
}

// Use adds middleware to the chain
func (r *Router) Use(middleware ...echo.MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) CONNECT(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodConnect, path, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (r *Router) DELETE(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodDelete, path, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (r *Router) GET(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodGet, path, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) HEAD(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodHead, path, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) OPTIONS(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodOptions, path, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) PATCH(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodPatch, path, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) POST(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodPost, path, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) PUT(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodPut, path, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (r *Router) TRACE(path string, m ...echo.MiddlewareFunc) {
	r.Add(http.MethodTrace, path, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (r *Router) Any(path string, middleware ...echo.MiddlewareFunc) {
	for _, m := range methods {
		r.Add(m, path, middleware...)
	}
}

// Add registers a new route for method and path with matching handler.
func (r *Router) Add(method, path string, middleware ...echo.MiddlewareFunc) {
	m := echo.ComposeMiddleware(middleware...)

	// Validate path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, lcpIndex := 0, len(path); i < lcpIndex; i++ {
		if path[i] == ':' {
			if i > 0 && path[i-1] == '\\' {
				path = path[:i-1] + path[i:]
				i--
				lcpIndex--
				continue
			}
			j := i + 1

			r.insert(method, path[:i], nil, staticKind, "", nil)
			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, lcpIndex = j, len(path)

			if i == lcpIndex {
				// path node is last fragment of route path. ie. `/users/:id`
				r.insert(method, path[:i], m, paramKind, ppath, pnames)
			} else {
				r.insert(method, path[:i], nil, paramKind, "", nil)
			}
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, staticKind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], m, anyKind, ppath, pnames)
		}
	}

	r.insert(method, path, m, staticKind, ppath, pnames)
}

// Routes return echo middlewareFunc
func (r *Router) Routes() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m := echo.ComposeMiddleware(r.middleware...)
			matchedMiddleware := r.find(c.Request().Method, GetPath(c.Request()), c)

			if matchedMiddleware == nil {
				return m(next)(c)
			}

			return m(matchedMiddleware(next))(c)
		}
	}
}

func (r *Router) insert(method, path string, m echo.MiddlewareFunc, t kind, ppath string, pnames []string) {
	// Adjust max param
	paramLen := len(pnames)
	if r.maxParam < paramLen {
		r.maxParam = paramLen
	}

	currentNode := r.tree // Current node as root
	if currentNode == nil {
		panic("echo: invalid method")
	}
	search := path

	for {
		searchLen := len(search)
		prefixLen := len(currentNode.prefix)
		lcpLen := 0

		// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
		max := prefixLen
		if searchLen < max {
			max = searchLen
		}
		for ; lcpLen < max && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			// At root node
			currentNode.label = search[0]
			currentNode.prefix = search
			if m != nil {
				currentNode.kind = t
				currentNode.addMiddleware(method, m)
				currentNode.ppath = ppath
				currentNode.pnames = pnames
			}
			currentNode.isLeaf = currentNode.staticChildren == nil && currentNode.paramChild == nil && currentNode.anyChild == nil
		} else if lcpLen < prefixLen {
			// Split node
			n := newNode(
				currentNode.kind,
				currentNode.prefix[lcpLen:],
				currentNode,
				currentNode.staticChildren,
				currentNode.methodHandler,
				currentNode.ppath,
				currentNode.pnames,
				currentNode.paramChild,
				currentNode.anyChild,
			)
			// Update parent path for all children to new node
			for _, child := range currentNode.staticChildren {
				child.parent = n
			}
			if currentNode.paramChild != nil {
				currentNode.paramChild.parent = n
			}
			if currentNode.anyChild != nil {
				currentNode.anyChild.parent = n
			}

			// Reset parent node
			currentNode.kind = staticKind
			currentNode.label = currentNode.prefix[0]
			currentNode.prefix = currentNode.prefix[:lcpLen]
			currentNode.staticChildren = nil
			currentNode.methodHandler = new(methodHandler)
			currentNode.ppath = ""
			currentNode.pnames = nil
			currentNode.paramChild = nil
			currentNode.anyChild = nil
			currentNode.isLeaf = false
			currentNode.isHandler = false

			// Only Static children could reach here
			currentNode.addStaticChild(n)

			if lcpLen == searchLen {
				// At parent node
				currentNode.kind = t
				currentNode.addMiddleware(method, m)
				currentNode.ppath = ppath
				currentNode.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[lcpLen:], currentNode, nil, new(methodHandler), ppath, pnames, nil, nil)
				n.addMiddleware(method, m)
				// Only Static children could reach here
				currentNode.addStaticChild(n)
			}
			currentNode.isLeaf = currentNode.staticChildren == nil && currentNode.paramChild == nil && currentNode.anyChild == nil
		} else if lcpLen < searchLen {
			search = search[lcpLen:]
			c := currentNode.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				currentNode = c
				continue
			}
			// Create child node
			n := newNode(t, search, currentNode, nil, new(methodHandler), ppath, pnames, nil, nil)
			n.addMiddleware(method, m)
			switch t {
			case staticKind:
				currentNode.addStaticChild(n)
			case paramKind:
				currentNode.paramChild = n
			case anyKind:
				currentNode.anyChild = n
			}
			currentNode.isLeaf = currentNode.staticChildren == nil && currentNode.paramChild == nil && currentNode.anyChild == nil
		} else {
			// Node already exists
			if m != nil {
				currentNode.addMiddleware(method, m)
				currentNode.ppath = ppath
				if len(currentNode.pnames) == 0 { // Issue #729
					currentNode.pnames = pnames
				}
			}
		}
		return
	}
}

func newNode(t kind, pre string, p *node, sc children, mh *methodHandler, ppath string, pnames []string, paramChildren, anyChildren *node) *node {
	return &node{
		kind:           t,
		label:          pre[0],
		prefix:         pre,
		parent:         p,
		staticChildren: sc,
		ppath:          ppath,
		pnames:         pnames,
		methodHandler:  mh,
		paramChild:     paramChildren,
		anyChild:       anyChildren,
		isLeaf:         sc == nil && paramChildren == nil && anyChildren == nil,
		isHandler:      mh.isHandler(),
	}
}

func (n *node) addStaticChild(c *node) {
	n.staticChildren = append(n.staticChildren, c)
}

func (n *node) findStaticChild(l byte) *node {
	for _, c := range n.staticChildren {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.staticChildren {
		if c.label == l {
			return c
		}
	}
	if l == paramLabel {
		return n.paramChild
	}
	if l == anyLabel {
		return n.anyChild
	}
	return nil
}

func (n *node) addMiddleware(method string, m echo.MiddlewareFunc) {
	switch method {
	case http.MethodConnect:
		n.methodHandler.connect = m
	case http.MethodDelete:
		n.methodHandler.delete = m
	case http.MethodGet:
		n.methodHandler.get = m
	case http.MethodHead:
		n.methodHandler.head = m
	case http.MethodOptions:
		n.methodHandler.options = m
	case http.MethodPatch:
		n.methodHandler.patch = m
	case http.MethodPost:
		n.methodHandler.post = m
	case echo.PROPFIND:
		n.methodHandler.propfind = m
	case http.MethodPut:
		n.methodHandler.put = m
	case http.MethodTrace:
		n.methodHandler.trace = m
	case echo.REPORT:
		n.methodHandler.report = m
	}

	n.methodHandler.updateAllowHeader()
	if m != nil {
		n.isHandler = true
	} else {
		n.isHandler = n.methodHandler.isHandler()
	}
}

func (n *node) findMiddleware(method string) echo.MiddlewareFunc {
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
	case echo.PROPFIND:
		return n.methodHandler.propfind
	case http.MethodPut:
		return n.methodHandler.put
	case http.MethodTrace:
		return n.methodHandler.trace
	case echo.REPORT:
		return n.methodHandler.report
	default:
		return nil
	}
}

func optionsMethodHandler(allowMethods string) func(c echo.Context) error {
	return func(c echo.Context) error {
		// Note: we are not handling most of the CORS headers here. CORS is handled by CORS middleware
		// 'OPTIONS' method RFC: https://httpwg.org/specs/rfc7231.html#OPTIONS
		// 'Allow' header RFC: https://datatracker.ietf.org/doc/html/rfc7231#section-7.4.1
		c.Response().Header().Add(echo.HeaderAllow, allowMethods)
		return c.NoContent(http.StatusNoContent)
	}
}

// find lookup a handler registered for method and path. It also parses URL for path
// parameters and load them into context.
//
// For performance:
//
// - Get context from `Echo#AcquireContext()`
// - Reset it `echo.Context#Reset()`
// - Return it `Echo#ReleaseContext()`.
func (r *Router) find(method, path string, c echo.Context) echo.MiddlewareFunc {
	c.SetPath(path)
	currentNode := r.tree // Current node as root

	var (
		previousBestMatchNode *node
		matchedHandler        echo.MiddlewareFunc
		// search stores the remaining path to check for match. By each iteration we move from start of path to end of the path
		// and search value gets shorter and shorter.
		search      = path
		searchIndex = 0
		paramIndex  int               // Param counter
		paramValues = c.ParamValues() // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	if len(paramValues) < r.maxParam {
		newPvalues := make([]string, r.maxParam)
		copy(newPvalues, paramValues)
		paramValues = newPvalues
	}

	// Backtracking is needed when a dead end (leaf node) is reached in the router tree.
	// To backtrack the current node will be changed to the parent node and the next kind for the
	// router logic will be returned based on fromKind or kind of the dead end node (static > param > any).
	// For example if there is no static node match we should check parent next sibling by kind (param).
	// Backtracking itself does not check if there is a next sibling, this is done by the router logic.
	backtrackToNextNodeKind := func(fromKind kind) (nextNodeKind kind, valid bool) {
		previous := currentNode
		currentNode = previous.parent
		valid = currentNode != nil

		// Next node type by priority
		if previous.kind == anyKind {
			nextNodeKind = staticKind
		} else {
			nextNodeKind = previous.kind + 1
		}

		if fromKind == staticKind {
			// when backtracking is done from static kind block we did not change search so nothing to restore
			return
		}

		// restore search to value it was before we move to current node we are backtracking from.
		if previous.kind == staticKind {
			searchIndex -= len(previous.prefix)
		} else {
			paramIndex--
			// for param/any node.prefix value is always `:` so we can not deduce searchIndex from that and must use pValue
			// for that index as it would also contain part of path we cut off before moving into node we are backtracking from
			searchIndex -= len(paramValues[paramIndex])
			paramValues[paramIndex] = ""
		}
		search = path[searchIndex:]
		return
	}

	// Router tree is implemented by longest common prefix array (LCP array) https://en.wikipedia.org/wiki/LCP_array
	// Tree search is implemented as for loop where one loop iteration is divided into 3 separate blocks
	// Each of these blocks checks specific kind of node (static/param/any). Order of blocks reflex their priority in routing.
	// Search order/priority is: static > param > any.
	//
	// Note: backtracking in tree is implemented by replacing/switching currentNode to previous node
	// and hoping to (goto statement) next block by priority to check if it is the match.
	for {
		prefixLen := 0 // Prefix length
		lcpLen := 0    // LCP (longest common prefix) length

		if currentNode.kind == staticKind {
			searchLen := len(search)
			prefixLen = len(currentNode.prefix)

			// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
			max := prefixLen
			if searchLen < max {
				max = searchLen
			}
			for ; lcpLen < max && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision path
			nk, ok := backtrackToNextNodeKind(staticKind)
			if !ok {
				return nil // No other possibilities on the decision path
			} else if nk == paramKind {
				goto Param
				// NOTE: this case (backtracking from static node to previous any node) can not happen by current any matching logic. Any node is end of search currently
				//} else if nk == anyKind {
				//	goto Any
			} else {
				// Not found (this should never be possible for static node we are looking currently)
				break
			}
		}

		// The full prefix has matched, remove the prefix from the remaining search
		search = search[lcpLen:]
		searchIndex = searchIndex + lcpLen

		// Finish routing if no remaining search and we are on a node with handler and matching method type
		if search == "" && currentNode.isHandler {
			// check if current node has handler registered for http method we are looking for. we store currentNode as
			// best matching in case we do no find no more routes matching this path+method
			if previousBestMatchNode == nil {
				previousBestMatchNode = currentNode
			}
			if h := currentNode.findMiddleware(method); h != nil {
				matchedHandler = h
				break
			}
		}

		// Static node
		if search != "" {
			if child := currentNode.findStaticChild(search[0]); child != nil {
				currentNode = child
				continue
			}
		}

	Param:
		// Param node
		if child := currentNode.paramChild; search != "" && child != nil {
			currentNode = child
			i := 0
			l := len(search)
			if currentNode.isLeaf {
				// when param node does not have any children then param node should act similarly to any node - consider all remaining search as match
				i = l
			} else {
				for ; i < l && search[i] != '/'; i++ {
				}
			}

			paramValues[paramIndex] = search[:i]
			paramIndex++
			search = search[i:]
			searchIndex = searchIndex + i
			continue
		}

	Any:
		// Any node
		if child := currentNode.anyChild; child != nil {
			// If any node is found, use remaining path for paramValues
			currentNode = child
			paramValues[len(currentNode.pnames)-1] = search
			// update indexes/search in case we need to backtrack when no handler match is found
			paramIndex++
			searchIndex += +len(search)
			search = ""

			// check if current node has handler registered for http method we are looking for. we store currentNode as
			// best matching in case we do no find no more routes matching this path+method
			if previousBestMatchNode == nil {
				previousBestMatchNode = currentNode
			}
			if h := currentNode.findMiddleware(method); h != nil {
				matchedHandler = h
				break
			}
		}

		// Let's backtrack to the first possible alternative node of the decision path
		nk, ok := backtrackToNextNodeKind(anyKind)
		if !ok {
			break // No other possibilities on the decision path
		} else if nk == paramKind {
			goto Param
		} else if nk == anyKind {
			goto Any
		} else {
			// Not found
			break
		}
	}

	if currentNode == nil && previousBestMatchNode == nil {
		return NotFoundMiddleware // nothing matched at all
	}

	if matchedHandler == nil {
		// use previous match as basis. although we have no matching handler we have path match.
		// so we can send http.StatusMethodNotAllowed (405) instead of http.StatusNotFound (404)
		currentNode = previousBestMatchNode
		matchedHandler = NotFoundMiddleware
		if currentNode.isHandler {
			c.Set(echo.ContextKeyHeaderAllow, currentNode.methodHandler.allowHeader)
			matchedHandler = MethodNotAllowedMiddleware
			if method == http.MethodOptions {
				matchedHandler = func(next echo.HandlerFunc) echo.HandlerFunc {
					return optionsMethodHandler(currentNode.methodHandler.allowHeader)
				}
			}
		}
	}

	c.SetParamValues(paramValues...)
	c.SetPath(currentNode.ppath)
	c.SetParamNames(currentNode.pnames...)

	return matchedHandler
}

// GetPath returns RawPath, if it's empty returns Path from URL
// Difference between RawPath and Path is:
//  * Path is where request path is stored. Value is stored in decoded form: /%47%6f%2f becomes /Go/.
//  * RawPath is an optional field which only gets set if the default encoding is different from Path.
func GetPath(r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	return path
}

func handlerName(h echo.HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

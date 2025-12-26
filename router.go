// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
)

// Router is interface for routing request contexts to registered routes.
//
// Contract between Echo/Context instance and the router:
//   - all routes must be added through methods on echo.Echo instance.
//     Reason: Echo instance uses RouteInfo.Params() length to allocate slice for paths parameters (see `Echo.contextPathParamAllocSize`).
//   - Router must populate Context during Router.Route call with:
//   - Context.InitializeRoute (IMPORTANT! to reduce allocations use same slice that c.PathValues() returns)
//   - Optionally can set additional information to Context with Context.Set
type Router interface {
	// Add registers Routable with the Router and returns registered RouteInfo.
	//
	// Router may change Route.Path value in returned RouteInfo.Path.
	// Router generates RouteInfo.Parameters values from Route.Path.
	// Router generates RouteInfo.Name value if it is not provided.
	Add(routable Route) (RouteInfo, error)

	// Remove removes route from the Router.
	//
	// Router may choose not to implement this method.
	Remove(method string, path string) error

	// Routes returns information about all registered routes
	Routes() Routes

	// Route searches Router for matching route and applies it to the given context. In case when no matching method
	// was not found (405) or no matching route exists for path (404), router will return its implementation of 405/404
	// handler function.
	//
	// Router must populate Context during Router.Route call with:
	// - Context.InitializeRoute() (IMPORTANT! to reduce allocations use same slice that c.PathValues() returns)
	// - optionally can set additional information to Context with Context.Set()
	Route(c *Context) HandlerFunc
}

const (
	// NotFoundRouteName is name of RouteInfo returned when router did not find matching route (404: not found).
	NotFoundRouteName = "echo_route_not_found_name"
	// MethodNotAllowedRouteName is name of RouteInfo returned when router did not find matching method for route  (405: method not allowed).
	MethodNotAllowedRouteName = "echo_route_method_not_allowed_name"
)

// Routes is collection of RouteInfo instances with various helper methods.
type Routes []RouteInfo

// DefaultRouter is the registry of all registered routes for an `Echo` instance for
// request matching and URL path parameter parsing.
// Note: DefaultRouter is not coroutine-safe. Do not Add/Remove routes after HTTP server has been started with Echo.
type DefaultRouter struct {
	tree                    *node
	notFoundHandler         HandlerFunc
	methodNotAllowedHandler HandlerFunc
	optionsMethodHandler    HandlerFunc
	routes                  Routes
	// maxPathParamsLength tracks highest count of PathValues for all routes.
	maxPathParamsLength int

	allowOverwritingRoute    bool
	unescapePathParamValues  bool
	useEscapedPathForRouting bool
}

// RouterConfig is configuration options for (default) router
type RouterConfig struct {
	NotFoundHandler           HandlerFunc
	MethodNotAllowedHandler   HandlerFunc
	OptionsMethodHandler      HandlerFunc
	AllowOverwritingRoute     bool
	UnescapePathParamValues   bool
	UseEscapedPathForMatching bool
}

// NewRouter returns a new Router instance.
func NewRouter(config RouterConfig) *DefaultRouter {
	r := &DefaultRouter{
		tree: &node{
			methods:   new(routeMethods),
			isLeaf:    true,
			isHandler: false,
		},
		routes: make(Routes, 0),

		allowOverwritingRoute:    config.AllowOverwritingRoute,
		unescapePathParamValues:  config.UnescapePathParamValues,
		useEscapedPathForRouting: config.UseEscapedPathForMatching,

		notFoundHandler:         notFoundHandler,
		methodNotAllowedHandler: methodNotAllowedHandler,
		optionsMethodHandler:    optionsMethodHandler,
	}
	if config.NotFoundHandler != nil {
		r.notFoundHandler = config.NotFoundHandler
	}
	if config.MethodNotAllowedHandler != nil {
		r.methodNotAllowedHandler = config.MethodNotAllowedHandler
	}
	if config.OptionsMethodHandler != nil {
		r.optionsMethodHandler = config.OptionsMethodHandler
	}
	return r
}

type children []*node

type node struct {
	parent         *node
	methods        *routeMethods
	paramChild     *node
	anyChild       *node
	prefix         string
	originalPath   string
	staticChildren children
	paramsCount    int
	kind           kind
	label          byte
	isLeaf         bool
	isHandler      bool
}

type kind uint8

const (
	staticKind kind = iota
	paramKind
	anyKind

	paramLabel = byte(':')
	anyLabel   = byte('*')
)

type routeMethod struct {
	*RouteInfo
	handler      HandlerFunc
	orgRouteInfo RouteInfo
}

type routeMethods struct {
	connect  *routeMethod
	delete   *routeMethod
	get      *routeMethod
	head     *routeMethod
	options  *routeMethod
	patch    *routeMethod
	post     *routeMethod
	propfind *routeMethod
	put      *routeMethod
	trace    *routeMethod
	report   *routeMethod
	any      *routeMethod
	anyOther map[string]*routeMethod

	// notFoundHandler is handler registered with RouteNotFound method and is executed for 404 cases
	notFoundHandler *routeMethod

	// allowHeader contains comma-separated list of Methods registered to this node path.
	// it is optimization for http.StatusMethodNotAllowed (405) handling.
	allowHeader string
}

func (m *routeMethods) set(method string, r *routeMethod) {
	switch method {
	case http.MethodConnect:
		m.connect = r
	case http.MethodDelete:
		m.delete = r
	case http.MethodGet:
		m.get = r
	case http.MethodHead:
		m.head = r
	case http.MethodOptions:
		m.options = r
	case http.MethodPatch:
		m.patch = r
	case http.MethodPost:
		m.post = r
	case PROPFIND:
		m.propfind = r
	case http.MethodPut:
		m.put = r
	case http.MethodTrace:
		m.trace = r
	case REPORT:
		m.report = r
	case RouteAny:
		m.any = r
	case RouteNotFound:
		m.notFoundHandler = r
		return // RouteNotFound/404 is not considered as a handler so no further logic needs to be executed
	default:
		if m.anyOther == nil {
			m.anyOther = make(map[string]*routeMethod)
		}
		if r.handler == nil {
			delete(m.anyOther, method)
		} else {
			m.anyOther[method] = r
		}
	}
	m.updateAllowHeader()
}

func (m *routeMethods) find(method string, fallbackToAny bool) *routeMethod {
	var r *routeMethod
	switch method {
	case http.MethodConnect:
		r = m.connect
	case http.MethodDelete:
		r = m.delete
	case http.MethodGet:
		r = m.get
	case http.MethodHead:
		r = m.head
	case http.MethodOptions:
		r = m.options
	case http.MethodPatch:
		r = m.patch
	case http.MethodPost:
		r = m.post
	case PROPFIND:
		r = m.propfind
	case http.MethodPut:
		r = m.put
	case http.MethodTrace:
		r = m.trace
	case REPORT:
		r = m.report
	case RouteAny:
		r = m.any
	case RouteNotFound:
		r = m.notFoundHandler
	default:
		r = m.anyOther[method]
	}
	if r != nil || !fallbackToAny {
		return r
	}
	return m.any
}

func (m *routeMethods) updateAllowHeader() {
	buf := new(bytes.Buffer)
	buf.WriteString(http.MethodOptions)
	hasAnyMethod := m.any != nil

	if hasAnyMethod || m.connect != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodConnect)
	}
	if hasAnyMethod || m.delete != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodDelete)
	}
	if hasAnyMethod || m.get != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodGet)
	}
	if hasAnyMethod || m.head != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodHead)
	}
	if hasAnyMethod || m.patch != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPatch)
	}
	if hasAnyMethod || m.post != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPost)
	}
	if hasAnyMethod || m.propfind != nil {
		buf.WriteString(", PROPFIND")
	}
	if hasAnyMethod || m.put != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodPut)
	}
	if hasAnyMethod || m.trace != nil {
		buf.WriteString(", ")
		buf.WriteString(http.MethodTrace)
	}
	if hasAnyMethod || m.report != nil {
		buf.WriteString(", REPORT")
	}
	for method := range m.anyOther { // for simplicity, we use map and therefore order is not deterministic here
		buf.WriteString(", ")
		buf.WriteString(method)
	}
	m.allowHeader = buf.String()
}

func (m *routeMethods) isHandler() bool {
	return m.get != nil ||
		m.post != nil ||
		m.options != nil ||
		m.put != nil ||
		m.delete != nil ||
		m.connect != nil ||
		m.head != nil ||
		m.patch != nil ||
		m.propfind != nil ||
		m.trace != nil ||
		m.report != nil ||
		m.any != nil ||
		len(m.anyOther) != 0
	// RouteNotFound/404 is not considered as a handler
}

// Routes returns all registered routes
func (r *DefaultRouter) Routes() Routes {
	return r.routes
}

// Remove unregisters registered route
func (r *DefaultRouter) Remove(method string, path string) error {
	currentNode := r.tree
	if currentNode == nil || (currentNode.isLeaf && !currentNode.isHandler) {
		return errors.New("router has no routes to remove")
	}

	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	var nodeToRemove *node
	prefixLen := 0
	for {
		if currentNode.originalPath == path && currentNode.isHandler {
			nodeToRemove = currentNode
			break
		}
		if currentNode.kind == staticKind {
			prefixLen = prefixLen + len(currentNode.prefix)
		} else {
			prefixLen = len(currentNode.originalPath)
		}

		if prefixLen >= len(path) {
			break
		}

		next := path[prefixLen]
		switch next {
		case paramLabel:
			currentNode = currentNode.paramChild
		case anyLabel:
			currentNode = currentNode.anyChild
		default:
			currentNode = currentNode.findStaticChild(next)
		}

		if currentNode == nil {
			break
		}
	}

	if nodeToRemove == nil {
		return errors.New("could not find route to remove by given path")
	}

	if !nodeToRemove.isHandler {
		return errors.New("could not find route to remove by given path")
	}

	if mh := nodeToRemove.methods.find(method, false); mh == nil {
		return errors.New("could not find route to remove by given path and method")
	}
	nodeToRemove.setHandler(method, nil)

	var rIndex int
	for i, rr := range r.routes {
		if rr.Method == method && rr.Path == path {
			rIndex = i
			break
		}
	}
	r.routes = append(r.routes[:rIndex], r.routes[rIndex+1:]...)

	if !nodeToRemove.isHandler && nodeToRemove.isLeaf {
		// TODO: if !nodeToRemove.isLeaf and has at least 2 children merge paths for remaining nodes?
		current := nodeToRemove
		for {
			parent := current.parent
			if parent == nil {
				break
			}
			switch current.kind {
			case staticKind:
				var index int
				for i, c := range parent.staticChildren {
					if c == current {
						index = i
						break
					}
				}
				parent.staticChildren = append(parent.staticChildren[:index], parent.staticChildren[index+1:]...)
			case paramKind:
				parent.paramChild = nil
			case anyKind:
				parent.anyChild = nil
			}

			parent.isLeaf = parent.anyChild == nil && parent.paramChild == nil && len(parent.staticChildren) == 0
			if !parent.isLeaf || parent.isHandler {
				break
			}
			current = parent
		}
	}

	return nil
}

// AddRouteError is error returned by Router.Add containing information what actual route adding failed. Useful for
// mass adding (i.e. Any() routes)
type AddRouteError struct {
	Err    error
	Method string
	Path   string
}

func (e *AddRouteError) Error() string { return e.Method + " " + e.Path + ": " + e.Err.Error() }

func (e *AddRouteError) Unwrap() error { return e.Err }

func newAddRouteError(route Route, err error) *AddRouteError {
	return &AddRouteError{
		Method: route.Method,
		Path:   route.Path,
		Err:    err,
	}
}

// Add registers a new route for method and path with matching handler.
func (r *DefaultRouter) Add(route Route) (RouteInfo, error) {
	if route.Handler == nil {
		return RouteInfo{}, newAddRouteError(route, errors.New("adding route without handler function"))
	}
	method := route.Method
	path := normalizePathSlash(route.Path)

	h := applyMiddleware(route.Handler, route.Middlewares...)
	if !r.allowOverwritingRoute {
		for _, rr := range r.routes {
			if route.Method == rr.Method && route.Path == rr.Path {
				return RouteInfo{}, newAddRouteError(route, errors.New("adding duplicate route (same method+path) is not allowed"))
			}
		}
	}

	paramNames := make([]string, 0)
	originalPath := path
	wasAdded := false
	var ri RouteInfo
	for i, lcpIndex := 0, len(path); i < lcpIndex; i++ {
		if path[i] == paramLabel {
			if i > 0 && path[i-1] == '\\' {
				path = path[:i-1] + path[i:]
				i--
				lcpIndex--
				continue
			}
			j := i + 1

			r.insert(staticKind, path[:i], method, routeMethod{RouteInfo: &RouteInfo{Method: method}})
			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			paramNames = append(paramNames, path[j:i])
			path = path[:j] + path[i:]
			i, lcpIndex = j, len(path)

			if i == lcpIndex {
				// path node is last fragment of route path. ie. `/users/:id`
				ri = route.ToRouteInfo(paramNames)
				rm := routeMethod{
					RouteInfo:    &RouteInfo{Method: method, Path: originalPath, Parameters: paramNames, Name: route.Name},
					handler:      h,
					orgRouteInfo: ri,
				}
				r.insert(paramKind, path[:i], method, rm)
				wasAdded = true
				break
			} else {
				r.insert(paramKind, path[:i], method, routeMethod{RouteInfo: &RouteInfo{Method: method}})
			}
		} else if path[i] == anyLabel {
			r.insert(staticKind, path[:i], method, routeMethod{RouteInfo: &RouteInfo{Method: method}})
			paramNames = append(paramNames, "*")
			ri = route.ToRouteInfo(paramNames)
			rm := routeMethod{
				RouteInfo:    &RouteInfo{Method: method, Path: originalPath, Parameters: paramNames, Name: route.Name},
				handler:      h,
				orgRouteInfo: ri,
			}
			r.insert(anyKind, path[:i+1], method, rm)
			wasAdded = true
			break
		}
	}

	if !wasAdded {
		ri = route.ToRouteInfo(paramNames)
		rm := routeMethod{
			RouteInfo:    &RouteInfo{Method: method, Path: originalPath, Parameters: paramNames, Name: route.Name},
			handler:      h,
			orgRouteInfo: ri,
		}
		r.insert(staticKind, path, method, rm)
	}

	r.storeRouteInfo(ri)

	return ri, nil
}

func normalizePathSlash(path string) string {
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}
	return path
}

func (r *DefaultRouter) storeRouteInfo(ri RouteInfo) {
	for i, rr := range r.routes {
		if ri.Method == rr.Method && ri.Path == rr.Path {
			r.routes[i] = ri
			return
		}
	}
	r.routes = append(r.routes, ri)
}

func (r *DefaultRouter) insert(t kind, path string, method string, ri routeMethod) {
	if len(ri.Parameters) > r.maxPathParamsLength {
		r.maxPathParamsLength = len(ri.Parameters)
	}
	currentNode := r.tree // Current node as root
	search := path

	for {
		searchLen := len(search)
		prefixLen := len(currentNode.prefix)
		lcpLen := 0

		// LCP - Longest Common Prefix (https://en.wikipedia.org/wiki/LCP_array)
		maxL := prefixLen
		if searchLen < maxL {
			maxL = searchLen
		}
		for ; lcpLen < maxL && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
		}

		if lcpLen == 0 {
			// At root node
			currentNode.label = search[0]
			currentNode.prefix = search
			if ri.handler != nil {
				currentNode.kind = t
				currentNode.setHandler(method, &ri)
				currentNode.paramsCount = len(ri.Parameters)
				currentNode.originalPath = ri.Path
			}
			currentNode.isLeaf = currentNode.staticChildren == nil && currentNode.paramChild == nil && currentNode.anyChild == nil
		} else if lcpLen < prefixLen {
			// Split node into two before we insert new node.
			// This happens when we are inserting path that is submatch of any existing inserted paths.
			// For example, we have node `/test` and now are about to insert `/te/*`. In that case
			// 1. overlapping part is `/te` that is used as parent node
			// 2. `st` is part from existing node that is not matching - it gets its own node (child to `/te`)
			// 3. `/*` is the new part we are about to insert (child to `/te`)
			n := newNode(
				currentNode.kind,
				currentNode.prefix[lcpLen:],
				currentNode,
				currentNode.staticChildren,
				currentNode.methods,
				currentNode.paramsCount,
				currentNode.originalPath,
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
			currentNode.methods = new(routeMethods)
			currentNode.originalPath = ""
			currentNode.paramsCount = 0
			currentNode.paramChild = nil
			currentNode.anyChild = nil
			currentNode.isLeaf = false
			currentNode.isHandler = false

			// Only Static children could reach here
			currentNode.addStaticChild(n)

			if lcpLen == searchLen {
				// At parent node
				currentNode.kind = t
				if ri.handler != nil {
					currentNode.setHandler(method, &ri)
					currentNode.paramsCount = len(ri.Parameters)
					currentNode.originalPath = ri.Path
				}
			} else {
				// Create child node
				n = newNode(t, search[lcpLen:], currentNode, nil, new(routeMethods), 0, ri.Path, nil, nil)
				if ri.handler != nil {
					n.setHandler(method, &ri)
					n.paramsCount = len(ri.Parameters)
				}
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
			n := newNode(t, search, currentNode, nil, new(routeMethods), 0, ri.Path, nil, nil)
			if ri.handler != nil {
				n.setHandler(method, &ri)
				n.paramsCount = len(ri.Parameters)
			}
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
			if ri.handler != nil {
				currentNode.setHandler(method, &ri)
				currentNode.paramsCount = len(ri.Parameters)
				currentNode.originalPath = ri.Path
			}
		}
		return
	}
}

func newNode(
	t kind,
	pre string,
	p *node,
	sc children,
	mh *routeMethods,
	paramsCount int,
	ppath string,
	paramChildren,
	anyChildren *node,
) *node {
	return &node{
		kind:           t,
		label:          pre[0],
		prefix:         pre,
		parent:         p,
		staticChildren: sc,
		originalPath:   ppath,
		paramsCount:    paramsCount,
		methods:        mh,
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
	if c := n.findStaticChild(l); c != nil {
		return c
	}
	if l == paramLabel {
		return n.paramChild
	}
	if l == anyLabel {
		return n.anyChild
	}
	return nil
}

func (n *node) setHandler(method string, r *routeMethod) {
	n.methods.set(method, r)
	n.isHandler = n.methods.isHandler()
}

// Note: notFoundRouteInfo exists to avoid allocations when setting 404 RouteInfo to Context
var notFoundRouteInfo = &RouteInfo{
	Method:     "",
	Path:       "",
	Parameters: nil,
	Name:       NotFoundRouteName,
}

// Note: methodNotAllowedRouteInfo exists to avoid allocations when setting 405 RouteInfo to Context
var methodNotAllowedRouteInfo = &RouteInfo{
	Method:     "",
	Path:       "",
	Parameters: nil,
	Name:       MethodNotAllowedRouteName,
}

// notFoundHandler is handler for 404 cases
// Handle returned ErrNotFound errors in Echo.HTTPErrorHandler
var notFoundHandler = func(c *Context) error {
	return ErrNotFound
}

// methodNotAllowedHandler is handler for case when route for path+method match was not found (http code 405)
// Handle returned ErrMethodNotAllowed errors in Echo.HTTPErrorHandler
var methodNotAllowedHandler = func(c *Context) error {
	// See RFC 7231 section 7.4.1: An origin server MUST generate an Allow field in a 405 (Method Not Allowed)
	// response and MAY do so in any other response. For disabled resources an empty Allow header may be returned
	routerAllowMethods, ok := c.Get(ContextKeyHeaderAllow).(string)
	if ok && routerAllowMethods != "" {
		c.Response().Header().Set(HeaderAllow, routerAllowMethods)
	}
	return ErrMethodNotAllowed
}

// optionsMethodHandler is default handler for OPTIONS method.
// Use `middleware.CORS` if you need support for preflighted requests in CORS
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/OPTIONS
var optionsMethodHandler = func(c *Context) error {
	// See RFC 7231 section 7.4.1: An origin server MUST generate an Allow field in a 405 (Method Not Allowed)
	// response and MAY do so in any other response. For disabled resources an empty Allow header may be returned
	routerAllowMethods, ok := c.Get(ContextKeyHeaderAllow).(string)
	if ok && routerAllowMethods != "" {
		c.Response().Header().Set(HeaderAllow, routerAllowMethods)
	}
	return c.NoContent(http.StatusNoContent)
}

// Route looks up a handler registered for method and path. It also parses URL for path parameters and loads them
// into context.
//
// For performance:
//
// - Get context from `Echo#AcquireContext()`
// - Reset it `Context#Reset()`
// - Return it `Echo#ReleaseContext()`.
func (r *DefaultRouter) Route(c *Context) HandlerFunc {
	pathValues := c.PathValues()
	if cap(pathValues) < r.maxPathParamsLength {
		pathValues = make(PathValues, 0, r.maxPathParamsLength)
	} else {
		pathValues = pathValues[0:cap(pathValues)] // resize slice to maximum capacity so we can index set values
	}

	req := c.Request()
	path := req.URL.Path
	if !r.useEscapedPathForRouting && req.URL.RawPath != "" {
		// Difference between URL.RawPath and URL.Path is:
		//  * URL.Path is where request path is stored. Value is stored in decoded form: /%47%6f%2f becomes /Go/.
		//  * URL.RawPath is an optional field which only gets set if the default encoding is different from Path.
		path = req.URL.RawPath
	}
	var (
		currentNode           = r.tree // root as current node
		previousBestMatchNode *node
		matchedRouteMethod    *routeMethod
		// search stores the remaining path to check for match. By each iteration we move from start of path to end of the path
		// and search value gets shorter and shorter.
		search      = path
		searchIndex = 0
		paramIndex  int // Param counter
	)

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
			searchIndex -= len(pathValues[paramIndex].Value)
			pathValues[paramIndex].Value = ""
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
			lMax := prefixLen
			if searchLen < lMax {
				lMax = searchLen
			}
			for ; lcpLen < lMax && search[lcpLen] == currentNode.prefix[lcpLen]; lcpLen++ {
			}
		}

		if lcpLen != prefixLen {
			// No matching prefix, let's backtrack to the first possible alternative node of the decision path
			nk, ok := backtrackToNextNodeKind(staticKind)
			if !ok {
				break // No other possibilities on the decision path, handler will be whatever context is reset to.
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

		// Finish routing if is no request path remaining to search
		if search == "" {
			// in case of node that is handler we have exact method type match or something for 405 to use
			if currentNode.isHandler {
				// check if current node has handler registered for http method we are looking for. we store currentNode as
				// best matching in case we do no find no more routes matching this path+method
				if previousBestMatchNode == nil {
					previousBestMatchNode = currentNode
				}
				if h := currentNode.methods.find(req.Method, true); h != nil {
					matchedRouteMethod = h
					break
				}
			} else if currentNode.methods.notFoundHandler != nil {
				matchedRouteMethod = currentNode.methods.notFoundHandler
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
				// when param node does not have any children (path param is last piece of route path) then param node should
				// act similarly to any node - consider all remaining search as match
				i = l
			} else {
				for ; i < l && search[i] != '/'; i++ {
				}
			}

			pathValues[paramIndex].Value = search[:i]
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
			pathValues[currentNode.paramsCount-1].Value = search
			// update indexes/search in case we need to backtrack when no handler match is found
			paramIndex++
			searchIndex += len(search)
			search = ""

			if rMethod := currentNode.methods.find(req.Method, true); rMethod != nil {
				matchedRouteMethod = rMethod
				break
			}
			// we store currentNode as best matching in case we do not find more routes matching this path+method. Needed for 405
			if previousBestMatchNode == nil {
				previousBestMatchNode = currentNode
			}
			if currentNode.methods.notFoundHandler != nil {
				matchedRouteMethod = currentNode.methods.notFoundHandler
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
		pathValues = pathValues[0:0]

		c.InitializeRoute(notFoundRouteInfo, &pathValues)
		return r.notFoundHandler // nothing matched at all with given path
	}

	var rHandler HandlerFunc
	var rPath string
	var rInfo *RouteInfo
	if matchedRouteMethod != nil {
		rHandler = matchedRouteMethod.handler
		rPath = matchedRouteMethod.RouteInfo.Path
		rInfo = matchedRouteMethod.RouteInfo
	} else {
		// use previous match as basis. although we have no matching handler we have path match.
		// so we can send http.StatusMethodNotAllowed (405) instead of http.StatusNotFound (404)
		currentNode = previousBestMatchNode

		rPath = currentNode.originalPath
		rInfo = notFoundRouteInfo
		if currentNode.methods.notFoundHandler != nil {
			matchedRouteMethod = currentNode.methods.notFoundHandler

			rInfo = matchedRouteMethod.RouteInfo
			rPath = matchedRouteMethod.Path
			rHandler = matchedRouteMethod.handler
		} else if currentNode.isHandler {
			rInfo = methodNotAllowedRouteInfo

			c.Set(ContextKeyHeaderAllow, currentNode.methods.allowHeader)
			rHandler = r.methodNotAllowedHandler
			if req.Method == http.MethodOptions {
				rHandler = r.optionsMethodHandler
			}
		}
	}

	pathValues = pathValues[0:currentNode.paramsCount]
	if matchedRouteMethod != nil {
		for i, name := range matchedRouteMethod.Parameters {
			pathValues[i].Name = name
		}
	}

	if r.unescapePathParamValues {
		// See issue #1531, #1258 - there are cases when path parameter need to be unescaped
		for i, p := range pathValues {
			tmpVal, err := url.PathUnescape(p.Value)
			if err == nil { // handle problems by ignoring them.
				pathValues[i].Value = tmpVal
			}
		}
	}

	c.InitializeRoute(rInfo, &pathValues)
	c.SetPath(rPath) // after InitializeRoute so we would not accidentally change `notFoundRouteInfo` or `methodNotAllowedRouteInfo` Path

	return rHandler
}

// PathValues is collections of PathValue instances with various helper methods
type PathValues []PathValue

// PathValue is tuple pf path parameter name and its value in request path
type PathValue struct {
	Name  string
	Value string
}

// Get returns path parameter value for given name or false.
func (p PathValues) Get(name string) (string, bool) {
	for _, param := range p {
		if param.Name == name {
			return param.Value, true
		}
	}
	return "", false
}

// GetOr returns path parameter value for given name or default value if the name does not exist.
func (p PathValues) GetOr(name string, defaultValue string) string {
	for _, param := range p {
		if param.Name == name {
			return param.Value
		}
	}
	return defaultValue
}

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/url"
	"strings"
)

// Route looks up a handler and path parameters. It uses the established fast
// path unless the router has an inline verb and the request contains a colon.
func (r *DefaultRouter) Route(c *Context) HandlerFunc {
	req := c.Request()
	path := req.URL.Path
	if !r.useEscapedPathForRouting && req.URL.RawPath != "" {
		// Difference between URL.RawPath and URL.Path is:
		//  * URL.Path is where request path is stored. Value is stored in decoded form: /%47%6f%2f becomes /Go/.
		//  * URL.RawPath is an optional field which only gets set if the default encoding is different from Path.
		path = req.URL.RawPath
	}
	if r.hasInlineVerb && strings.IndexByte(path, ':') >= 0 {
		return r.routeInline(c)
	}
	pathValues := c.PathValues()
	if cap(pathValues) < r.maxPathParamsLength {
		pathValues = make(PathValues, 0, r.maxPathParamsLength)
	} else {
		pathValues = pathValues[0:cap(pathValues)] // resize slice to maximum capacity so we can index set values
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
			// param/any node prefixes are a single marker byte, so restore searchIndex
			// from the value stored for that param instead
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
			lMax := min(searchLen, prefixLen)
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
				if h := currentNode.methods.find(req.Method, true, r.autoHandleHEAD); h != nil {
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

			if rMethod := currentNode.methods.find(req.Method, true, r.autoHandleHEAD); rMethod != nil {
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
		if req.Method == http.MethodHead && matchedRouteMethod.wrappedHeadHandler != nil {
			rHandler = matchedRouteMethod.wrappedHeadHandler
			// we are not touching rInfo.Method and let it be value from GET routeInfo
		}

		rPath = matchedRouteMethod.Path
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
	c.SetPath(rPath)          // after InitializeRoute so we would not accidentally change `notFoundRouteInfo` or `methodNotAllowedRouteInfo` Path
	c.request.Pattern = rPath // help standard library based middlewares. This is a deliberate choice not to call `request.SetPathValue` for params.
	return rHandler
}

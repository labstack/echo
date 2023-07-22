package echo

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// Route contains information to adding/registering new route with the router.
// Method+Path pair uniquely identifies the Route. It is mandatory to provide Method+Path+Handler fields.
type Route struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Middlewares []MiddlewareFunc

	Name string
}

// ToRouteInfo converts Route to RouteInfo
func (r Route) ToRouteInfo(params []string) RouteInfo {
	name := r.Name
	if name == "" {
		name = r.Method + ":" + r.Path
	}

	return routeInfo{
		method: r.Method,
		path:   r.Path,
		params: append([]string(nil), params...),
		name:   name,
	}
}

// ToRoute returns Route which Router uses to register the method handler for path.
func (r Route) ToRoute() Route {
	return r
}

// ForGroup recreates Route with added group prefix and group middlewares it is grouped to.
func (r Route) ForGroup(pathPrefix string, middlewares []MiddlewareFunc) Routable {
	r.Path = pathPrefix + r.Path

	if len(middlewares) > 0 {
		m := make([]MiddlewareFunc, 0, len(middlewares)+len(r.Middlewares))
		m = append(m, middlewares...)
		m = append(m, r.Middlewares...)
		r.Middlewares = m
	}
	return r
}

type routeInfo struct {
	method string
	path   string
	params []string
	name   string
}

func (r routeInfo) Method() string {
	return r.method
}

func (r routeInfo) Path() string {
	return r.path
}

func (r routeInfo) Params() []string {
	return append([]string(nil), r.params...)
}

func (r routeInfo) Name() string {
	return r.name
}

// Reverse reverses route to URL string by replacing path parameters with given params values.
func (r routeInfo) Reverse(params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for i, l := 0, len(r.path); i < l; i++ {
		hasBackslash := r.path[i] == '\\'
		if hasBackslash && i+1 < l && r.path[i+1] == ':' {
			i++ // backslash before colon escapes that colon. in that case skip backslash
		}
		if n < ln && (r.path[i] == anyLabel || (!hasBackslash && r.path[i] == paramLabel)) {
			// in case of `*` wildcard or `:` (unescaped colon) param we replace everything till next slash or end of path
			for ; i < l && r.path[i] != '/'; i++ {
			}
			uri.WriteString(fmt.Sprintf("%v", params[n]))
			n++
		}
		if i < l {
			uri.WriteByte(r.path[i])
		}
	}
	return uri.String()
}

// HandlerName returns string name for given function.
func HandlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// Reverse reverses route to URL string by replacing path parameters with given params values.
func (r Routes) Reverse(name string, params ...interface{}) (string, error) {
	for _, rr := range r {
		if rr.Name() == name {
			return rr.Reverse(params...), nil
		}
	}
	return "", errors.New("route not found")
}

// FindByMethodPath searched for matching route info by method and path
func (r Routes) FindByMethodPath(method string, path string) (RouteInfo, error) {
	if r == nil {
		return nil, errors.New("route not found by method and path")
	}

	for _, rr := range r {
		if rr.Method() == method && rr.Path() == path {
			return rr, nil
		}
	}
	return nil, errors.New("route not found by method and path")
}

// FilterByMethod searched for matching route info by method
func (r Routes) FilterByMethod(method string) (Routes, error) {
	if r == nil {
		return nil, errors.New("route not found by method")
	}

	result := make(Routes, 0)
	for _, rr := range r {
		if rr.Method() == method {
			result = append(result, rr)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("route not found by method")
	}
	return result, nil
}

// FilterByPath searched for matching route info by path
func (r Routes) FilterByPath(path string) (Routes, error) {
	if r == nil {
		return nil, errors.New("route not found by path")
	}

	result := make(Routes, 0)
	for _, rr := range r {
		if rr.Path() == path {
			result = append(result, rr)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("route not found by path")
	}
	return result, nil
}

// FilterByName searched for matching route info by name
func (r Routes) FilterByName(name string) (Routes, error) {
	if r == nil {
		return nil, errors.New("route not found by name")
	}

	result := make(Routes, 0)
	for _, rr := range r {
		if rr.Name() == name {
			result = append(result, rr)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("route not found by name")
	}
	return result, nil
}

package echo

import (
	"sync"
)

// NewConcurrentRouter creates concurrency safe Router which routes can be added/removed safely
// even after http.Server has been started.
func NewConcurrentRouter(r Router) Router {
	return &concurrentRouter{
		mu:     sync.RWMutex{},
		router: r,
	}
}

type concurrentRouter struct {
	mu     sync.RWMutex
	router Router
}

func (r *concurrentRouter) Route(c *Context) HandlerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.router.Route(c)
}

func (r *concurrentRouter) Routes() Routes {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.router.Routes().Clone()
}

func (r *concurrentRouter) Add(routable Route) (RouteInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.router.Add(routable)
}

func (r *concurrentRouter) Remove(method string, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.router.Remove(method, path)
}

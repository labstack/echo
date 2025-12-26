// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentRouter_Remove(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	_, err := router.Add(Route{
		Method:  http.MethodGet,
		Path:    "/initial1",
		Handler: handlerFunc,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(router.Routes()), 1)

	err = router.Remove(http.MethodGet, "/initial1")
	assert.NoError(t, err)
	assert.Equal(t, len(router.Routes()), 0)
}

func TestConcurrentRouter_ConcurrentReads(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	testPaths := []string{"/route1", "/route2", "/route3", "/route4", "/route5"}
	for _, path := range testPaths {
		_, err := router.Add(Route{
			Method:  http.MethodGet,
			Path:    path,
			Handler: handlerFunc,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Launch 10 goroutines for concurrent reads
	var wg sync.WaitGroup
	var routeCallCount atomic.Int64
	var routesCallCount atomic.Int64

	numGoroutines := 10
	routeCallsPerGoroutine := 50
	routesCallsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Call Route() 50 times
			for j := 0; j < routeCallsPerGoroutine; j++ {
				path := testPaths[j%len(testPaths)]
				req := httptest.NewRequest(http.MethodGet, path, nil)
				rec := httptest.NewRecorder()
				c := newContext(req, rec, nil)

				handler := router.Route(c)
				if handler != nil {
					routeCallCount.Add(1)
				}
			}

			// Call Routes() 20 times
			for j := 0; j < routesCallsPerGoroutine; j++ {
				routes := router.Routes()
				if len(routes) == 5 {
					routesCallCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all operations succeeded
	expectedRouteCalls := int64(numGoroutines * routeCallsPerGoroutine)
	expectedRoutesCalls := int64(numGoroutines * routesCallsPerGoroutine)

	assert.Equal(t, expectedRouteCalls, routeCallCount.Load(), "all Route() calls should succeed")
	assert.Equal(t, expectedRoutesCalls, routesCallCount.Load(), "all Routes() calls should succeed")
}

func TestConcurrentRouter_ConcurrentWrites(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	_, _ = router.Add(Route{Method: http.MethodGet, Path: "/initial1", Handler: handlerFunc})
	_, _ = router.Add(Route{Method: http.MethodGet, Path: "/initial2", Handler: handlerFunc})

	// Launch 5 goroutines, each adds 10 unique routes
	var wg sync.WaitGroup
	var addCount atomic.Int64

	numGoroutines := 5
	addsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < addsPerGoroutine; j++ {
				path := fmt.Sprintf("/route-g%d-n%d", goroutineID, j)
				_, err := router.Add(Route{
					Method:  http.MethodGet,
					Path:    path,
					Handler: handlerFunc,
				})
				if err == nil {
					addCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final route count
	expectedAdds := int64(numGoroutines * addsPerGoroutine)
	assert.Equal(t, expectedAdds, addCount.Load(), "all Add() calls should succeed")

	expectedTotal := 2 + int(expectedAdds) // 2 initial + 50 added
	assert.Len(t, router.Routes(), expectedTotal, "route count mismatch")

	// Verify all routes are accessible
	allRoutes := router.Routes()
	assert.Len(t, allRoutes, expectedTotal)
}

func TestConcurrentRouter_ConcurrentReadWrite(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	initialPaths := []string{"/read1", "/read2", "/read3"}
	for _, path := range initialPaths {
		_, err := router.Add(Route{Method: http.MethodGet, Path: path, Handler: handlerFunc})
		if err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	var routeCallCount atomic.Int64
	var addCallCount atomic.Int64
	var routesCallCount atomic.Int64

	// Launch 4 reader goroutines: call Route() 100 times each
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				path := initialPaths[j%len(initialPaths)]

				req := httptest.NewRequest(http.MethodGet, path, nil)
				rec := httptest.NewRecorder()
				c := newContext(req, rec, nil)

				handler := router.Route(c)
				if handler != nil {
					routeCallCount.Add(1)
				}
			}
		}()
	}

	// Launch 2 writer goroutines: call Add() 20 times each
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				path := fmt.Sprintf("/write-g%d-n%d", goroutineID, j)
				_, err := router.Add(Route{
					Method:  http.MethodGet,
					Path:    path,
					Handler: handlerFunc,
				})
				if err == nil {
					addCallCount.Add(1)
				}
			}
		}(i)
	}

	// Launch 2 inspector goroutines: call Routes() 50 times each
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				routes := router.Routes()
				if routes != nil {
					routesCallCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	// Verify all operations succeeded
	assert.Equal(t, int64(400), routeCallCount.Load(), "all Route() calls should succeed")
	assert.Equal(t, int64(40), addCallCount.Load(), "all Add() calls should succeed")
	assert.Equal(t, int64(100), routesCallCount.Load(), "all Routes() calls should succeed")

	// Verify final route count
	expectedTotal := 3 + 40 // 3 initial + 40 added
	assert.Len(t, router.Routes(), expectedTotal, "route count mismatch")
}

// TestConcurrentRouter_RoutesIterationDuringModification verifies that iterating over
// Routes() while Add/Remove operations are happening doesn't cause data races.
// This test specifically validates that Routes() returns a copy, not a reference.
func TestConcurrentRouter_RoutesIterationDuringModification(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	// Add initial routes
	for i := 0; i < 10; i++ {
		_, err := router.Add(Route{
			Method:  http.MethodGet,
			Path:    fmt.Sprintf("/initial-%d", i),
			Handler: handlerFunc,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	var iterationCount atomic.Int64
	var addRemoveCount atomic.Int64

	// Launch 3 goroutines that iterate over Routes() and access each element
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				routes := router.Routes()
				// Actually iterate and access the route data
				// This would cause a data race if Routes() returned a direct reference
				for _, route := range routes {
					_ = route.Method // Read the method
					_ = route.Path   // Read the path
					_ = route.Name   // Read the name
					if len(route.Parameters) > 0 {
						_ = route.Parameters[0] // Read parameters if present
					}
				}
				iterationCount.Add(1)
			}
		}(i)
	}

	// Launch 2 goroutines that continuously Add routes
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				path := fmt.Sprintf("/add-g%d-n%d", goroutineID, j)
				_, err := router.Add(Route{
					Method:  http.MethodPost,
					Path:    path,
					Handler: handlerFunc,
				})
				if err == nil {
					addRemoveCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify operations completed
	assert.Equal(t, int64(300), iterationCount.Load(), "all iterations should complete")
	assert.Equal(t, int64(60), addRemoveCount.Load(), "all add operations should succeed")

	// Verify final state
	finalRoutes := router.Routes()
	assert.Len(t, finalRoutes, 70, "should have 10 initial + 60 added routes")
}

// TestConcurrentRouter_ParametersNoRace verifies that accessing RouteInfo.Parameters
// while routes are being added concurrently doesn't cause data races.
// This test validates that Routes() deep-copies RouteInfo, not just the Routes slice.
func TestConcurrentRouter_ParametersNoRace(t *testing.T) {
	router := NewConcurrentRouter(NewRouter(RouterConfig{}))

	// Add routes with parameters
	_, err := router.Add(Route{
		Method:  http.MethodGet,
		Path:    "/users/:id/:name",
		Handler: handlerFunc,
	})
	assert.NoError(t, err)

	_, err = router.Add(Route{
		Method:  http.MethodPost,
		Path:    "/posts/:postId/comments/:commentId",
		Handler: handlerFunc,
	})
	assert.NoError(t, err)

	var wg sync.WaitGroup
	var paramsAccessCount atomic.Int64
	var addCount atomic.Int64

	// Launch 3 goroutines that read Parameters repeatedly
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				routes := router.Routes()
				// Actually access the Parameters slice data
				// This would cause a data race if Parameters weren't deep-copied
				for _, r := range routes {
					for _, p := range r.Parameters {
						_ = len(p)      // Read parameter name length
						if len(p) > 0 { // Read first character
							_ = p[0]
						}
					}
					paramsAccessCount.Add(int64(len(r.Parameters)))
				}
			}
		}(i)
	}

	// Launch 2 goroutines that add routes with parameters concurrently
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				path := fmt.Sprintf("/api/:v%d/resource/:id", goroutineID*100+j)
				_, err := router.Add(Route{
					Method:  http.MethodPost,
					Path:    path,
					Handler: handlerFunc,
				})
				if err == nil {
					addCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify operations completed
	assert.Equal(t, int64(40), addCount.Load(), "all add operations should succeed")
	assert.Greater(t, paramsAccessCount.Load(), int64(0), "should have accessed parameters")

	// Verify final state
	finalRoutes := router.Routes()
	assert.Len(t, finalRoutes, 42, "should have 2 initial + 40 added routes")

	// Verify we can still safely access Parameters after concurrent operations
	for _, route := range finalRoutes {
		for _, param := range route.Parameters {
			assert.NotEmpty(t, param, "parameter name should not be empty")
		}
	}
}

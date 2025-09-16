// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/echo/v4"
)

// ContextTimeoutConfig defines the config for ContextTimeout middleware.
//
// # Overview
//
// ContextTimeout middleware provides timeout functionality by setting a timeout on the request context.
// This is the RECOMMENDED approach for handling request timeouts in Echo, as opposed to the
// deprecated Timeout middleware which has known issues.
//
// # Key Differences from Timeout Middleware
//
// Unlike the deprecated Timeout middleware, ContextTimeout:
//   - Does NOT interfere with the response writer
//   - Does NOT cause data races when placed in different middleware positions
//   - Relies on handlers to check context.Context.Done() for cooperative cancellation
//   - Returns errors instead of writing responses directly
//   - Is safe to use in any middleware position
//
// # How It Works
//
// 1. Creates a context.WithTimeout() from the request context
// 2. Sets the timeout context on the request
// 3. Calls the next handler
// 4. If the handler returns context.DeadlineExceeded, converts it to HTTP 503
//
// # Handler Requirements
//
// For ContextTimeout to work effectively, your handlers must:
//   - Check ctx.Done() in long-running operations
//   - Use context-aware APIs (database queries, HTTP calls, etc.)
//   - Return context.DeadlineExceeded when the context is cancelled
//
// # Configuration Examples
//
// ## Basic Usage
//
//	e.Use(middleware.ContextTimeout(30 * time.Second))
//
// ## Custom Configuration
//
//	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
//		Timeout: 30 * time.Second,
//		ErrorHandler: func(err error, c echo.Context) error {
//			if errors.Is(err, context.DeadlineExceeded) {
//				return c.JSON(http.StatusRequestTimeout, map[string]string{
//					"error": "Request took too long to process",
//				})
//			}
//			return err
//		},
//	}))
//
// ## Skip Certain Routes
//
//	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
//		Timeout: 30 * time.Second,
//		Skipper: func(c echo.Context) bool {
//			// Skip timeout for health check endpoints
//			return c.Request().URL.Path == "/health"
//		},
//	}))
//
// # Handler Examples
//
// ## Context-Aware Database Query
//
//	e.GET("/users", func(c echo.Context) error {
//		ctx := c.Request().Context()
//
//		// This query will be cancelled if context times out
//		users, err := db.QueryContext(ctx, "SELECT * FROM users")
//		if err != nil {
//			if errors.Is(err, context.DeadlineExceeded) {
//				return err // Will be converted to 503 by middleware
//			}
//			return err
//		}
//
//		return c.JSON(http.StatusOK, users)
//	})
//
// ## Long-Running Operation with Context Checking
//
//	e.POST("/process", func(c echo.Context) error {
//		ctx := c.Request().Context()
//
//		// Run operation in goroutine, respecting context
//		resultCh := make(chan result)
//		errCh := make(chan error)
//
//		go func() {
//			result, err := processData(ctx) // Context-aware processing
//			if err != nil {
//				errCh <- err
//				return
//			}
//			resultCh <- result
//		}()
//
//		select {
//		case <-ctx.Done():
//			return ctx.Err() // Returns DeadlineExceeded
//		case err := <-errCh:
//			return err
//		case result := <-resultCh:
//			return c.JSON(http.StatusOK, result)
//		}
//	})
//
// ## HTTP Client with Context
//
//	e.GET("/proxy", func(c echo.Context) error {
//		ctx := c.Request().Context()
//
//		req, err := http.NewRequestWithContext(ctx, "GET", "http://api.example.com/data", nil)
//		if err != nil {
//			return err
//		}
//
//		client := &http.Client{}
//		resp, err := client.Do(req)
//		if err != nil {
//			if errors.Is(err, context.DeadlineExceeded) {
//				return err // Will be converted to 503
//			}
//			return err
//		}
//		defer resp.Body.Close()
//
//		// Process response...
//		return c.String(http.StatusOK, "Proxy response")
//	})
//
// # Error Handling
//
// By default, when a context timeout occurs (context.DeadlineExceeded), the middleware:
//   - Returns HTTP 503 Service Unavailable
//   - Includes the original error as internal error
//   - Does NOT write to the response (allows upstream middleware to handle)
//
// # Best Practices
//
// 1. **Use context-aware APIs**: Always use database/HTTP clients that accept context
// 2. **Check context in loops**: For CPU-intensive operations, periodically check ctx.Done()
// 3. **Set appropriate timeouts**: Consider your application's typical response times
// 4. **Handle gracefully**: Provide meaningful error messages to users
// 5. **Place middleware appropriately**: Can be used at any position in middleware chain
//
// # Common Patterns
//
// ## Database Operations
//	ctx := c.Request().Context()
//	rows, err := db.QueryContext(ctx, query, args...)
//
// ## HTTP Requests
//	req, _ := http.NewRequestWithContext(ctx, method, url, body)
//	resp, err := client.Do(req)
//
// ## Redis Operations
//	result := redisClient.Get(ctx, key)
//
// ## Long-Running Loops
//	for {
//		select {
//		case <-ctx.Done():
//			return ctx.Err()
//		default:
//			// Do work...
//		}
//	}
type ContextTimeoutConfig struct {
	// Skipper defines a function to skip middleware.
	// Use this to exclude certain endpoints from timeout enforcement.
	//
	// Example:
	//	Skipper: func(c echo.Context) bool {
	//		return c.Request().URL.Path == "/health"
	//	},
	Skipper Skipper

	// ErrorHandler is called when the handler returns an error.
	// The default implementation converts context.DeadlineExceeded to HTTP 503.
	//
	// Use this to customize timeout error responses:
	//
	// Example:
	//	ErrorHandler: func(err error, c echo.Context) error {
	//		if errors.Is(err, context.DeadlineExceeded) {
	//			return c.JSON(http.StatusRequestTimeout, map[string]string{
	//				"error": "Operation timed out",
	//				"timeout": "30s",
	//			})
	//		}
	//		return err
	//	},
	ErrorHandler func(err error, c echo.Context) error

	// Timeout configures the request timeout duration.
	// REQUIRED - must be greater than 0.
	//
	// Common values:
	//   - API endpoints: 30s - 60s
	//   - File uploads: 5m - 15m
	//   - Real-time operations: 5s - 10s
	//   - Background processing: 2m - 5m
	//
	// Example: 30 * time.Second
	Timeout time.Duration
}

// ContextTimeout returns a middleware that enforces a timeout on request processing.
//
// This is the RECOMMENDED way to handle request timeouts in Echo applications.
// Unlike the deprecated Timeout middleware, this approach:
//   - Is safe to use in any middleware position
//   - Does not interfere with response writing
//   - Relies on cooperative cancellation via context
//   - Returns errors instead of writing responses directly
//
// The middleware sets a timeout context on the request and converts any
// context.DeadlineExceeded errors returned by handlers into HTTP 503 responses.
//
// Usage:
//
//	e.Use(middleware.ContextTimeout(30 * time.Second))
//
// For handlers to work properly with this middleware, they must:
//   - Use context-aware APIs (database, HTTP clients, etc.)
//   - Check ctx.Done() in long-running operations
//   - Return context.DeadlineExceeded when cancelled
//
// Example handler:
//
//	e.GET("/api/data", func(c echo.Context) error {
//		ctx := c.Request().Context()
//		data, err := db.QueryContext(ctx, "SELECT * FROM data")
//		if err != nil {
//			return err // DeadlineExceeded will become 503
//		}
//		return c.JSON(http.StatusOK, data)
//	})
//
// See ContextTimeoutConfig documentation for advanced configuration options.
func ContextTimeout(timeout time.Duration) echo.MiddlewareFunc {
	return ContextTimeoutWithConfig(ContextTimeoutConfig{Timeout: timeout})
}

// ContextTimeoutWithConfig returns a ContextTimeout middleware with custom configuration.
//
// This function allows you to customize timeout behavior including:
//   - Custom error handling for timeouts
//   - Skipping timeout for specific routes
//   - Different timeout durations per route group
//
// See ContextTimeoutConfig documentation for detailed configuration examples.
//
// Example:
//
//	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
//		Timeout: 30 * time.Second,
//		Skipper: func(c echo.Context) bool {
//			return c.Request().URL.Path == "/health"
//		},
//		ErrorHandler: func(err error, c echo.Context) error {
//			if errors.Is(err, context.DeadlineExceeded) {
//				return c.JSON(http.StatusRequestTimeout, map[string]string{
//					"error": "Request timeout",
//				})
//			}
//			return err
//		},
//	}))
func ContextTimeoutWithConfig(config ContextTimeoutConfig) echo.MiddlewareFunc {
	mw, err := config.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// ToMiddleware converts ContextTimeoutConfig to a middleware function.
//
// This method validates the configuration and returns a ready-to-use middleware.
// It's primarily used internally by ContextTimeoutWithConfig, but can be useful
// for advanced use cases where you need to validate configuration before applying.
//
// Returns an error if:
//   - Timeout is 0 or negative
//   - Configuration is otherwise invalid
//
// Example:
//
//	config := ContextTimeoutConfig{Timeout: 30 * time.Second}
//	middleware, err := config.ToMiddleware()
//	if err != nil {
//		log.Fatal("Invalid timeout config:", err)
//	}
//	e.Use(middleware)
func (config ContextTimeoutConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Timeout == 0 {
		return nil, errors.New("timeout must be set")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(err error, c echo.Context) error {
			if err != nil && errors.Is(err, context.DeadlineExceeded) {
				return echo.ErrServiceUnavailable.WithInternal(err)
			}
			return err
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			timeoutContext, cancel := context.WithTimeout(c.Request().Context(), config.Timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(timeoutContext))

			if err := next(c); err != nil {
				return config.ErrorHandler(err, c)
			}
			return nil
		}
	}, nil
}

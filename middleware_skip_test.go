package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkMiddleware404 compares performance when middleware is skipped on 404
func BenchmarkMiddleware404(b *testing.B) {
	e := New()

	// Simulate a "Heavy" middleware (e.g., Auth or Logging)
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			c.Set("user_id", "12345") // Simulate some work
			return next(c)
		}
	})

	e.GET("/exists", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Case 1: Standard behavior (Middleware runs on 404)
	b.Run("Normal_404", func(b *testing.B) {
		e.SkipMiddlewareOnNotFound = false
		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		w := httptest.NewRecorder()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			e.ServeHTTP(w, req)
		}
	})

	// Case 2: Optimized behavior (Middleware skipped on 404)
	b.Run("Optimized_404", func(b *testing.B) {
		e.SkipMiddlewareOnNotFound = true
		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		w := httptest.NewRecorder()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			e.ServeHTTP(w, req)
		}
	})
}

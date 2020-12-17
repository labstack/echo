package middleware

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
)

// TokenStore is the interface to be implemented by custom stores.
type TokenStore interface {
	// Stores for the rate limiter have to implement the ShouldAllow method
	ShouldAllow(identifier string) bool
}

type (
	// RateLimiterConfig defines the configuration for the rate limiter
	RateLimiterConfig struct {
		Skipper    Skipper
		BeforeFunc BeforeFunc
		// SourceFunc uses echo.Context to extract the identifier for a visitor
		SourceFunc func(context echo.Context) string
		// Store defines a store for the rate limiter
		Store TokenStore
	}
)

var DefaultRateLimiterConfig = RateLimiterConfig{
	Skipper: DefaultSkipper,
}

// RateLimiter returns a rate limiting middleware
func RateLimiter(source func(context echo.Context) string, store TokenStore) echo.MiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.SourceFunc = source
	config.Store = store

	return RateLimiterWithConfig(config)
}

// RateLimiterWithConfig returns a rate limiting middleware
func RateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRateLimiterConfig.Skipper
	}
	if config.SourceFunc == nil {
		panic("Source function must be provided")
	}
	if config.Store == nil {
		panic("Store configuration must be provided")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			identifier := config.SourceFunc(c)

			allowed := config.Store.ShouldAllow(identifier)
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, nil)
			} else {
				return next(c)
			}
		}
	}
}

// InMemoryStore is the built-in store implementation for RateLimiter
type InMemoryStore struct {
	visitors map[string]*rate.Limiter
	mutex    sync.Mutex
	rate     rate.Limit
	burst    int
}

// implements TokenStore.ShouldAllow
func (store *InMemoryStore) ShouldAllow(identifier string) bool {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.visitors == nil {
		store.visitors = make(map[string]*rate.Limiter)
	}

	limiter, exists := store.visitors[identifier]
	if !exists {
		limiter = rate.NewLimiter(store.rate, store.burst)
		store.visitors[identifier] = limiter
	}

	return limiter.Allow()
}

package middleware

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
)

// RateLimiterStore is the interface to be implemented by custom stores.
type RateLimiterStore interface {
	// Stores for the rate limiter have to implement the Allow method
	Allow(identifier string) bool
}

type (
	// RateLimiterConfig defines the configuration for the rate limiter
	RateLimiterConfig struct {
		Skipper    Skipper
		BeforeFunc BeforeFunc
		// SourceFunc uses echo.Context to extract the identifier for a visitor
		SourceFunc func(context echo.Context) string
		// Store defines a store for the rate limiter
		Store RateLimiterStore
	}
)

// DefaultRateLimiterConfig defines default values for RateLimiterConfig
var DefaultRateLimiterConfig = RateLimiterConfig{
	Skipper: DefaultSkipper,
	SourceFunc: func(ctx echo.Context) string {
		return ctx.RealIP()
	},
}

// RateLimiter returns a rate limiting middleware
func RateLimiter(store RateLimiterStore) echo.MiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.Store = store

	return RateLimiterWithConfig(config)
}

// RateLimiterWithConfig returns a rate limiting middleware
func RateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRateLimiterConfig.Skipper
	}
	if config.SourceFunc == nil {
		config.SourceFunc = DefaultRateLimiterConfig.SourceFunc
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

			allowed := config.Store.Allow(identifier)
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, nil)
			}
			return next(c)
		}
	}
}

// RateLimiterMemoryStore is the built-in store implementation for RateLimiter
type RateLimiterMemoryStore struct {
	visitors map[string]*rate.Limiter
	mutex    sync.Mutex
	rate     rate.Limit
	burst    int
}

// Allow implements RateLimiterStore.Allow
func (store *RateLimiterMemoryStore) Allow(identifier string) bool {
	store.mutex.Lock()

	if store.visitors == nil {
		store.visitors = make(map[string]*rate.Limiter)
	}

	limiter, exists := store.visitors[identifier]
	if !exists {
		limiter = rate.NewLimiter(store.rate, store.burst)
		store.visitors[identifier] = limiter
	}

	store.mutex.Unlock()
	return limiter.Allow()
}

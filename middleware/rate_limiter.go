package middleware

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
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
		// IdentifierExtractor uses echo.Context to extract the identifier for a visitor
		IdentifierExtractor Extractor
		// Store defines a store for the rate limiter
		Store        RateLimiterStore
		ErrorHandler ErrorHandler
		DenyHandler  ErrorHandler
	}
	// ErrorHandler provides a handler for returning errors from the middleware
	ErrorHandler func(context echo.Context) error
	// Extractor is used to extract data from echo.Context
	Extractor func(context echo.Context) (string, error)
)

// DefaultRateLimiterConfig defines default values for RateLimiterConfig
var DefaultRateLimiterConfig = RateLimiterConfig{
	Skipper: DefaultSkipper,
	IdentifierExtractor: func(ctx echo.Context) (string, error) {
		id := ctx.RealIP()
		return id, nil
	},
	ErrorHandler: func(context echo.Context) error {
		return context.JSON(http.StatusTooManyRequests, nil)
	},
	DenyHandler: func(context echo.Context) error {
		return context.JSON(http.StatusForbidden, nil)
	},
}

/*
RateLimiter returns a rate limiting middleware

	e := echo.New()

	var inMemoryStore = new(RateLimiterMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	e.GET("/rate-limited", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, RateLimiter(inMemoryStore))
 */
func RateLimiter(store RateLimiterStore) echo.MiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.Store = store

	return RateLimiterWithConfig(config)
}

/*
RateLimiterWithConfig returns a rate limiting middleware

	e := echo.New()

	var inMemoryStore = new(RateLimiterMemoryStore)
	inMemoryStore.rate = 1
	inMemoryStore.burst = 3

	config := RateLimiterConfig{
		Skipper: DefaultSkipper,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
		DenyHandler: func(context echo.Context) error {
			return context.JSON(http.StatusForbidden, nil)
		},
	}

	e.GET("/rate-limited", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, RateLimiterWithConfig(config))
*/
func RateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRateLimiterConfig.Skipper
	}
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = DefaultRateLimiterConfig.IdentifierExtractor
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultRateLimiterConfig.ErrorHandler
	}
	if config.DenyHandler == nil {
		config.DenyHandler = DefaultRateLimiterConfig.DenyHandler
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

			identifier, err := config.IdentifierExtractor(c)
			if err != nil {
				return config.DenyHandler(c)
			}

			if !config.Store.Allow(identifier) {
				return config.ErrorHandler(c)
			}
			return next(c)
		}
	}
}

// RateLimiterMemoryStore is the built-in store implementation for RateLimiter
type (
	RateLimiterMemoryStore struct {
		visitors map[string]Visitor
		mutex    sync.Mutex
		rate     rate.Limit
		burst    int
	}
	// Visitor signifies a unique user's limiter details
	Visitor struct {
		*rate.Limiter
		lastSeen time.Time
	}
)

// Allow implements RateLimiterStore.Allow
func (store *RateLimiterMemoryStore) Allow(identifier string) bool {
	store.mutex.Lock()

	if store.visitors == nil {
		store.visitors = make(map[string]Visitor)
	}

	limiter, exists := store.visitors[identifier]
	if !exists {
		limiter.Limiter = rate.NewLimiter(store.rate, store.burst)
		limiter.lastSeen = time.Now()
		store.visitors[identifier] = limiter
	}
	limiter.lastSeen = time.Now()
	store.mutex.Unlock()
	return limiter.Allow()
}

/*
CleanupStaleVisitors helps manage the size of the visitors map by removing stale records
of users who haven't visited again in the past 3 mins
 */
func (store *RateLimiterMemoryStore) CleanupStaleVisitors() {
	store.mutex.Lock()
	for id, visitor := range store.visitors {
		if time.Since(visitor.lastSeen) > 3*time.Minute {
			delete(store.visitors, id)
		}
	}
	store.mutex.Unlock()
}

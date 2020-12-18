package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
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
		visitors    map[string]Visitor
		mutex       sync.Mutex
		rate        rate.Limit
		burst       int
		expiresIn   time.Duration
		lastCleanup time.Time
	}
	// Visitor signifies a unique user's limiter details
	Visitor struct {
		*rate.Limiter
		lastSeen time.Time
	}
)

// NewRateLimiterMemoryStore returns an instance of RateLimiterMemoryStore
func NewRateLimiterMemoryStore(
	rate rate.Limit,
	burst int,
	/* using this variadic because I don't see a better way to make it optional. suggestions are welcome */
	expiresIn ...time.Duration) (store RateLimiterMemoryStore) {
	store.rate = rate
	store.burst = burst
	if len(expiresIn) == 0 || expiresIn[0] == 0 {
		store.expiresIn = 3 * time.Minute
	} else {
		store.expiresIn = expiresIn[0]
	}
	store.lastCleanup = time.Now()
	return
}

// Allow implements RateLimiterStore.Allow
func (store *RateLimiterMemoryStore) Allow(identifier string) bool {
	store.mutex.Lock()
	if time.Since(store.lastCleanup) > store.expiresIn {
		store.cleanupStaleVisitors()
	}

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
cleanupStaleVisitors helps manage the size of the visitors map by removing stale records
of users who haven't visited again after the configured expiry time has elapsed
*/
func (store *RateLimiterMemoryStore) cleanupStaleVisitors() {
	for id, visitor := range store.visitors {
		if time.Since(visitor.lastSeen) > store.expiresIn {
			delete(store.visitors, id)
		}
	}
	store.lastCleanup = time.Now()
}

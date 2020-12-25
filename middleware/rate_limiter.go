package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type (
	// RateLimiterStore is the interface to be implemented by custom stores.
	RateLimiterStore interface {
		// Stores for the rate limiter have to implement the Allow method
		Allow(identifier string) bool
	}
)

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
		return context.JSON(http.StatusForbidden, nil)
	},
	DenyHandler: func(context echo.Context) error {
		return context.JSON(http.StatusTooManyRequests, nil)
	},
}

/*
RateLimiter returns a rate limiting middleware

	e := echo.New()

	limiterStore := middleware.NewRateLimiterMemoryStore(
		middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 30, ExpiresIn: 3 * time.Minute}
	)

	e.GET("/rate-limited", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, RateLimiter(limiterStore))
*/
func RateLimiter(store RateLimiterStore) echo.MiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.Store = store

	return RateLimiterWithConfig(config)
}

/*
RateLimiterWithConfig returns a rate limiting middleware

	e := echo.New()

	config := middleware.RateLimiterConfig{
		Skipper: DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStore(
			middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 30, ExpiresIn: 3 * time.Minute}
		)
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
	}, middleware.RateLimiterWithConfig(config))
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
				return config.ErrorHandler(c)
			}

			if !config.Store.Allow(identifier) {
				return config.DenyHandler(c)
			}
			return next(c)
		}
	}
}

type (
	// RateLimiterMemoryStore is the built-in store implementation for RateLimiter
	RateLimiterMemoryStore struct {
		visitors    map[string]*Visitor
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
func NewRateLimiterMemoryStore(config RateLimiterMemoryStoreConfig) (store *RateLimiterMemoryStore) {
	store = &RateLimiterMemoryStore{}

	store.rate = config.Rate
	store.burst = config.Burst
	if config.ExpiresIn == 0 {
		store.expiresIn = DefaultRateLimiterMemoryStoreConfig.ExpiresIn
	} else {
		store.expiresIn = config.ExpiresIn
	}
	store.visitors = make(map[string]*Visitor)
	store.lastCleanup = now()
	return
}

// RateLimiterMemoryStoreConfig represents configuration for RateLimiterMemoryStore
type RateLimiterMemoryStoreConfig struct {
	Rate      rate.Limit
	Burst     int
	ExpiresIn time.Duration
}

// DefaultRateLimiterMemoryStoreConfig provides default configuration values for RateLimiterMemoryStore
var DefaultRateLimiterMemoryStoreConfig = RateLimiterMemoryStoreConfig{
	ExpiresIn: 3 * time.Minute,
}

// Allow implements RateLimiterStore.Allow
func (store *RateLimiterMemoryStore) Allow(identifier string) bool {
	store.mutex.Lock()
	limiter, exists := store.visitors[identifier]
	if !exists {
		limiter = new(Visitor)
		limiter.Limiter = rate.NewLimiter(store.rate, store.burst)
		limiter.lastSeen = now()
		store.visitors[identifier] = limiter
	}
	limiter.lastSeen = now()
	if now().Sub(store.lastCleanup) > store.expiresIn {
		store.cleanupStaleVisitors()
	}
	store.mutex.Unlock()
	return limiter.AllowN(now(), 1)
}

/*
cleanupStaleVisitors helps manage the size of the visitors map by removing stale records
of users who haven't visited again after the configured expiry time has elapsed
*/
func (store *RateLimiterMemoryStore) cleanupStaleVisitors() {
	for id, visitor := range store.visitors {
		if now().Sub(visitor.lastSeen) > store.expiresIn {
			delete(store.visitors, id)
		}
	}
	store.lastCleanup = now()
}

/*
actual time method which is mocked in test file
*/
var now = func() time.Time {
	return time.Now()
}

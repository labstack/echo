// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RetryMaxStore is the interface to be implemented by custom retry stores.
type RetryMaxStore interface {
	// AllowAttempt returns true if another attempt is allowed for the identifier.
	AllowAttempt(identifier string) (bool, error)
	// ResetAttempts resets the attempt counter for the identifier.
	ResetAttempts(identifier string) error
}

// RetryableChecker determines if an error should trigger a retry
type RetryableChecker func(err error) bool

// BackoffStrategy defines how to calculate retry delays
type BackoffStrategy func(attempt int, minTimeout, maxTimeout time.Duration) time.Duration

// RetryMaxConfig defines the configuration for the RetryMax middleware.
type RetryMaxConfig struct {
	Skipper    Skipper
	BeforeFunc BeforeFunc
	// IdentifierExtractor extracts an identifier (e.g., IP, user ID) for tracking retries.
	IdentifierExtractor Extractor
	// Store defines a store for retry counts (optional - if nil, retries per request).
	Store RetryMaxStore
	// MaxAttempts is the total number of attempts before giving up.
	MaxAttempts int
	// MinTimeout is the minimum wait between attempts.
	MinTimeout time.Duration
	// MaxTimeout is the maximum wait between attempts.
	MaxTimeout time.Duration
	// BackoffStrategy determines how retry delays are calculated.
	BackoffStrategy BackoffStrategy
	// RetryableChecker determines if an error should trigger a retry.
	RetryableChecker RetryableChecker
	// ErrorHandler is called if IdentifierExtractor fails.
	ErrorHandler func(c echo.Context, err error) error
	// DenyHandler is called when the max attempts are exhausted.
	DenyHandler func(c echo.Context, identifier string, err error) error
	// OnRetry is called before each retry attempt.
	OnRetry func(c echo.Context, identifier string, attempt int, err error)
}

// ExponentialBackoff implements exponential backoff with jitter
func ExponentialBackoff(attempt int, minTimeout, maxTimeout time.Duration) time.Duration {
	backoff := min(time.Duration(math.Pow(2, float64(attempt-1)))*minTimeout, maxTimeout)
	// Add up to 10% jitter to prevent thundering herd
	jitter := time.Duration(float64(backoff) * 0.1 * (0.5 - float64(time.Now().UnixNano()%1000)/1000))
	return backoff + jitter
}

// LinearBackoff implements linear backoff
func LinearBackoff(attempt int, minTimeout, maxTimeout time.Duration) time.Duration {
	backoff := time.Duration(attempt) * minTimeout
	if backoff > maxTimeout {
		backoff = maxTimeout
	}
	return backoff
}

// DefaultRetryableChecker determines if an error is retryable
func DefaultRetryableChecker(err error) bool {
	if err == nil {
		return false
	}

	// Check for HTTP errors that shouldn't be retried
	if httpErr, ok := err.(*echo.HTTPError); ok {
		code := httpErr.Code
		// Don't retry client errors (4xx) except specific ones
		if code >= 400 && code < 500 {
			return code == http.StatusRequestTimeout || code == http.StatusTooManyRequests
		}
	}

	// Retry on server errors and other errors by default
	return true
}

// DefaultRetryMaxConfig provides default values for RetryMaxConfig.
var DefaultRetryMaxConfig = RetryMaxConfig{
	Skipper: DefaultSkipper,
	IdentifierExtractor: func(c echo.Context) (string, error) {
		return c.RealIP(), nil
	},
	MaxAttempts:      5,
	MinTimeout:       1 * time.Second,
	MaxTimeout:       5 * time.Second,
	BackoffStrategy:  ExponentialBackoff,
	RetryableChecker: DefaultRetryableChecker,
	ErrorHandler: func(c echo.Context, err error) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "error extracting identifier")
	},
	DenyHandler: func(c echo.Context, identifier string, err error) error {
		return echo.NewHTTPError(http.StatusTooManyRequests, "max retry attempts exceeded")
	},
}

// RetryMax returns a RetryMax middleware with the given store.
func RetryMax(store RetryMaxStore) echo.MiddlewareFunc {
	config := DefaultRetryMaxConfig
	config.Store = store
	return RetryMaxWithConfig(config)
}

// RetryMaxWithConfig returns a RetryMax middleware with custom config.
func RetryMaxWithConfig(config RetryMaxConfig) echo.MiddlewareFunc {
	// Apply defaults for nil fields
	if config.Skipper == nil {
		config.Skipper = DefaultRetryMaxConfig.Skipper
	}
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = DefaultRetryMaxConfig.IdentifierExtractor
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultRetryMaxConfig.ErrorHandler
	}
	if config.DenyHandler == nil {
		config.DenyHandler = DefaultRetryMaxConfig.DenyHandler
	}
	if config.BackoffStrategy == nil {
		config.BackoffStrategy = DefaultRetryMaxConfig.BackoffStrategy
	}
	if config.RetryableChecker == nil {
		config.RetryableChecker = DefaultRetryMaxConfig.RetryableChecker
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultRetryMaxConfig.MaxAttempts
	}
	if config.MinTimeout <= 0 {
		config.MinTimeout = DefaultRetryMaxConfig.MinTimeout
	}
	if config.MaxTimeout <= 0 {
		config.MaxTimeout = DefaultRetryMaxConfig.MaxTimeout
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			// If using a store, check identifier extraction
			var identifier string
			var err error
			if config.Store != nil {
				identifier, err = config.IdentifierExtractor(c)
				if err != nil {
					return config.ErrorHandler(c, err)
				}
			}

			var lastErr error
			ctx := c.Request().Context()

			for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
				// If using a store, check if we're allowed to attempt
				if config.Store != nil {
					allowed, err := config.Store.AllowAttempt(identifier)
					if err != nil {
						return config.ErrorHandler(c, err)
					}
					if !allowed {
						return config.DenyHandler(c, identifier, lastErr)
					}
				}

				// Execute the handler
				lastErr = next(c)

				// Success case
				if lastErr == nil {
					if config.Store != nil {
						_ = config.Store.ResetAttempts(identifier)
					}
					return nil
				}

				// Check if error is retryable
				if !config.RetryableChecker(lastErr) {
					return lastErr
				}

				// If this was the last attempt, return the error
				if attempt >= config.MaxAttempts {
					return lastErr
				}

				// Call retry callback if provided
				if config.OnRetry != nil {
					config.OnRetry(c, identifier, attempt, lastErr)
				}

				// Calculate and apply backoff
				backoff := config.BackoffStrategy(attempt, config.MinTimeout, config.MaxTimeout)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}

			return lastErr
		}
	}
}

// RetryMaxMemoryStore is an improved in-memory store for RetryMax.
type RetryMaxMemoryStore struct {
	attempts    map[string]int
	lastAccess  map[string]time.Time
	mutex       sync.RWMutex
	expiresIn   time.Duration
	maxAttempts int
}

// NewRetryMaxMemoryStore creates a new RetryMaxMemoryStore.
func NewRetryMaxMemoryStore(expiresIn time.Duration) *RetryMaxMemoryStore {
	return NewRetryMaxMemoryStoreWithAttempts(expiresIn, 5)
}

// NewRetryMaxMemoryStoreWithAttempts creates a new RetryMaxMemoryStore with custom max attempts.
func NewRetryMaxMemoryStoreWithAttempts(expiresIn time.Duration, maxAttempts int) *RetryMaxMemoryStore {
	store := &RetryMaxMemoryStore{
		attempts:    make(map[string]int),
		lastAccess:  make(map[string]time.Time),
		expiresIn:   expiresIn,
		maxAttempts: maxAttempts,
	}

	// Start cleanup goroutine
	go store.startCleanup()

	return store
}

// AllowAttempt implements RetryMaxStore.AllowAttempt.
func (s *RetryMaxMemoryStore) AllowAttempt(identifier string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if expired and reset if so
	if s.isExpiredUnsafe(identifier) {
		s.attempts[identifier] = 0
	}

	// Increment attempt count
	s.attempts[identifier]++
	s.lastAccess[identifier] = time.Now()

	// Check if we've exceeded max attempts
	return s.attempts[identifier] <= s.maxAttempts, nil
}

// ResetAttempts implements RetryMaxStore.ResetAttempts.
func (s *RetryMaxMemoryStore) ResetAttempts(identifier string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.attempts[identifier] = 0
	s.lastAccess[identifier] = time.Now()
	return nil
}

// GetAttempts returns the current attempt count for an identifier.
func (s *RetryMaxMemoryStore) GetAttempts(identifier string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.isExpiredUnsafe(identifier) {
		return 0
	}

	return s.attempts[identifier]
}

// Cleanup removes expired entries.
func (s *RetryMaxMemoryStore) Cleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for identifier, lastAccess := range s.lastAccess {
		if now.Sub(lastAccess) > s.expiresIn {
			delete(s.attempts, identifier)
			delete(s.lastAccess, identifier)
		}
	}

	return nil
}

// isExpiredUnsafe checks if an entry has expired (no locking, caller must hold lock).
func (s *RetryMaxMemoryStore) isExpiredUnsafe(identifier string) bool {
	lastAccess, exists := s.lastAccess[identifier]
	return !exists || time.Since(lastAccess) > s.expiresIn
}

// startCleanup runs periodic cleanup in a goroutine.
func (s *RetryMaxMemoryStore) startCleanup() {
	ticker := time.NewTicker(s.expiresIn / 2) // Clean up twice per expiry period
	defer ticker.Stop()

	for range ticker.C {
		_ = s.Cleanup()
	}
}

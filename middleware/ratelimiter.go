package middleware

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type (
	// RateLimiterConfig defines the config for RateLimiter middleware.
	RateLimiterConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// The max count in duration for no policy, default is 100.
		Max      int

		// Count duration for no policy, default is 1 Minute.
		Duration time.Duration
		//key prefix, default is "LIMIT:".
		Prefix   string

	}

	limiter struct {
		abstractLimiter
		prefix string
	}

	// Result of limiter.Get
	Result struct {
		Total     int           // It Equals Options.Max, or policy max
		Remaining int           // It will always >= -1
		Duration  time.Duration // It Equals Options.Duration, or policy duration
		Reset     time.Time     // The limit record reset time
	}

	abstractLimiter interface {
		getLimit(key string, policy ...int) ([]interface{}, error)
		removeLimit(key string) error
	}

)

var (
	// DefaultRateLimiterConfig is the default rate limit middleware config.
	DefaultRateLimiterConfig = RateLimiterConfig{
		Skipper:      DefaultSkipper,
		Max:100,
		Duration: time.Minute * 1,
		Prefix:"LIMIT",
	}
	limiterImp *limiter
)

// RateLimiter returns a rate limit middleware.
func RateLimiter() echo.MiddlewareFunc {
	return RateLimiterWithConfig(DefaultRateLimiterConfig)
}

// RateLimiterWithConfig returns a RateLimiter middleware with config.
// See: `RateLimiter()`.
func RateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCORSConfig.Skipper
	}

	if config.Prefix == "" {
		config.Prefix = "LIMIT:"
	}
	if config.Max <= 0 {
		config.Max = 100
	}
	if config.Duration <= 0 {
		config.Duration = time.Minute *1
	}
	limiterImp = newMemoryLimiter(&config)
	/*
	if config.Client == nil {

	}else{
		//setup redis client
		//limiter = newRedisLimiter(&config)
	}
	*/

	fmt.Printf("Max:%d",config.Max)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			response := c.Response()

			result, err := limiterImp.Get(c.Path())

			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			response.Header().Set("X-Ratelimit-Limit", strconv.FormatInt(int64(result.Total), 10))
			response.Header().Set("X-Ratelimit-Remaining", strconv.FormatInt(int64(result.Remaining), 10))
			response.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(result.Reset.Unix(), 10))

			if result.Remaining < 0 {

				after := int64(result.Reset.Sub(time.Now())) / 1e9
				response.Header().Set("Retry-After", strconv.FormatInt(after, 10))
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded, retry in %d seconds.\n", after)
			}
			return next(c)
		}
	}
}

// get & remove

func (l *limiter) Get(id string, policy ...int) (Result, error) {
	var result Result
	key := l.prefix + id

	if odd := len(policy) % 2; odd == 1 {
		return result, errors.New("ratelimiter: must be paired values")
	}

	res, err := l.getLimit(key, policy...)
	if err != nil {
		return result, err
	}

	result = Result{}
	switch res[3].(type) {
	case time.Time: // result from memory limiter
		result.Remaining = res[0].(int)
		result.Total = res[1].(int)
		result.Duration = res[2].(time.Duration)
		result.Reset = res[3].(time.Time)
	default: // result from redis limiter
		result.Remaining = int(res[0].(int64))
		result.Total = int(res[1].(int64))
		result.Duration = time.Duration(res[2].(int64) * 1e6)

		timestamp := res[3].(int64)
		sec := timestamp / 1000
		result.Reset = time.Unix(sec, (timestamp-(sec*1000))*1e6)
	}
	return result, nil
}

// Remove remove limiter record for id
func (l *limiter) Remove(id string) error {
	return l.removeLimit(l.prefix + id)
}


// Inmemory Limiter imp

// policy status
type (
	statusCacheItem struct {
		index  int
		expire time.Time
	}

	// limit status
	limiterCacheItem struct {
		total     int
		remaining int
		duration  time.Duration
		expire    time.Time
	}

	memoryLimiter struct {
		max      int
		duration time.Duration
		status   map[string]*statusCacheItem
		store    map[string]*limiterCacheItem
		ticker   *time.Ticker
		lock     sync.Mutex
	}
)

func newMemoryLimiter(opts *RateLimiterConfig) *limiter {
	m := &memoryLimiter{
		max:      opts.Max,
		duration: opts.Duration,
		store:    make(map[string]*limiterCacheItem),
		status:   make(map[string]*statusCacheItem),
		ticker:   time.NewTicker(time.Second),
	}
	go m.cleanCache()
	return &limiter{m, opts.Prefix}
}

// abstractLimiter interface
func (m *memoryLimiter) getLimit(key string, policy ...int) ([]interface{}, error) {
	length := len(policy)
	var args []int
	if length == 0 {
		args = []int{m.max, int(m.duration / time.Millisecond)}
	} else {
		args = make([]int, length)
		for i, val := range policy {
			if val <= 0 {
				return nil, errors.New("ratelimiter: must be positive integer")
			}
			args[i] = policy[i]
		}
	}

	res := m.getItem(key, args...)
	m.lock.Lock()
	defer m.lock.Unlock()
	return []interface{}{res.remaining, res.total, res.duration, res.expire}, nil
}

// abstractLimiter interface
func (m *memoryLimiter) removeLimit(key string) error {
	statusKey := "{" + key + "}:S"
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.store, key)
	delete(m.status, statusKey)
	return nil
}

func (m *memoryLimiter) clean() {
	m.lock.Lock()
	defer m.lock.Unlock()
	start := time.Now()
	expireTime := start.Add(time.Millisecond * 100)
	frequency := 24
	var expired int
	for {
	label:
		for i := 0; i < frequency; i++ {
			for key, value := range m.store {
				if value.expire.Add(value.duration).Before(start) {
					statusKey := "{" + key + "}:S"
					delete(m.store, key)
					delete(m.status, statusKey)
					expired++
				}
				break
			}
		}
		if expireTime.Before(time.Now()) {
			return
		}
		if expired > frequency/4 {
			expired = 0
			goto label
		}
		return
	}
}

func (m *memoryLimiter) getItem(key string, args ...int) (res *limiterCacheItem) {
	policyCount := len(args) / 2
	statusKey := "{" + key + "}:S"

	m.lock.Lock()
	defer m.lock.Unlock()
	var ok bool
	if res, ok = m.store[key]; !ok {
		res = &limiterCacheItem{
			total:     args[0],
			remaining: args[0] - 1,
			duration:  time.Duration(args[1]) * time.Millisecond,
			expire:    time.Now().Add(time.Duration(args[1]) * time.Millisecond),
		}
		m.store[key] = res
		return
	}
	if res.expire.After(time.Now()) {
		if policyCount > 1 && res.remaining-1 == -1 {
			statusItem, ok := m.status[statusKey]
			if ok {
				statusItem.expire = time.Now().Add(res.duration * 2)
				statusItem.index++
			} else {
				statusItem := &statusCacheItem{
					index:  2,
					expire: time.Now().Add(time.Duration(args[1]) * time.Millisecond * 2),
				}
				m.status[statusKey] = statusItem
			}
		}
		if res.remaining >= 0 {
			res.remaining--
		} else {
			res.remaining = -1
		}
	} else {
		index := 1
		if policyCount > 1 {
			if statusItem, ok := m.status[statusKey]; ok {
				if statusItem.expire.Before(time.Now()) {
					index = 1
				} else if statusItem.index > policyCount {
					index = policyCount
				} else {
					index = statusItem.index
				}
				statusItem.index = index
			}
		}
		total := args[(index*2)-2]
		duration := args[(index*2)-1]
		res.total = total
		res.remaining = total - 1
		res.duration = time.Duration(duration) * time.Millisecond
		res.expire = time.Now().Add(time.Duration(duration) * time.Millisecond)
	}
	return
}

func (m *memoryLimiter) cleanCache() {
	for range m.ticker.C {
		m.clean()
	}
}

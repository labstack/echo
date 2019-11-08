package middleware

import (
	"errors"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"sync"
	"time"
	"fmt"
	"strings"
)
type (
	// RateLimiterConfig defines the config for RateLimiter middleware.
	RateLimiterConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		LimitConfig LimiterConfig
		//key prefix, default is "LIMIT:".
		Prefix   string

		// Use a redis client for limiter, if omit, it will use a memory limiter.
		Client   RedisClient

		//If request gets a  internal limiter error, just skip the limiter and let it go to next middleware
		SkipRateLimiterInternalError	bool

		//TODO: WhiteList
	}

	LimiterConfig struct {
		//The max count in duration for no policy, default is 100.
		Max      int
		//ip or header(key)
		Strategy string
		//Count duration for no policy, default is 1 Minute.
		Duration time.Duration
		//If the strategy is header which header key will be used for limits. such as Header("Client-Token")
		Key           string
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

	//RedisClient interface
	RedisClient interface {
		DeleteKey(string) error
		EvalulateSha(string, []string, ...interface{}) (interface{}, error)
		LuaScriptLoad(string) (string, error)
	}

)

var (
	// DefaultRateLimiterConfig is the default rate limit middleware config.
	DefaultRateLimiterConfig = RateLimiterConfig{
		Skipper:      DefaultSkipper,
		LimitConfig:LimiterConfig{
			Max:100,
			Duration: time.Minute * 1,
			Strategy:"ip",
			Key:"",
		},
		Prefix:"LIMIT",
		Client:nil,
		SkipRateLimiterInternalError:false,
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
		config.Skipper = DefaultRateLimiterConfig.Skipper
	}
	if len(config.LimitConfig.Strategy) == 0 {
		panic("echo: rate limiter middleware requires Strategy (use ip or header(key)")
	}
	if config.Prefix == "" {
		config.Prefix = "LIMIT:"
	}
	if config.LimitConfig.Max <= 0 {
		config.LimitConfig.Max = 100
	}
	if config.LimitConfig.Duration <= 0 {
		config.LimitConfig.Duration = time.Minute *1
	}

	//If config.Client omit, the limiter is a memory limiter
	if config.Client == nil {
		limiterImp = newMemoryLimiter(&config)
	}else{
		//setup redis client
		limiterImp = newRedisLimiter(&config)
	}

	var tokenExtractor TokenExtractor

	switch strings.ToLower(config.LimitConfig.Strategy) {
	case "ip":
		tokenExtractor = IPTokenExtractor()
	case "header":
		tokenExtractor = HeaderTokenExtractor(config.LimitConfig.Key)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			response := c.Response()

			policy := []int{}
			result, err := limiterImp.Get(tokenExtractor(c),policy...)

			if err != nil {

				if config.SkipRateLimiterInternalError{
					return next(c)
				}
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			response.Header().Set("X-Ratelimit-Limit", strconv.FormatInt(int64(result.Total), 10))
			response.Header().Set("X-Ratelimit-Remaining", strconv.FormatInt(int64(result.Remaining), 10))
			response.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(result.Reset.Unix(), 10))

			if result.Remaining <= 0 {

				after := int64(result.Reset.Sub(time.Now())) / 1e9
				response.Header().Set("Retry-After", strconv.FormatInt(after, 10))
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded, retry in %d seconds.\n", after)
			}
			return next(c)
		}
	}
}


// TokenExtractor defines the interface of the functions to use in order to extract a token for each request
type TokenExtractor func(echo.Context) string

// IPTokenExtractor extracts the IP of the request
func IPTokenExtractor() TokenExtractor {
	return func(c echo.Context) string { return strings.Split(c.RealIP(), ":")[0]  }
}

// HeaderTokenExtractor returns a TokenExtractor that looks for the value of the designed header
func HeaderTokenExtractor(header string) TokenExtractor {
	return func(c echo.Context) string { return c.Request().Header.Get(header) }
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
	default: // result from disteributed limiter
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
		max:      opts.LimitConfig.Max,
		duration: opts.LimitConfig.Duration,
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




// Redis limiter imp here
type redisLimiter struct {
	sha1, max, duration string
	rc                  RedisClient
}

func newRedisLimiter(options *RateLimiterConfig) *limiter {
	sha1, err := options.Client.LuaScriptLoad(LuaScriptForRedis)
	if err != nil {
		fmt.Println("redis is not working properly. use; docker run -it -p 6379:6379 --name my-redis -d redis")
		panic(err)
	}
	r := &redisLimiter{
		rc:       options.Client,
		sha1:     sha1,
		max:      strconv.FormatInt(int64(options.LimitConfig.Max), 10),
		duration: strconv.FormatInt(int64(options.LimitConfig.Duration/time.Millisecond), 10),
	}
	return &limiter{r, options.Prefix}
}

func (r *redisLimiter) removeLimit(key string) error {
	return r.rc.DeleteKey(key)
}

func (r *redisLimiter) getLimit(key string, policy ...int) ([]interface{}, error) {
	keys := []string{key, fmt.Sprintf("{%s}:S", key)}
	capacity := 3
	length := len(policy)
	if length > 2 {
		capacity = length + 1
	}

	//fmt.Printf("redis max limit (%s) for (%s)",r.max,key)
	args := make([]interface{}, capacity, capacity)
	args[0] = genTimestamp()
	if length == 0 {
		args[1] = r.max
		args[2] = r.duration
	} else {
		for i, val := range policy {
			if val <= 0 {
				return nil, errors.New("ratelimiter: must be positive integer")
			}
			args[i+1] = strconv.FormatInt(int64(val), 10)
		}
	}

	res, err := r.rc.EvalulateSha(r.sha1, keys, args...)
	if err != nil && isNoScriptErr(err) {
		// try to load lua for cluster client and ring client for nodes changing.
		_, err = r.rc.LuaScriptLoad(LuaScriptForRedis)
		if err == nil {
			res, err = r.rc.EvalulateSha(r.sha1, keys, args...)
		}
	}

	if err == nil {
		arr, ok := res.([]interface{})
		if ok && len(arr) == 4 {
			return arr, nil
		}
		err = errors.New("Invalid result")
	}
	return nil, err
}

func genTimestamp() string {
	time := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(time, 10)
}
func isNoScriptErr(err error) bool {
	return strings.HasPrefix(err.Error(), "NOSCRIPT ")
}

//LuaScriptForRedis script loading for cluster client and ring client for nodes changing. based on links below
//https://github.com/thunks/thunk-ratelimiter
//https://github.com/thunks/thunk-ratelimiter/blob/master/ratelimite.lua
const LuaScriptForRedis string = `
-- KEYS[1] target hash key
-- KEYS[2] target status hash key
-- ARGV[n >= 3] current timestamp, max count, duration, max count, duration, ...
-- HASH: KEYS[1]
--   field:ct(count)
--   field:lt(limit)
--   field:dn(duration)
--   field:rt(reset)
local res = {}
local policyCount = (#ARGV - 1) / 2
local limit = redis.call('hmget', KEYS[1], 'ct', 'lt', 'dn', 'rt')
if limit[1] then
  res[1] = tonumber(limit[1]) - 1
  res[2] = tonumber(limit[2])
  res[3] = tonumber(limit[3]) or ARGV[3]
  res[4] = tonumber(limit[4])
  if policyCount > 1 and res[1] == -1 then
    redis.call('incr', KEYS[2])
    redis.call('pexpire', KEYS[2], res[3] * 2)
    local index = tonumber(redis.call('get', KEYS[2]))
    if index == 1 then
      redis.call('incr', KEYS[2])
    end
  end
  if res[1] >= -1 then
    redis.call('hincrby', KEYS[1], 'ct', -1)
  else
    res[1] = -1
  end
else
  local index = 1
  if policyCount > 1 then
    index = tonumber(redis.call('get', KEYS[2])) or 1
    if index > policyCount then
      index = policyCount
    end
  end
  local total = tonumber(ARGV[index * 2])
  res[1] = total - 1
  res[2] = total
  res[3] = tonumber(ARGV[index * 2 + 1])
  res[4] = tonumber(ARGV[1]) + res[3]
  redis.call('hmset', KEYS[1], 'ct', res[1], 'lt', res[2], 'dn', res[3], 'rt', res[4])
  redis.call('pexpire', KEYS[1], res[3])
end
return res
`

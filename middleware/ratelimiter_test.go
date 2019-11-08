package middleware

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"

	"crypto/rand"
	"encoding/hex"
	"time"
	"sync"
	"gopkg.in/redis.v5"
	"github.com/alicebob/miniredis"
	"errors"
	"fmt"


)

func TestRateLimiter(t *testing.T) {

	t.Run("ratelimiter middleware", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		t.Run("should return ok with x-remaining and x-limit value", func(t *testing.T) {

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			rateLimit := RateLimiter()

			h := rateLimit(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			h(c)
			assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "99")
			assert.Contains(t, rec.Header().Get("X-Ratelimit-Limit"), "100")

		})

		t.Run("should throw too many request", func(t *testing.T) {

			//ratelimit with config
			rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
				LimitConfig:LimiterConfig{
					Max: 2,
				},
			})

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			hx := rateLimitWithConfig(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			hx(c)
			hx(c)
			expectedErrorStatus := hx(c).(*echo.HTTPError)

			assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "-1")
			assert.Equal(t, http.StatusTooManyRequests, expectedErrorStatus.Code)

		})

		t.Run("should return status ok after to many request status expired", func(t *testing.T) {

			expectedDuration := time.Millisecond * 5

			//ratelimit with config; expected result getting 429 after 5 second it should return 200
			rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
				LimitConfig:LimiterConfig{
					Max: 2,
					Duration: expectedDuration,
				},
			})

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			hx := rateLimitWithConfig(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			hx(c)
			hx(c)
			expectedErrorStatus := hx(c).(*echo.HTTPError)
			assert.Equal(t, http.StatusTooManyRequests, expectedErrorStatus.Code)
			time.Sleep(expectedDuration)
			exceptedHTTPStatusOk, ok := hx(c).(*echo.HTTPError)

			if ok {
				assert.Equal(t, http.StatusOK, exceptedHTTPStatusOk.Code)
			}

		})

		t.Run("should return ok even limiter throw an exception", func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/t", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
				SkipRateLimiterInternalError:true,
			})

			h := rateLimitWithConfig(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			h(c)
			fmt.Println(rec.Result().Status)
			assert.Contains(t, rec.Result().Status, "200 OK")

		})
	})
}

// Implements RedisClient for redis.Client
type redisClient struct {
	*redis.Client
}
func (c *redisClient) DeleteKey(key string) error {
	return c.Del(key).Err()
}

func (c *redisClient) EvalulateSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.EvalSha(sha1, keys, args...).Result()
}

func (c *redisClient) LuaScriptLoad(script string) (string, error) {
	return c.ScriptLoad(script).Result()
}

// Implements RedisClient for redis.Client
type failedClient struct {
	*redis.Client
}

func (c *failedClient) DeleteKey(key string) error {
	return c.Del(key).Err()
}

func (c *failedClient) EvalulateSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return nil, errors.New("noscript mock error")
}

func (c *failedClient) LuaScriptLoad(script string) (string, error) {
	return c.ScriptLoad(script).Result()
}

func TestRedisRatelimiter(t *testing.T) {

	e := echo.New()
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	var client = redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer client.Close()

	t.Run("Redis ratelimiter middleware", func(t *testing.T) {

		t.Run("FakeRedis is running as excepted", func(t *testing.T) {
			pong, err := client.Ping().Result()
			assert.Nil(t, err)
			assert.Equal(t, "PONG", pong)
		})

		t.Run("New instrance running with redis option as excepted", func(t *testing.T) {

			rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
				Client:&redisClient{client},
			})
			assert.NotNil(t,rateLimitWithConfig)
		})

		t.Run("Get method should return ok with excepted remaining and limit values", func(t *testing.T) {


			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
				LimitConfig:LimiterConfig{
					Max:100,
					Duration:time.Minute*1,
				},
				Client:&redisClient{client},
			})

			hx := rateLimitWithConfig(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			hx(c)
			assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "99")
			assert.Contains(t, rec.Header().Get("X-Ratelimit-Limit"), "100")

		})

		t.Run("Get method should throw too many request", func(t *testing.T) {


			req := httptest.NewRequest(http.MethodGet, "/alper", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			xx := RateLimiterWithConfig(RateLimiterConfig{
				Client:&redisClient{client},
				LimitConfig:LimiterConfig{
					Max:2,
					Duration:time.Minute *1,
				},
			})

			hx := xx(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})
			hx(c)
			hx(c)
			expectedErrorStatus,ok := hx(c).(*echo.HTTPError)

			if ok{
				assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "-1")
				assert.Equal(t, http.StatusTooManyRequests, expectedErrorStatus.Code)
			}	else {
				assert.Error(t,errors.New("it should throw too many ruqest exception"))
			}
		})

	})

	t.Run("Redis ratelimiter implementation", func(t *testing.T) {
		var id= genID()
		var duration= time.Duration(60 * 1e9)
		var redisLimiter *limiter
		redisLimiter = newRedisLimiter(&RateLimiterConfig{

			Client:   &redisClient{client},
			LimitConfig:LimiterConfig{
				Max:      100,
				Duration: duration,
			},
		})

		t.Run("New instance running with failedClient should be", func(t *testing.T) {

			var limiter *limiter
			limiter = newRedisLimiter(&RateLimiterConfig{

				Client:&failedClient{client},
			})
			policy := []int{2, 100}
			res, err := limiter.Get(id, policy...)

			assert.Equal(t,"noscript mock error", err.Error())
			assert.Equal(t,0, res.Total)
			assert.Equal(t,0, res.Remaining)
			assert.Equal(t,time.Duration(0), res.Duration)
		})

		t.Run("Redislimiter.Get method should be", func(t *testing.T) {

			res, err := redisLimiter.Get(id)
			assert.Nil(t, err)
			assert.Equal(t, res.Total, 100)
			assert.Equal(t, res.Remaining, 99)
			assert.Equal(t, res.Duration, duration)
			assert.True(t, res.Reset.UnixNano() > time.Now().UnixNano())

			res, err = redisLimiter.Get(id)
			assert.Nil(t, err)
			assert.Equal(t, res.Total, 100)
			assert.Equal(t, res.Remaining, 98)

		})

		t.Run("Redislimiter.Get with invalid args should throw error", func(t *testing.T) {
			_, err := redisLimiter.Get(id, 10)
			assert.Equal(t, "ratelimiter: must be paired values", err.Error())

			_, err2 := redisLimiter.Get(id, -1, 10)
			assert.Equal(t, "ratelimiter: must be positive integer", err2.Error())

			_, err3 := redisLimiter.Get(id, 10, 0)
			assert.Equal(t, "ratelimiter: must be positive integer", err3.Error())
		})

		t.Run("Redislimiter.Get with policy", func(t *testing.T) {

			idx := genID()
			assert := assert.New(t)

			policy := []int{2, 100}

			res, err := redisLimiter.Get(idx, policy...)
			assert.Nil(err)
			assert.Equal(2, res.Total)
			assert.Equal(1, res.Remaining)
			assert.Equal(time.Millisecond*100, res.Duration)

			res, err = redisLimiter.Get(idx, policy...)
			assert.Nil(err)
			assert.Equal(0, res.Remaining)

			res, err = redisLimiter.Get(idx, policy...)
			assert.Nil(err)
			assert.Equal(-1, res.Remaining)
		})

		t.Run("Redislimiter.Remove method should be", func(t *testing.T) {

			err := redisLimiter.Remove(id)
			assert.Nil(t, err)

			err = redisLimiter.Remove(id)
			assert.Nil(t, err)

			res, err := redisLimiter.Get(id)
			assert.Nil(t, err)
			assert.Equal(t, res.Total, 100)
			assert.Equal(t, res.Remaining, 99)

		})

	})
}

func TestMemoryRateLimiter(t *testing.T) {
	t.Run("ratelimiter with default Options should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})

		id := genID()
		policy := []int{10, 1000}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		assert.Equal(1000, int(res.Duration/time.Millisecond))
		assert.True(res.Reset.After(time.Now()))
		res, err = limiter.Get(id, policy...)
		assert.Equal(10, res.Total)
		assert.Equal(8, res.Remaining)
	})

	t.Run("ratelimiter with expire should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})

		id := genID()
		policy := []int{10, 100}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(8, res.Remaining)

		time.Sleep(100 * time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
	})

	t.Run("ratelimiter with goroutine should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		policy := []int{10, 500}
		id := genID()
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		var wait sync.WaitGroup
		wait.Add(100)
		for i := 0; i < 100; i++ {
			go func() {
				limiter.Get(id, policy...)
				wait.Done()
			}()
		}
		wait.Wait()
		time.Sleep(200 * time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(-1, res.Remaining)
	})

	t.Run("ratelimiter with multi-policy should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		id := genID()
		policy := []int{3, 100, 2, 200}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.True(res.Reset.After(time.Now()))

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)
	})

	t.Run("ratelimiter with Remove id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		id := genID()
		policy := []int{10, 1000}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		limiter.Remove(id)
		res, err = limiter.Get(id, policy...)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
	})

	t.Run("ratelimiter with wrong policy id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		id := genID()
		policy := []int{10, 1000, 1}

		res, err := limiter.Get(id, policy...)
		assert.Error(err)
		assert.Equal(0, res.Total)
		assert.Equal(0, res.Remaining)
	})

	t.Run("ratelimiter with empty policy id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		id := genID()
		policy := []int{}

		res, _ := limiter.Get(id, policy...)
		assert.Equal(100, res.Total)
		assert.Equal(99, res.Remaining)
		assert.Equal(time.Minute, res.Duration)
	})

	t.Run("limiter.Get with invalid args", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		id := genID()
		_, err := limiter.Get(id, 10)
		assert.Equal("ratelimiter: must be paired values", err.Error())

		_, err2 := limiter.Get(id, -1, 10)
		assert.Equal("ratelimiter: must be positive integer", err2.Error())

		_, err3 := limiter.Get(id, 10, 0)
		assert.Equal("ratelimiter: must be positive integer", err3.Error())
	})

	t.Run("ratelimiter with Clean cache should be", func(t *testing.T) {
		assert := assert.New(t)


		limiter := &memoryLimiter{
			max:      100,
			duration: time.Minute,
			store:    make(map[string]*limiterCacheItem),
			status:   make(map[string]*statusCacheItem),
			ticker:   time.NewTicker(time.Minute),
		}

		id := genID()
		policy := []int{10, 100}

		res, _ := limiter.getLimit(id, policy...)

		assert.Equal(10, res[1].(int))
		assert.Equal(9, res[0].(int))

		time.Sleep(res[2].(time.Duration) + time.Millisecond)
		limiter.clean()
		res, _ = limiter.getLimit(id, policy...)
		assert.Equal(10, res[1].(int))
		assert.Equal(9, res[0].(int))

		time.Sleep(res[2].(time.Duration)*2 + time.Millisecond)
		limiter.clean()
		res, _ = limiter.getLimit(id, policy...)
		assert.Equal(10, res[1].(int))
		assert.Equal(9, res[0].(int))
		limiter.ticker = time.NewTicker(time.Millisecond)
		go limiter.cleanCache()
		time.Sleep(2 * time.Millisecond)
		res, _ = limiter.getLimit(id, policy...)
		assert.Equal(10, res[1].(int))
		assert.Equal(8, res[0].(int))
	})

	t.Run("ratelimiter with big goroutine should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		policy := []int{1000, 1000}
		id := genID()

		var wg sync.WaitGroup
		wg.Add(1000)
		for i := 0; i < 1000; i++ {
			go func() {
				newid := genID()
				limiter.Get(newid, policy...)
				limiter.Get(id, policy...)
				wg.Done()
			}()
		}
		wg.Wait()
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(1000, res.Total)
		assert.Equal(-1, res.Remaining)
	})

	t.Run("limiter.Get with multi-policy for expired", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})

		id := genID()
		policy := []int{2, 100, 2, 200, 3, 300, 3, 400}

		//First policy
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		//Second policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Third policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		// restore to First policy after Third policy*2 Duration
		time.Sleep(res.Duration*2 + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)
		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Second policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		//Third policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Fourth policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*400, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(-1, res.Remaining)

		// restore to First policy after Fourth policy*2 Duration
		time.Sleep(res.Duration*2 + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)
	})

	t.Run("limiter.Get with multi-policy situation for expired", func(t *testing.T) {
		assert := assert.New(t)

		var id = genID()
		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		policy := []int{2, 150, 2, 200, 3, 300, 3, 400}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)
		res, err = limiter.Get(id, policy...)

		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*400, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*400, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*150, res.Duration)

	})

	t.Run("limiter.Get with different policy time situation for expired", func(t *testing.T) {
		assert := assert.New(t)

		var id = genID()
		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		policy := []int{2, 300, 3, 100}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)
	})

	t.Run("limiter.Get with normal situation for expired", func(t *testing.T) {
		assert := assert.New(t)

		var id = genID()
		limiter := newMemoryLimiter(&RateLimiterConfig{
			LimitConfig:LimiterConfig{
				Max:100,
				Duration:time.Minute,
			},
			Prefix:"Test",
		})
		policy := []int{3, 300, 2, 200}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

	})
}

func genID() string {
	buf := make([]byte, 12)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}


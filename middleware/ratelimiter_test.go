package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"crypto/rand"
	"encoding/hex"
	"time"
	"sync"
)

func TestRateLimiter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rateLimit := RateLimiter()

	h := rateLimit(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// g
	h(c)
	assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "99")
	assert.Contains(t, rec.Header().Get("X-Ratelimit-Limit"), "100")


	//ratelimit with config
	rateLimitWithConfig := RateLimiterWithConfig(RateLimiterConfig{
		Max:2,
	})

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	hx := rateLimitWithConfig(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	hx(c)
	hx(c)
	expectedErrorStatus := hx(c).(*echo.HTTPError)

	assert.Contains(t, rec.Header().Get("X-Ratelimit-Remaining"), "-1")
	assert.Equal(t, http.StatusTooManyRequests, expectedErrorStatus.Code)

}

func TestMemoryRateLimiter(t *testing.T) {
	t.Run("ratelimiter with default Options should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := newMemoryLimiter(&RateLimiterConfig{
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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
			Max:100,
			Duration:time.Minute,
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


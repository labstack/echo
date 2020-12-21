package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	numShards = 15
)

type (
	// RateLimiterShardedMemoryStore is the built-in store implementation for RateLimiter
	RateLimiterShardedMemoryStore struct {
		shards    []*VisitorShard
		rate      rate.Limit
		burst     int
		expiresIn time.Duration
	}
	// VisitorShard is a lockable memory map
	VisitorShard struct {
		visitors  map[string]*Visitor
		mutex     sync.Mutex
		expiresIn time.Duration
		expires   time.Time
	}
)

// NewRateLimiterShardedMemoryStore returns an instance of RateLimiterMemoryStore
func NewRateLimiterShardedMemoryStore(config RateLimiterMemoryStoreConfig) (store *RateLimiterShardedMemoryStore) {
	store = &RateLimiterShardedMemoryStore{}

	store.rate = config.rate
	store.burst = config.burst
	if config.expiresIn == 0 {
		store.expiresIn = DefaultRateLimiterShardedMemoryStoreConfig.expiresIn
	} else {
		store.expiresIn = config.expiresIn
	}
	store.shards = make([]*VisitorShard, numShards)
	for i := 0; i < numShards; i++ {
		store.shards[i] = NewVisitorShard(store.expiresIn)
	}

	return store
}

// RateLimiterShardedMemoryStoreConfig represents configuration for RateLimiterShardedMemoryStore
type RateLimiterShardedMemoryStoreConfig struct {
	rate      rate.Limit
	burst     int
	expiresIn time.Duration
}

// DefaultRateLimiterShardedMemoryStoreConfig provides default configuration values for RateLimiterShardedMemoryStore
var DefaultRateLimiterShardedMemoryStoreConfig = RateLimiterMemoryStoreConfig{
	expiresIn: 3 * time.Minute,
}

// Allow implements RateLimiterStore.Allow
func (store *RateLimiterShardedMemoryStore) Allow(identifier string) bool {

	shardIdx := byte(0)
	for _, b := range []byte(identifier) {
		shardIdx += b
	}
	shardIdx = shardIdx % numShards
	// println(shardIdx)
	shard := store.shards[shardIdx]

	shard.mutex.Lock()
	limiter, exists := shard.visitors[identifier]
	if !exists {
		// Not found in current shard, attempt other shard too
		limiter, exists = shard.visitors[identifier]
		if !exists || now().Sub(limiter.lastSeen) < store.expiresIn {
			// Create a new limiter unless exists and not expired
			limiter = new(Visitor)
			limiter.Limiter = rate.NewLimiter(store.rate, store.burst)
			limiter.lastSeen = now()
		}
		shard.visitors[identifier] = limiter
	}
	limiter.lastSeen = now()

	if now().After(shard.expires) {
		shard.expires = now().Add(shard.expiresIn)
		go shard.Cleanup()
	}

	shard.mutex.Unlock()

	return limiter.AllowN(now(), 1)
}

// NewVisitorShard returns a visitor shard
func NewVisitorShard(expiresIn time.Duration) *VisitorShard {
	v := &VisitorShard{
		visitors:  make(map[string]*Visitor),
		expiresIn: expiresIn,
		expires:   now().Add(expiresIn),
	}
	return v
}

// Cleanup a shards completely
func (shard *VisitorShard) Cleanup() {
	shard.mutex.Lock()
	for id, visitor := range shard.visitors {
		if now().Sub(visitor.lastSeen) > shard.expiresIn {
			delete(shard.visitors, id)
		}
	}
	shard.expires = now().Add(shard.expiresIn)
	shard.mutex.Unlock()
}

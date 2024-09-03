package testing

import (
	"sync"
	"testing"

	"sustainyfacts.dev/anycache/adapters/any_redis"
	"sustainyfacts.dev/anycache/adapters/any_ristretto"
	"sustainyfacts.dev/anycache/cache"
)

var (
	once sync.Once

	redisStore     cache.Store
	ristrettoStore = any_ristretto.NewAdapter()
)

func setup() {
	rs, err := any_redis.NewAdapter("redis://localhost:6379/0?protocol=3")
	if err != nil {
		panic(err)
	}
	redisStore = rs
}

func init() {
	cache.SetDefaultStore(ristrettoStore)
}

func TestRistrettoFirstRedisSecond(t *testing.T) {
	once.Do(setup)
	loader := func(key string) (string, error) {
		return "value for " + key, nil
	}

	group := cache.NewFactory("TestCacheLoader", loader).WithStore(ristrettoStore).WithSecondLevelStore(redisStore).Cache()

	v1, _ := group.Get("key1")
	if v1 != "value for key1" {
		t.Errorf("value for key1 should be 'value for 1', but got %v", v1)
	}

	v2, _ := group.Get("key2")
	if v2 != "value for key2" {
		t.Errorf("value for key2 should be 'value for 2', but got '%v'", v2)
	}

}

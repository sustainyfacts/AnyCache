package any_redis

import (
	"fmt"

	"sustainyfacts.dev/anycache/cache"
)

func TestRedis() {
	rds, _ := NewAdapter("redis://localhost:6379/0?protocol=3")
	cache.SetDefaultStore(rds)

	group := cache.NewFactory("TestRedis",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}

package any_redis

import (
	"fmt"

	"gitlab.com/sustainyfacts/anycache/cache"
)

func TestRistretto() {
	cache.SetDefaultStore(NewAdapter("localhost:6379", ""))

	group := cache.NewFactory("TestRedis",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}

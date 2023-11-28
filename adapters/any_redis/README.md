# Redis Adapter

This is the Redis adapter for AnyCache.

## Usage

```go
import (
	"fmt"

	"sustainyfacts.dev/anycache/adapters/any_redis"
	"sustainyfacts.dev/anycache/cache"
)

func TestRedis() {
	cache.SetDefaultStore(any_redis.NewAdapter("localhost:6379", ""))

	group := cache.NewFactory("TestRedis",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}
```

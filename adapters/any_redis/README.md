# Redis Adapter

This is the Redis adapter for AnyCache.

## Usage

```go
import (
	"fmt"

	"go.sustainyfacts.org/anycache/adapters/any_redis"
	"go.sustainyfacts.org/anycache/cache"
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

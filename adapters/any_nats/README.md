# Ristretto Adapter

This is the NATS adapter for AnyCache, to provide cluster communication.

## Usage

```go
import (
	"fmt"

	"sustainyfacts.dev/anycache/adapters/any_redis"
	"sustainyfacts.dev/anycache/cache"
)

func TestRedis() {
	rda, _ := any_redis.NewAdapter("redis://localhost:6379/0?protocol=3")
	cache.SetDefaultStore(rda)

	group := cache.NewFactory("TestRedis",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}
```

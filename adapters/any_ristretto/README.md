# Ristretto Adapter

This is the ristretto adapter for AnyCache.

## Usage

```go
import (
	"fmt"

	"sustainyfacts.dev/anycache/adapters/any_ristretto"
	"sustainyfacts.dev/anycache/cache"
)

func TestRistretto() {
	cache.SetDefaultStore(any_ristretto.NewAdapter())

	group := cache.NewFactory("TestRistretto",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}
```

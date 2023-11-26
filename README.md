# AnyCache 

> 🚧 Work in Progress
> 
> This library is a working in progress, and as a result even the public API will probably change

Simple Cache wrapper that allows selecting the cache implementation without changing your code.

It is inspired by [Gocache](https://github.com/eko/gocache) and [GroupCache](https://github.com/golang/groupcache)

Requires Go 1.21 or newer.

## Features

* ✅ __No external dependencies__: besides the store implementation of your choice, there are no depencies to 3rd party libraries
* ✅ __Type-safe, loadable cache__: you defined a cacheLoader function to load your data. Because AnyCache is using generics, you can use your actual types instead of `any`
* ✅ __Cache groups__: several groups using a single underlying store for optimal performance and memory usage
* ✅ __Configurable cache stores__: in-memory, redis, or your own custom store
* 🚧 Second level cache: use a primary memory store with a fallback to a redis shared cache for instance 
* 🚧 A marshaler to automatically marshal/unmarshal your cache values as a struct 
* 🚧 Cache invalidation by expiration time
* ✅ __Distributed invalidation__: inject a message broker to enable distributed invalidation of the in-memory caches in your cluster


## Built-in adapters

* [Memory (ristretto)](https://github.com/dgraph-io/ristretto) (dgraph-io/ristretto)

## Usage

### Simple initialization

```go
import (
	"fmt"

	"go.sustainyfacts.org/anycache/adapters/store_ristretto"
	"go.sustainyfacts.org/anycache/cache"
)

func TestRistretto() {
	cache.SetDefaultStore(store_ristretto.NewStore())

	group := cache.NewFactory("TestRistretto",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}
```


## Benchmarks

Todo

## License

AnyCache is released under the Apache 2.0 license (see [LICENSE](LICENSE))
# AnyCache 

> 🚧 Work in Progress
> 
> This library is a working in progress, and as a result even the public API will probably change

Simple Cache wrapper that allows selecting the cache implementation without changing your code.

It is inspired by [Gocache](https://github.com/eko/gocache) and [GroupCache](https://github.com/golang/groupcache)

Requires Go 1.21 or newer.

## Features

* ✅ __No external dependencies__: all dependencies to 3rd party libraries are in the adapters provided, so you can choose what code you will deploy into your project.
* ✅ __Type-safe, loadable cache__: uses a cacheLoader function to load your data into the cache. Because AnyCache is using generics, you can use your actual types instead of `any`.
* ✅ __Cache groups__: several groups using a single underlying store for optimal performance and memory usage.
* ✅ __Configurable cache stores__: in-memory, redis, or your own custom store.
* 🚧 __Second level store__: back your in-memory store by a redis instance, so that you cache survives deployment of a new version of your application.
* 🚧 Cache invalidation by expiration time
* ✅ __Distributed invalidation__: inject a message broker to enable distributed invalidation of the in-memory caches in your cluster
* 🚧 __Prometheus metrics__: provides metrics, for each group, globally, and for first and second level separately


## Built-in adapters

* [Memory (ristretto)](adapters/any_ristretto/README.md) (dgraph-io/ristretto)

## Usage

### Simple initialization

```go
import (
	"fmt"

	"go.sustainyfacts.org/anycache/cache"
)

func SimpleExample() {
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
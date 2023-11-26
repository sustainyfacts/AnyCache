package examples

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

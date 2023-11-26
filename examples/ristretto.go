package examples

import (
	"fmt"

	"gitlab.com/sustainyfacts/anycache/adapters/store_ristretto"
	"gitlab.com/sustainyfacts/anycache/cache"
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

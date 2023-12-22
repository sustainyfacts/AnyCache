/*
Copyright © 2023 The Authors (See AUTHORS file)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package examples

import (
	"fmt"

	"sustainyfacts.dev/anycache/adapters/any_redis"
	"sustainyfacts.dev/anycache/adapters/any_ristretto"
	"sustainyfacts.dev/anycache/cache"
)

func TestDistributed() {
	cache.SetDefaultStore(any_ristretto.NewAdapter())
	rda, _ := any_redis.NewAdapterWithMessaging("redis://localhost:6379/0?protocol=3", "cache.flush.topic")

	// Use the same group name on each node
	group := cache.NewFactory("TestDistributed",
		func(key string) (string, error) {
			return "value for " + key, nil
		}).WithSecondLevelStore(rda).WithBroker(rda).Cache()

	v, _ := group.Get("my-unique-key")
	fmt.Println(v)
}

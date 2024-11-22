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
package tests

import (
	"sync"
	"testing"

	"sustainyfacts.dev/anycache/adapters/any_redis"
	"sustainyfacts.dev/anycache/adapters/any_ristretto"
	"sustainyfacts.dev/anycache/cache"
)

var (
	once sync.Once

	redisStore     cache.Store
	ristrettoStore = any_ristretto.NewAdapter()
)

func setup() {
	rs, err := any_redis.NewAdapter("redis://localhost:6379/0?protocol=3")
	if err != nil {
		panic(err)
	}
	redisStore = rs
}

func init() {
	cache.SetDefaultStore(ristrettoStore)
}

func TestRistrettoFirstRedisSecond(t *testing.T) {
	once.Do(setup)
	loader := func(key string) (string, error) {
		return "value for " + key, nil
	}

	group := cache.NewFactory("TestCacheLoader", loader).WithStore(ristrettoStore).WithSecondLevelStore(redisStore).Cache()

	v1, _ := group.Get("key1")
	if v1 != "value for key1" {
		t.Errorf("value for key1 should be 'value for 1', but got %v", v1)
	}

	v2, _ := group.Get("key2")
	if v2 != "value for key2" {
		t.Errorf("value for key2 should be 'value for 2', but got '%v'", v2)
	}

}

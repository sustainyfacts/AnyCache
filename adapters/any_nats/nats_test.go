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
package any_redis

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sustainyfacts.dev/anycache/cache"
)

// Those tests requires a redis instance running on localhost:6379.
// Use the docker compose to start the instance.

var (
	testTTL    = 3 * time.Second // Otherwise the test cannot run twice
	once       sync.Once
	natsBroker cache.MessageBroker
)

func setup() {
	b, err := NewAdapter("nats://localhost:4222", "cache.flush.topic")
	if err != nil {
		panic(err)
	}
	natsBroker = b
}

func TestCacheLoader(t *testing.T) {
	once.Do(setup)

	var message string

	natsBroker.Subscribe(func(msg []byte) {
		message = string(msg)
	})

	natsBroker.Send([]byte("Hello"))

	assert.Equal(t, "hello", message, "Message received")
}

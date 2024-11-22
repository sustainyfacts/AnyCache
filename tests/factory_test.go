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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"sustainyfacts.dev/anycache/cache"
)

// Test that UUID can be used as a key
func Test_UUID(t *testing.T) {
	const nbElements = 100
	cacheLoads := 0
	group := cache.NewFactory("TestUUID",
		func(key uuid.UUID) (string, error) {
			cacheLoads++
			return key.String(), nil
		}).Cache()

	var ids []uuid.UUID
	for i := 0; i < nbElements; i++ {
		id, _ := uuid.NewRandom()
		ids = append(ids, id)
	}
	for i := 0; i < 2; i++ { // Do it twice to make sure to hit the cache
		for _, id := range ids {
			fromCache, err := group.Get(id)
			assert.NoError(t, err)
			assert.Equal(t, id.String(), fromCache)
		}
	}
	assert.Equal(t, nbElements, cacheLoads, "second lookup should be from cache")
}

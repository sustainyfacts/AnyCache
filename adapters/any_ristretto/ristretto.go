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
package any_ristretto

import (
	"math/rand"

	"github.com/dgraph-io/ristretto"
	"github.com/dgraph-io/ristretto/z"
	"sustainyfacts.dev/anycache/cache"
)

func NewAdapter() cache.Store {
	return NewAdapterWithConfig(&ristretto.Config{
		NumCounters: 1e7,       // number of keys to track frequency of (10M).
		MaxCost:     1 << 30,   // maximum cost of cache (1GB).
		BufferItems: 64,        // number of keys per Get buffer.
		KeyToHash:   keyToHash, // allows sharing a single cache between different groups
	})
}

func NewAdapterWithConfig(config *ristretto.Config) cache.Store {
	r, _ := ristretto.NewCache(config)

	return &store{store: r, groupHashes: make(map[string]groupHash),
		groupConfigs: make(map[string]cache.GroupConfig)}
}

type store struct {
	store        *ristretto.Cache
	groupHashes  map[string]groupHash
	groupConfigs map[string]cache.GroupConfig
}

func (s *store) ConfigureGroup(name string, config cache.GroupConfig) {
	if config.Ttl != 0 {
		panic("stores does not support TTL")
	} else if config.Cost != 0 {
		panic("stores does not support Cost")
	}
	h1, h2 := z.KeyToHash(name)
	s.groupHashes[name] = groupHash{h1: h1, h2: h2}
	s.groupConfigs[name] = config
}

func (s *store) Get(key cache.GroupKey) (any, error) {
	if v, ok := s.store.Get(key.StoreKey); ok {
		return v, nil
	}
	return nil, cache.ErrKeyNotFound
}

func (s *store) Set(key cache.GroupKey, value any) error {
	s.store.Set(key.StoreKey, value, 0)
	return nil // dropped values (above return false) are not errors
}

func (s *store) Del(key cache.GroupKey) error {
	s.store.Del(key.StoreKey)
	return nil
}

// Note that this does not free memory, but rotates the hashes
// so querying the same key will require a call to the cache loader function
func (s *store) Clear(groupName string) {
	gh := s.groupHashes[groupName]
	// This replaces the groupHash, which means that the hashes of all the keys will be changed
	newGroupHash := groupHash{h1: gh.h1*hashMultiplierValue + rand.Uint64(), h2: gh.h2*hashMultiplierValue + rand.Uint64()}
	s.groupHashes[groupName] = newGroupHash
}

func (s *store) Key(groupName string, key any) cache.GroupKey {
	gh := s.groupHashes[groupName]
	return cache.GroupKey{GroupName: groupName, StoreKey: groupKey{hash: &gh, key: key}}
}

const (
	hashMultiplierValue = 37 // See apache-commons HashCodeBuilder
)

// Hash values corresponding to the group name, kept together for convenience
type groupHash struct {
	h1, h2 uint64
}

// Composite key with the hash values of the name of the group
// The purpose of this composite key is to have unique keys within
// a group
type groupKey struct {
	key  any
	hash *groupHash
}

// Hashing function
func keyToHash(key interface{}) (uint64, uint64) {
	if key == nil {
		panic("key cannot be nil")
	}
	groupKey, ok := key.(groupKey)
	if !ok {
		panic("unexpected key type")
	}
	h1, h2 := z.KeyToHash(groupKey.key)
	return h1*hashMultiplierValue + groupKey.hash.h1, h2*hashMultiplierValue + groupKey.hash.h2
}

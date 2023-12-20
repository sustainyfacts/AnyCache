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
package cache

import (
	"reflect"
	"regexp"
	"time"

	"sustainyfacts.dev/anycache/cache/singleflight"
)

var (
	nameRegex = regexp.MustCompile("[a-zA-Z0-9_-]+")
)

type Factory[K int64 | string | uint64, V any] struct {
	Name                     string
	CacheLoader              func(key K) (V, error) // Loader in case of cache miss
	LoadDuplicateSuppression bool                   // To avoid multiple concurrent loads for the same entry
	MessageBroker            MessageBroker          // Message broker for distributed cache flush messages
	Store                    Store                  // Store to cache entries
	Ttl                      time.Duration          // Time to live for a cache entry
	allowDuplicates          bool                   // Allow duplicate names for testing of distributed functionality
	debug                    bool                   // Prints debug information
}

func (f Factory[K, V]) Cache() *Group[K, V] {
	if !nameRegex.MatchString(f.Name) {
		panic("allowed characters in the name are: [a-zA-Z0-9_-]")
	}

	if f.CacheLoader == nil {
		panic("no CacheLoader defined")
	}
	// Using default store unless another one is specified
	store := defaultStore
	if f.Store != nil {
		store = f.Store
	}
	if store == nil {
		panic("no default store set and no store provided in factory")
	}
	if stores, ok := allGroups[f.Name]; ok {
		if !f.allowDuplicates {
			panic("cannot create two groups with the same name")
		}
		for _, s := range stores {
			if &s == &store {
				panic("cannot create two groups with the same name for a given store")
			}
		}
	}
	allGroups[f.Name] = append(allGroups[f.Name], store)

	group := Group[K, V]{store: store, name: f.Name,
		load: f.CacheLoader, messageBroker: f.MessageBroker,
		debug: f.debug}
	if f.LoadDuplicateSuppression {
		group.loadGroup = &singleflight.Group[K, V]{}
	}

	// Configure the group for the store
	config := GroupConfig{Ttl: f.Ttl, Cost: 0, ValueType: reflect.TypeOf(*new(V))}
	store.ConfigureGroup(f.Name, config)

	if group.messageBroker != nil {
		group.messageBroker.Subscribe(group.handleMessage)
	}

	return &group
}

// Convenience method to inject the cache into other libraries as a function decorator
//
// Example:
//
//	decorator := cache.NewDecorator[string, string]("secrets")
//	var fetchSecret func(param string) string = ...
//	getCachedSecret := decorator.Decorate(fetchFunction)
//	value := getCachedSecret("topsecret")
func (f Factory[K, V]) Decorate(cacheLoader func(key K) (V, error)) func(key K) (V, error) {
	f.CacheLoader = cacheLoader
	return f.Cache().Get
}

func (f Factory[K, V]) WithLoadDuplicateSuppression() Factory[K, V] {
	f.LoadDuplicateSuppression = true
	return f
}

func (f Factory[K, V]) WithStore(s Store) Factory[K, V] {
	f.Store = s
	return f
}

func (f Factory[K, V]) WithDebug() Factory[K, V] {
	f.debug = true
	return f
}

func (f Factory[K, V]) WithTTL(ttl time.Duration) Factory[K, V] {
	f.Ttl = ttl
	return f
}

// Note: if you use the a broker for different cache groups, make sure that
// the different groups are using different topics, so they do not receive
// each others messages.
func (f Factory[K, V]) WithBroker(broker MessageBroker) Factory[K, V] {
	f.MessageBroker = broker
	return f
}

// Convenience methods to inject the cache into other libraries
func NewFactory[K int64 | string | uint64, V any](name string, cacheLoader func(key K) (V, error)) Factory[K, V] {
	return Factory[K, V]{Name: name, CacheLoader: cacheLoader}
}

func NewDecorator[K int64 | string | uint64, V any](name string) Factory[K, V] {
	return Factory[K, V]{Name: name}
}

// Note: if you use the a broker for different cache groups, make sure that
// the different groups are using different topics, so they do not receive
// each others messages.
func (f Factory[K, V]) withDuplicates() Factory[K, V] {
	f.allowDuplicates = true
	return f
}

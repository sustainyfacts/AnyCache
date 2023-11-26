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
	"go.sustainyfacts.org/anycache/cache/singleflight"
)

type Factory[K int64 | string | uint64, V any] struct {
	Name                     string
	CacheLoader              func(key K) (V, error)
	LoadDuplicateSuppression bool
	MessageBroker            MessageBroker
	Store                    Store
}

func (f Factory[K, V]) Cache() *Group[K, V] {
	if _, ok := allGroups[f.Name]; ok {
		panic("cannot create two groups with the same name")
	}
	if f.CacheLoader == nil {
		panic("no Cache Loader defined")
	}
	allGroups[f.Name] = true

	// Using default store unless another one is specified
	store := defaultStore
	if f.Store != nil {
		store = f.Store
	}
	if store == nil {
		panic("no default store set and no store provided in factory")
	}
	group := Group[K, V]{store: store, name: f.Name, load: f.CacheLoader, messageBroker: f.MessageBroker}
	if f.LoadDuplicateSuppression {
		group.loadGroup = &singleflight.Group[K, V]{}
	}
	if group.messageBroker != nil {
		group.messageBroker.Subscribe(group.handleMessage)
	}

	// Configure the group for the store
	config := GroupConfig{Ttl: 0, Cost: 0}
	store.ConfigureGroup(f.Name, config)

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

// MessageBroker is an interface that can be used to provide clustered communication
// to the cache, for sending a receiving Flush and DeleteKey messages
type MessageBroker interface {
	Send([]byte)                // Sends a message to all other caches
	Subscribe(func(msg []byte)) // Subcribe to messages from another caches
}

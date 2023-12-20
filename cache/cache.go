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
	"log"
)

var (
	defaultStore Store = NewHashMapStore()    // Default underlying cache implementation
	allGroups          = map[string][]Store{} // To avoid instanciating the same group twice for the same store
)

// Sets the default store for all the next groups to be created
func SetDefaultStore(store Store) {
	defaultStore = store
}

type Group[K comparable, V any] struct {
	store Store // The underlying cache engine
	name  string
	load  func(key K) (V, error)

	// loadGroup ensures that each key is only fetched once
	// (either locally or remotely), regardless of the number of
	// concurrent callers.
	loadGroup flightGroup[K, V]

	// messageBroker is used for clustered events like flushing of entries
	messageBroker MessageBroker

	// debug enabled
	debug bool
}

// flightGroup is defined as an interface which flightgroup.Group
// satisfies.  We define this so that we may test with an alternate
// implementation.
type flightGroup[K comparable, V any] interface {
	Do(key K, fn func() (V, error)) (V, error)
}

func (g *Group[K, V]) Get(key K) (V, error) {
	gk := g.store.Key(g.name, key)
	if v, err := g.store.Get(gk); err == nil || err != ErrKeyNotFound {
		return v.(V), err
	}

	loadAndSet := func() (V, error) {
		g.log("loading key %v", key)
		// Not found in cache, using loader
		v, err := g.load(key)

		if err != nil {
			return v, err
		}

		// Set the value
		g.store.Set(gk, v)
		return v, nil
	}

	if g.loadGroup != nil {
		return g.loadGroup.Do(key, loadAndSet)
	}
	return loadAndSet()
}

func (g *Group[K, V]) Del(key K) {
	g.delNoFlush(key)
	if g.messageBroker != nil {
		msg := cacheMsg{Group: g.name, Key: key}
		go g.messageBroker.Send(msg.bytes()) // async call
	}
}

func (g *Group[K, V]) delNoFlush(key K) {
	g.log("delete key %v", key)
	gk := g.store.Key(g.name, key)
	g.store.Del(gk)
}

func (g *Group[K, V]) log(message string, args ...any) {
	if g.debug {
		log.Printf("group(%s): "+message, g.name, args)
	}
}

func (g *Group[K, V]) warn(message string, args ...any) {
	log.Printf("Warn - group(%s): "+message, g.name, args)
}

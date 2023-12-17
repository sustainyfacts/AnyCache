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
	"io"
	"testing"
	"time"
)

// Tests that a group can be flushed
func TestDistributedFlush(t *testing.T) {
	counter := 0
	loader := func(key string) (int, error) {
		counter++
		return counter, nil
	}
	broker := newSimpleBroker()
	// We use two stores to simulate two nodes with separate in-memory stores
	group1 := NewFactory("dist-flush", loader).WithBroker(broker).WithStore(NewHashMapStore()).Cache()
	group2 := NewFactory("dist-flush", loader).WithBroker(broker).WithStore(NewHashMapStore()).withDuplicates().Cache()

	v, _ := group1.Get("key")
	if v != 1 {
		t.Errorf("group1 key lookup should be 1, but got %v", v)
	}

	v, _ = group2.Get("key")
	if v != 2 { // This is not the same store so counter should increase
		t.Errorf("group1 key lookup should be 2, but got %v", v)
	}

	group1.Del("key")

	time.Sleep(10 * time.Millisecond) // Wait the cache flush message has been propagated

	v, _ = group1.Get("key")
	if v != 3 { // Count is increased by new call to loader
		t.Errorf("group1 key lookup after flush should be 3, but got %v", v)
	}
	v, _ = group2.Get("key")
	if v != 4 { // Count is increased by new call to loader as well
		t.Errorf("group2 key lookup after flush should still be 2, but got %v", v)
	}
}

type simpleBroker struct {
	subscribers []func(message []byte)
}

func newSimpleBroker() *simpleBroker {
	return &simpleBroker{}
}
func (b *simpleBroker) Send(message []byte) error {
	for _, subscriber := range b.subscribers {
		// Handle in a go-routine
		go subscriber(message)
	}
	return nil
}

func (b *simpleBroker) Subscribe(handler func(message []byte)) (io.Closer, error) {
	b.subscribers = append(b.subscribers, handler)
	var cf CloserFunc = func() error {
		for i, h := range b.subscribers {
			if &h == &handler {
				b.subscribers = remove(b.subscribers, i)
				return nil
			}
		}
		panic("subscriber not found")
	}
	return cf, nil
}

func remove[V any](s []V, i int) []V {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// To be able to return an anonymous function in Subscribe()
type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}

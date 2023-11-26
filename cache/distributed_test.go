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
	group1 := NewFactory("group1-dist-flush", loader).WithBroker(broker).Cache()
	group2 := NewFactory("group2-dist-flush", loader).WithBroker(broker).Cache()

	v, _ := group1.Get("key")
	if v != 1 {
		t.Errorf("group1 key lookup should be 1, but got %v", v)
	}

	v, _ = group2.Get("key")
	if v != 2 {
		t.Errorf("group1 key lookup should be 2, but got %v", v)
	}

	group1.Del("key")

	time.Sleep(50 * time.Millisecond) // Wait until flush is done

	v, _ = group1.Get("key")
	if v != 3 {
		t.Errorf("group1 key lookup after flush should be 3, but got %v", v) // Count is increased by new call to loader
	}
	v, _ = group2.Get("key")
	if v != 2 {
		t.Errorf("group2 key lookup after flush should still be 2, but got %v", v) // Count is increased by new call to loader
	}
}

type simpleBroker struct {
	subscribers []func(message []byte)
}

func newSimpleBroker() *simpleBroker {
	return &simpleBroker{}
}
func (b *simpleBroker) Send(message []byte) {
	for _, subscriber := range b.subscribers {
		// Handle in a go-routine
		go subscriber(message)
	}
}

func (b *simpleBroker) Subscribe(handler func(message []byte)) {
	b.subscribers = append(b.subscribers, handler)
}

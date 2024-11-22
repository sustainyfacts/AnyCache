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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sustainyfacts.dev/anycache/cache"
)

func TestCacheLoader(t *testing.T) {
	loader := func(key string) (string, error) {
		return "value for " + key, nil
	}

	group := cache.NewFactory("TestCacheLoader", loader).Cache()

	v1, _ := group.Get("key1")
	if v1 != "value for key1" {
		t.Errorf("value for key1 should be 'value for 1', but got %v", v1)
	}

	//cache.Wait() // Wait until it is in cache and get another key

	v2, _ := group.Get("key2")
	if v2 != "value for key2" {
		t.Errorf("value for key2 should be 'value for 2', but got '%v'", v2)
	}
}

func TestCacheLoaderNotFound(t *testing.T) {
	loader := func(key int64) (string, error) {
		if key%2 == 0 {
			return "", fmt.Errorf("Key not found")
		}
		return fmt.Sprintf("value for %d", key), nil
	}

	group := cache.NewFactory("TestCacheLoaderNotFound", loader).Cache()

	v1, _ := group.Get(1)
	if v1 != "value for 1" {
		t.Errorf("value for key 1 should be 'value for 1', but got %v", v1)
	}

	_, err := group.Get(2)
	if err == nil {
		t.Errorf("Key 1 should not have been found")
	}
}

func TestMultipleLoads(t *testing.T) {
	counter := 0
	loader := func(key string) (string, error) {
		counter++
		return fmt.Sprintf("value %d", counter), nil
	}

	group := cache.NewFactory("TestMultipleLoads", loader).Cache()

	v, _ := group.Get("key")
	if v != "value 1" {
		t.Errorf("Incorrect value for key: '%v', but expected 'value 1'", v)
	}

	v, _ = group.Get("key")
	if v != "value 1" {
		t.Errorf("Incorrect value for key: '%v', but expected 'value 1'", v)
	}

	group.Del("key")

	v, _ = group.Get("key") // Value is reloaded after cache is cleared
	if v != "value 2" {
		t.Errorf("Incorrect value for key: '%v', but expected 'value 2'", v)
	}
}

func TestMultipleGroups(t *testing.T) {
	loader1 := func(key string) (string, error) {
		return "1 - value for " + key, nil
	}
	loader2 := func(key string) (string, error) {
		return "2 - value for " + key, nil
	}

	group1 := cache.NewFactory("group1", loader1).Cache()
	group2 := cache.NewFactory("group2", loader2).Cache()

	v1, _ := group1.Get("key")
	if v1 != "1 - value for key" {
		t.Errorf("value for key should be '1 - value for key', but got %v", v1)
	}

	v2, _ := group2.Get("key")
	if v2 != "2 - value for key" {
		t.Errorf("value for key should be '2 - value for key', but got %v", v1)
	}
}

// This test makes sure that concurrent loads do not lead to multiple calls
// to the cacheLoader function
func TestConcurrentLoads(t *testing.T) {
	counter := 0
	loader := func(key string) (string, error) {
		counter++
		time.Sleep(100 * time.Millisecond) // Make sure the load is slow
		return fmt.Sprintf("value for %s", key), nil
	}
	group := cache.NewFactory("TestConcurrentLoads", loader).Cache()

	getAndWait(group, 2, t)

	// Only one load should have occured
	if counter != 2 {
		t.Errorf("CacheLoader should be called twice but got %v", counter)
	}
}

func getAndWait(group *cache.Group[string, string], concurrentRoutines int, t *testing.T) {
	start := make(chan string) // Coordination of start
	responseChannel := make(chan string, concurrentRoutines)
	defer close(responseChannel)

	// Multiple concurrent get
	for i := 0; i < concurrentRoutines; i++ {
		go func(val int) {
			<-start
			v, _ := group.Get("theKey")
			responseChannel <- v
		}(i)
	}

	close(start) // Signal routines to start

	// Wait until everyone is done
	for i := 0; i < concurrentRoutines; i++ {
		select {
		case v := <-responseChannel:
			if v != "value for theKey" {
				t.Errorf("Read incorrect value: '%s', expected 'theKey'", v)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("timeout waiting on getter #%d of %d", i+1, concurrentRoutines)
		}
	}
}

func TestDuplicateSuppression(t *testing.T) {
	counter := 0
	loader := func(key string) (string, error) {
		counter++
		time.Sleep(100 * time.Millisecond) // Make sure the load is slow
		return fmt.Sprintf("value for %s", key), nil
	}
	group := cache.NewFactory("TestDuplicateSuppression", loader).WithLoadDuplicateSuppression().Cache()

	getAndWait(group, 3, t)

	// Only one load should have occured
	if counter != 1 {
		t.Errorf("CacheLoader should be called once but got %v", counter)
	}
}

// Tests that a group can be flushed
func TestFlush(t *testing.T) {
	counter := 0
	loader := func(key string) (int, error) {
		counter++
		return counter, nil
	}

	group1 := cache.NewFactory("group1-flush", loader).Cache()
	group2 := cache.NewFactory("group2-flush", loader).Cache()

	v, _ := group1.Get("key")
	if v != 1 {
		t.Errorf("group1 key lookup should be 1, but got %v", counter)
	}

	v, _ = group2.Get("key")
	if v != 2 {
		t.Errorf("group2 key lookup should be 2, but got %v", counter)
	}

	group1.Del("key")

	v, _ = group1.Get("key")
	if v != 3 {
		t.Errorf("group1 key lookup after flush should be 3, but got %v", v) // Count is increased by new call to loader
	}

	v, _ = group2.Get("key")
	if v != 2 {
		t.Errorf("group2 key lookup after flush should be 2, but got %v", v) // Count unchanged, cached value returned
	}
}

func TestPanicLoad(t *testing.T) {
	counter := 0
	loader := func(key int64) (string, error) {
		counter++
		if counter == 1 {
			panic("first time panic")
		}
		return "nopanic", nil
	}

	group := cache.NewFactory("TestPanicLoad", loader).WithLoadDuplicateSuppression().Cache()

	panicHandler(group)

	v, _ := group.Get(2)
	if v != "nopanic" {
		t.Errorf("second time's the charm, but got %v", v)
	}
}

func panicHandler(group *cache.Group[int64, string]) {
	defer func() {
		// do not let the panic below leak to the test
		_ = recover()
	}()
	group.Get(1)
}

func TestWithReloadOnDelete(t *testing.T) {
	counter := 0
	loader := func(key string) (int, error) {
		counter++
		return counter, nil
	}

	group := cache.NewFactory("WithReloadOnDelete ", loader).WithReloadOnDelete().Cache()
	v, _ := group.Get("key")
	assert.Equal(t, 1, v, "incorrect value for 'key'")
	assert.Equal(t, 1, counter, "loader called once")

	group.Del("key") // Delete / Flush

	assert.Equal(t, 2, counter, "loader called on reload")
	v, _ = group.Get("key")
	assert.Equal(t, 2, counter, "loader not called again on GET")
	assert.Equal(t, 2, v, "incorrect value for 'key'")
}

// Test that a slow loader does not block for Get when using "ReloadOnDelete"
func TestWithReloadNonBlocking(t *testing.T) {
	block := make(chan bool, 2) // Block loader
	counter := 0
	loader := func(key string) (int, error) {
		<-block // Block
		counter++
		return counter, nil
	}
	group := cache.NewFactory("TestWithReloadNonBlocking ", loader).WithLoadDuplicateSuppression().WithReloadOnDelete().Cache()
	block <- true // Do not block first load
	v, _ := group.Get("key")
	assert.Equal(t, 1, v, "incorrect value for 'key'")
	assert.Equal(t, 1, counter, "loader called once")

	go group.Del("key") // Delete / Flush (reload blocked, so run in a goroutine)

	assert.Equal(t, 1, counter, "reload not complete")
	v, _ = group.Get("key") // We can still get values from the cache (old value)
	assert.Equal(t, 1, counter, "loader not called yet, getting old value")
	assert.Equal(t, 1, v, "incorrect value for 'key'")

	close(block)                      // Unblock loader
	time.Sleep(10 * time.Millisecond) // Wait a bit

	assert.Equal(t, 2, counter, "reload completed")
	v, _ = group.Get("key")
	assert.Equal(t, 2, counter, "reload done, getting new value")
	assert.Equal(t, 2, v, "incorrect value for 'key'")
}

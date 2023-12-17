package any_redis

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"sustainyfacts.dev/anycache/cache"
)

// Those tests requires a redis instance running on localhost:6379.
// Use the docker compose to start the instance.

var (
	testTTL = 3 * time.Second // Otherwise the test cannot run twice
	once    sync.Once
)

func setup() {
	redisStore, err := NewAdapter("redis://localhost:6379/0?protocol=3")
	if err != nil {
		panic(err)
	}
	cache.SetDefaultStore(redisStore)
}

func TestCacheLoader(t *testing.T) {
	once.Do(setup)

	loader := func(key string) (string, error) {
		return "value for " + key, nil
	}

	group := cache.NewFactory("TestCacheLoader", loader).WithTTL(testTTL).Cache()

	v1, _ := group.Get("key1")
	if v1 != "value for key1" {
		t.Errorf("value for key1 should be 'value for 1', but got %v", v1)
	}

	v2, _ := group.Get("key2")
	if v2 != "value for key2" {
		t.Errorf("value for key2 should be 'value for 2', but got '%v'", v2)
	}
}

func TestCacheLoaderNotFound(t *testing.T) {
	once.Do(setup)

	loader := func(key int64) (string, error) {
		if key%2 == 0 {
			return "", fmt.Errorf("Key not found")
		}
		return fmt.Sprintf("value for %d", key), nil
	}

	group := cache.NewFactory("TestCacheLoaderNotFound", loader).WithTTL(testTTL).Cache()

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
	once.Do(setup)

	counter := 0
	loader := func(key string) (string, error) {
		counter++
		return fmt.Sprintf("value %d", counter), nil
	}

	group := cache.NewFactory("TestMultipleLoads", loader).WithTTL(testTTL).Cache()

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
	once.Do(setup)

	loader1 := func(key string) (string, error) {
		return "1 - value for " + key, nil
	}
	loader2 := func(key string) (string, error) {
		return "2 - value for " + key, nil
	}

	group1 := cache.NewFactory("group1", loader1).WithTTL(testTTL).Cache()
	group2 := cache.NewFactory("group2", loader2).WithTTL(testTTL).Cache()

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
	once.Do(setup)

	counter := 0
	loader := func(key string) (string, error) {
		counter++
		time.Sleep(100 * time.Millisecond) // Make sure the load is slow
		return fmt.Sprintf("value for %s", key), nil
	}
	group := cache.NewFactory("TestConcurrentLoads", loader).WithTTL(testTTL).Cache()

	getAndWait(group, t)

	// Only one load should have occured
	if counter != 2 {
		t.Errorf("CacheLoader should be called twice but got %v", counter)
	}
}

func getAndWait(group *cache.Group[string, string], t *testing.T) {
	start := make(chan string) // Coordination of start
	responseChannel := make(chan string, 2)
	defer close(responseChannel)

	// Multiple concurrent get
	for i := 0; i < 2; i++ {
		go func(val int) {
			<-start
			v, _ := group.Get("theKey")
			responseChannel <- v
		}(i)
	}

	close(start) // Signal routines to start

	// Wait until everyone is done
	for i := 0; i < 2; i++ {
		select {
		case v := <-responseChannel:
			if v != "value for theKey" {
				t.Errorf("Read incorrect value: '%s', expected 'theKey'", v)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("timeout waiting on getter #%d of 2", i+1)
		}
	}
}

func TestDuplicateSuppression(t *testing.T) {
	once.Do(setup)

	counter := 0
	loader := func(key string) (string, error) {
		counter++
		time.Sleep(100 * time.Millisecond) // Make sure the load is slow
		return fmt.Sprintf("value for %s", key), nil
	}
	group := cache.NewFactory("TestDuplicateSuppression", loader).WithLoadDuplicateSuppression().WithTTL(testTTL).Cache()

	getAndWait(group, t)

	// Only one load should have occured
	if counter != 1 {
		t.Errorf("CacheLoader should be called once but got %v", counter)
	}
}

// Tests that a group can be flushed
func TestFlush(t *testing.T) {
	once.Do(setup)

	counter := 0
	loader := func(key string) (int, error) {
		counter++
		return counter, nil
	}

	group1 := cache.NewFactory("group1-flush", loader).WithTTL(testTTL).Cache()
	group2 := cache.NewFactory("group2-flush", loader).WithTTL(testTTL).Cache()

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
	once.Do(setup)

	counter := 0
	loader := func(key int64) (string, error) {
		counter++
		if counter == 1 {
			panic("first time panic")
		}
		return "nopanic", nil
	}

	group := cache.NewFactory("TestPanicLoad", loader).WithLoadDuplicateSuppression().WithTTL(testTTL).Cache()

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

func TestDistributedFlush(t *testing.T) {
	broker, err := NewAdapterWithMessaging("redis://localhost:6379/0?protocol=3", "cache.flush.topic")
	if err != nil {
		panic(err)
	}
	counter := 0
	loader := func(key string) (int, error) {
		counter++
		return counter, nil
	}
	store1 := cache.NewHashMapStore()
	store2 := cache.NewHashMapStore()
	group1 := cache.NewFactory("dist-flush", loader).WithBroker(broker).WithStore(store1).Cache()
	group2 := cache.NewFactory("dist-flush", loader).WithBroker(broker).WithStore(store2).Cache()

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

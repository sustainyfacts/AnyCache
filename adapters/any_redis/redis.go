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
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"sustainyfacts.dev/anycache/cache"
)

var ctx = context.Background()

// Creates a new adapter for Redis, and checks for its availability
// using the PING command and retrieves the server version. Uses the provided
// topic so it can be used for cluster communication (distributed cache flush)
func NewAdapterWithMessaging(url string, topic string) (cache.BrokerStore, error) {
	return newAdapter(url, topic)
}

// Creates a new adapter for Redis, and checks for its availability
// using the PING command and retrieves the server version.
func NewAdapter(url string) (cache.Store, error) {
	return newAdapter(url, "")
}

// Creates a new adapter for Redis with redis.Client given as parameter
func NewAdapterWithClient(rdb *redis.Client, topic string) (cache.Store, error) {
	return &adapter{rdb: rdb, groupConfigs: make(map[string]cache.GroupConfig), topic: topic}, nil
}

func newAdapter(url string, topic string) (*adapter, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	if pong != "PONG" {
		return nil, fmt.Errorf("invalid ping response: %s", pong)
	}
	info, err := rdb.Info(ctx, "server").Result()
	if err != nil {
		return nil, err
	}
	serverVersion := "unkown"
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "redis_version:") {
			serverVersion = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}
	log.Printf("Connected to Redis: %s", serverVersion)
	return &adapter{rdb: rdb, groupConfigs: make(map[string]cache.GroupConfig), topic: topic}, nil
}

type adapter struct {
	rdb          *redis.Client
	topic        string // For messaging
	groupConfigs map[string]cache.GroupConfig
}

func (a *adapter) ConfigureGroup(name string, config cache.GroupConfig) {
	if config.Cost != 0 {
		panic("Redis does not support Cost")
	}
	a.groupConfigs[name] = config
}

func (a *adapter) Get(key cache.GroupKey) (any, error) {

	v, err := a.rdb.Get(ctx, key.StoreKey.(string)).Result()
	if err == redis.Nil {
		return nil, cache.ErrKeyNotFound
	} else if err != nil {
		return nil, err
	}

	expectedType := a.groupConfigs[key.GroupName].ValueType
	switch expectedType.Kind() {
	case reflect.Int:
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		return i, nil
	case reflect.Int32:
		i, err := strconv.ParseInt(v, 10, expectedType.Bits())
		if err != nil {
			panic(err)
		}
		return int32(i), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(v, 10, expectedType.Bits())
		if err != nil {
			panic(err)
		}
		return int32(i), nil
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(v, expectedType.Bits())
		if err != nil {
			panic(err)
		}
		return i, nil
	case reflect.String:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type: %v", expectedType)
	}
}

// new test
func (a *adapter) Set(key cache.GroupKey, value any) error {
	ttl := a.groupConfigs[key.GroupName].Ttl
	return a.rdb.Set(ctx, key.StoreKey.(string), value, ttl).Err()
}

func (a *adapter) Del(key cache.GroupKey) error {
	return a.rdb.Del(ctx, key.StoreKey.(string)).Err()
}

func (a *adapter) Key(groupName string, key any) cache.GroupKey {
	adapterKey := fmt.Sprintf("%s:%v", groupName, key)
	return cache.GroupKey{GroupName: groupName, StoreKey: adapterKey}
}

// Send a message to all other caches

// Subcribe to messages from another caches

// Implement cache.Broker
func (a *adapter) Send(msg []byte) error {
	if a.topic == "" {
		panic("messing not configured")
	}
	return a.rdb.Publish(ctx, a.topic, msg).Err()
}

// Implement cache.Broker
func (a *adapter) Subscribe(handler func(msg []byte)) (io.Closer, error) {
	if a.topic == "" {
		panic("messing not configured")
	}
	pubsub := a.rdb.Subscribe(ctx, a.topic)

	// Start processing
	go func() {
		ch := pubsub.Channel()

		for msg := range ch {
			handler([]byte(msg.Payload))
		}
	}()

	var cf closerFunc = pubsub.Close
	return cf, nil
}

// To be able to return an anonymous function in Subscribe()
type closerFunc func() error

func (f closerFunc) Close() error {
	return f()
}

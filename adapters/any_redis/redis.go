package any_redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gitlab.com/sustainyfacts/anycache/cache"
)

var ctx = context.Background()

func NewAdapter(hostAndPort string, password string) cache.Store {
	return NewWithOptions(&redis.Options{
		Addr:     hostAndPort,
		Password: password,
		DB:       0, // use default DB
	})
}

func NewWithOptions(opt *redis.Options) cache.Store {
	rdb := redis.NewClient(opt)
	return &store{store: rdb, groupConfigs: make(map[string]cache.GroupConfig)}
}

type store struct {
	store        *redis.Client
	groupConfigs map[string]cache.GroupConfig
}

func (s *store) ConfigureGroup(name string, config cache.GroupConfig) {
	if config.Cost != 0 {
		panic("Redis does not support Cost")
	}
	s.groupConfigs[name] = config
}

func (s *store) Get(key cache.GroupKey) (any, bool) {
	v, err := s.store.Get(ctx, key.StoreKey.(string)).Result()
	// FIXME it is not good to swallow errors
	return v, err == nil
}

// new test
func (s *store) Set(key cache.GroupKey, value any) bool {
	ttl := s.groupConfigs[key.GroupName].Ttl
	err := s.store.Set(ctx, key.StoreKey.(string), value, ttl).Err()
	return err == nil
}

func (s *store) Del(key cache.GroupKey) {
	s.store.Del(ctx, key.StoreKey.(string))
}

// Note that this does not free memory, but rotates the hashes
// so querying the same key will require a call to the cache loader function
func (s *store) Clear(groupName string) {
	// Thoughts: use versioning and rely on TTL?
	panic("not implemented")
}

func (s *store) Key(groupName string, key any) cache.GroupKey {
	storeKey := fmt.Sprintf("%s:%v", groupName, key)
	return cache.GroupKey{GroupName: groupName, StoreKey: storeKey}
}

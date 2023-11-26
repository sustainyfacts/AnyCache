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
	"sync"
)

func NewHashMapStore() Store {
	return &store{stores: make(map[string]*sync.Map)}
}

type store struct {
	stores map[string]*sync.Map
}

func (s *store) ConfigureGroup(name string, config GroupConfig) {
	if config.Ttl != 0 {
		panic("stores does not support TTL")
	} else if config.Cost != 0 {
		panic("stores does not support Cost")
	}
	s.stores[name] = &sync.Map{}
}

func (s *store) Get(key GroupKey) (any, bool) {
	v, ok := s.stores[key.GroupName].Load(key.StoreKey)
	return v, ok
}

func (s *store) Set(key GroupKey, value any) bool {
	s.stores[key.GroupName].Store(key.StoreKey, value)
	return true
}

func (s *store) Del(key GroupKey) {
	s.stores[key.GroupName].Delete(key.StoreKey)
}

func (s *store) Clear(groupName string) {
	s.stores[groupName] = &sync.Map{}
}

func (s *store) Key(groupName string, key any) GroupKey {
	return GroupKey{GroupName: groupName, StoreKey: key}
}

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
	"time"
)

type GroupKey struct {
	GroupName string
	StoreKey  any
}

type GroupConfig struct {
	Ttl  time.Duration
	Cost int
}

type Store interface {
	// Notifies the store that a new group has been created. Will be called once per group
	// Parameters
	//  config
	ConfigureGroup(name string, config GroupConfig)
	Get(key GroupKey) (any, bool)
	Set(key GroupKey, value any) bool
	Del(key GroupKey)
	Key(groupName string, key any) GroupKey
	Clear(groupName string)
}

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
	"errors"
	"io"
	"reflect"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type GroupKey struct {
	GroupName string
	StoreKey  any
}

type GroupConfig struct {
	Ttl  time.Duration
	Cost int
	// Type of objects that are stored in the group. Some stores are not type safe
	// (for example redis stores and int an returns a string) so the adapter needs
	// to know what is the expected type of object to return
	ValueType reflect.Type
}

type Store interface {
	// Notifies the store that a new group has been created. Will be called once per group
	// Parameters
	//  config
	ConfigureGroup(name string, config GroupConfig)
	// Returns ErrKeyNotFound if the Key could not be found. Can return other types of errors
	// for example if there are connectivity issues.
	Get(key GroupKey) (any, error)
	Set(key GroupKey, value any) error
	Del(key GroupKey) error
	Key(groupName string, key any) GroupKey
}

// MessageBroker is an interface that can be used to provide clustered communication
// to the cache, for sending and receiving Flush messages
type MessageBroker interface {
	// Send a message to all other caches
	Send([]byte) error
	// Subcribe to messages from another caches
	Subscribe(func(msg []byte)) (io.Closer, error)
}

// Some brokers are as well message store
type BrokerStore interface {
	MessageBroker
	Store
}

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
	"encoding/json"
)

// Cache Message. For now only for flush
// Serialized to/from JSON to be sent by the Message Broker.
type cacheMsg struct {
	Group string
	Key   any
}

func (cm *cacheMsg) bytes() []byte {
	b, _ := json.Marshal(cm)
	return b
}

func fromBytes(b []byte) *cacheMsg {
	cm := cacheMsg{}
	json.Unmarshal(b, &cm)
	return &cm
}

// Message handler function to process messages from
// the message broker
func (g *Group[K, V]) handleMessage(msg []byte) {
	cm := fromBytes(msg)
	if cm.Group != g.name {
		return // Ignore messages from other groups
	}
	if cm.Key == nil {
		g.Clear()
	} else {
		gk := g.store.Key(g.name, cm.Key)
		g.store.Del(gk)
	}
}

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

import "testing"

func TestHashmapStore(t *testing.T) {
	var store Store = NewHashMapStore()
	store.ConfigureGroup("group", GroupConfig{})
	store.Set(GroupKey{"group", "key"}, "value")
	v, _ := store.Get(GroupKey{"group", "key"})

	if v != "value" {
		t.Errorf("value for key should be 'value' but got '%v'", v)
	}
}

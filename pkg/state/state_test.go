/*
 *
 * Copyright 2026 perrault authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package state

import (
	"testing"

	require "github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	type testData struct {
		data  map[string]any
		key   string
		value any
		want  map[string]any
	}

	testSet := func(name string, td testData) {
		t.Helper()

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			Set(td.data, td.key, td.value)
			require.Equal(t, td.data, td.want)
		})
	}

	testSet("sets top-level value", testData{
		data:  map[string]any{},
		key:   "name",
		value: "perrault",
		want:  map[string]any{"name": "perrault"},
	})

	testSet("creates nested maps", testData{
		data: map[string]any{
			"player": map[string]any{"class": "wizard"},
		},
		key:   "player.inventory.coins",
		value: 12,
		want: map[string]any{
			"player": map[string]any{
				"class":     "wizard",
				"inventory": map[string]any{"coins": 12},
			},
		},
	})

	testSet("updates existing nested value", testData{
		data: map[string]any{
			"player": map[string]any{"inventory": map[string]any{"coins": 12}},
		},
		key:   "player.inventory.coins",
		value: 24,
		want: map[string]any{
			"player": map[string]any{"inventory": map[string]any{"coins": 24}},
		},
	})

	testSet("replaces existing value with nested", testData{
		data: map[string]any{
			"player": map[string]any{"inventory": "coins=12"},
		},
		key:   "player.inventory.potions.healing",
		value: 5,
		want: map[string]any{
			"player": map[string]any{"inventory": map[string]any{"potions": map[string]any{"healing": 5}}},
		},
	})
}

func TestGet(t *testing.T) {
	inventory := map[string]any{
		"coins": 12,
	}

	player := map[string]any{
		"class":     "wizard",
		"inventory": inventory,
		"health":    100,
	}

	data := map[string]any{
		"name":     "perrault",
		"player":   player,
		"location": "castle",
	}

	testGet := func(name string, key string, want any) {
		t.Helper()

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			t.Helper()

			got := Get(data, key)
			require.Equal(t, want, got)
		})
	}

	testGet("basic string", "name", "perrault")
	testGet("nested string", "player.class", "wizard")
	testGet("nested int", "player.health", 100)
	testGet("deep nested int", "player.inventory.coins", 12)
	testGet("raw map", "player", player)
	testGet("nested raw map", "player.inventory", inventory)
	testGet("missing", "missing", nil)
	testGet("deep missing", "player.missing", nil)
	testGet("no nested value", "location.name", nil)
	testGet("untouched string", "location", "castle")
}

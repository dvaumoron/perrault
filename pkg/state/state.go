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
	"encoding/json/v2"
	"slices"
	"strings"
)

func Set(data map[string]any, key string, value any) {
	keys := strings.Split(key, ".")
	last := len(keys) - 1

	current := data
	for index, key := range keys[:last] {
		next, ok := current[key].(map[string]any)
		if !ok {
			current[key] = initNestedMap(keys[index+1:], value)
			return
		}
		current = next
	}

	current[keys[last]] = value
}

func Get(data map[string]any, key string) any {
	firstKey, remainingKeys, ok := strings.Cut(key, ".")
	current, _ := data[firstKey]
	if !ok {
		return current
	}

	for key := range strings.SplitSeq(remainingKeys, ".") {
		casted, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		value, ok := casted[key]
		if !ok {
			return nil
		}

		current = value
	}

	return current
}

func GetInt(data map[string]any, key string) (int, error) {
	return 0, nil
}

func GetString(data map[string]any, key string) (string, error) {
	return "", nil
}

func GetMap(data map[string]any, key string) (map[string]any, error) {
	return nil, nil
}

func GetSlice(data map[string]any, key string) ([]any, error) {
	return nil, nil
}

func Rollback(storage Storage) (map[string]any, error) {
	rawData, err := storage.Read()
	if err != nil {
		return nil, err
	}

	var data map[string]any
	return data, json.Unmarshal(rawData, &data)
}

func Commit(storage Storage, data map[string]any) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return storage.Write(rawData)
}

type Storage interface {
	Read() ([]byte, error)
	Write([]byte) error
}

func initNestedMap(keys []string, value any) any {
	for _, key := range slices.Backward(keys) {
		value = map[string]any{
			key: value,
		}
	}
	return value
}

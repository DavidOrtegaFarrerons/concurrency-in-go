package kvstore

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestInMemoryKVStoreGet(t *testing.T) {

	tests := []struct {
		name     string
		setup    func(s *InMemoryKVStore)
		key      string
		expected string
		exists   bool
	}{
		{
			name: "Add name in kv",
			setup: func(s *InMemoryKVStore) {
				s.Set("name", "david")
			},
			key:      "name",
			expected: "david",
			exists:   true,
		},
		{
			name: "Add name in kv",
			setup: func(s *InMemoryKVStore) {
				s.Get("name")
			},
			key:      "name",
			expected: "",
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvStore := NewInMemoryKVStore()

			tt.setup(kvStore)

			v, exists := kvStore.Get(tt.key)

			Equal(t, exists, tt.exists)
			Equal(t, v, tt.expected)
		})
	}
}

func TestInMemoryConcurrent(t *testing.T) {
	kvStore := NewInMemoryKVStore()

	wg := sync.WaitGroup{}

	for i := range 1000 {
		fmt.Println(i)
		wg.Add(1)

		go func() {
			kvStore.Set(strconv.Itoa(rand.Int()), strconv.Itoa(rand.Int()))
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			v, exists := kvStore.Get(strconv.Itoa(rand.Int()))
			if exists {
				fmt.Printf("Found value: %s", v)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

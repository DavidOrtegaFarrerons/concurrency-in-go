package kvstore

import (
	"sync"
)

type KVStore interface {
	Get(key string) (string, bool)
	Set(key, value string)
}

type InMemoryKVStore struct {
	rwMutex sync.RWMutex
	kv      map[string]string
	init    sync.Once
}

func NewInMemoryKVStore() *InMemoryKVStore {
	return &InMemoryKVStore{kv: make(map[string]string)}
}

func (s *InMemoryKVStore) Set(key, value string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	s.kv[key] = value
}

func (s *InMemoryKVStore) Get(key string) (string, bool) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	v, exists := s.kv[key]

	return v, exists
}

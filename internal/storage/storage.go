package storage

import (
	"sync"
)

type Storage struct {
	mu   sync.RWMutex
	data map[string]*Value
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]*Value),
	}
}

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &Value{
		Data: value,
	}

}

func (s *Storage) Get(key string) (*Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, false
	}

	return value, true
}

func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		return true
	}

	return false
}

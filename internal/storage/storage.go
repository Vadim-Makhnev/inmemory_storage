package storage

import (
	"sync"
	"time"
)

type Storage struct {
	mu   sync.RWMutex
	data map[string]*Value
}

func NewStorage() *Storage {
	s := &Storage{
		data: make(map[string]*Value),
	}

	s.StartGC(30 * time.Second)
	return s
}

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &Value{
		Data: value,
	}
}

func (s *Storage) SetEx(key, value string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt := time.Now().Add(duration)
	s.data[key] = &Value{
		Data:      value,
		ExpiresAt: &expiresAt,
	}
}

func (s *Storage) Get(key string) (*Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, false
	}

	if value.ExpiresAt != nil && value.ExpiresAt.Before(time.Now()) {
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

func (s *Storage) StartGC(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()

			for key, value := range s.data {
				if value.ExpiresAt != nil && value.ExpiresAt.Before(now) {
					delete(s.data, key)
				}
			}

			s.mu.Unlock()
		}
	}()
}

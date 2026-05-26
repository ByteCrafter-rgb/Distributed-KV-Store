package store

import (
	"sync"
	"time"
)

type Value struct {
	Data      string
	ExpiresAt *time.Time
}

func (v *Value) IsExpired() bool {
	if v.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*v.ExpiresAt)
}

type Store struct {
	mu   sync.RWMutex
	data map[string]*Value
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]*Value),
	}
}

func (s *Store) Set(key, value string, expireIn *time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt *time.Time
	if expireIn != nil {
		t := time.Now().Add(*expireIn)
		expiresAt = &t
	}

	s.data[key] = &Value{
		Data:      value,
		ExpiresAt: expiresAt,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return "", false
	}

	if val.IsExpired() {
		return "", false
	}

	return val.Data, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return false
	}

	return !val.IsExpired()
}

func (s *Store) TTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists || val.IsExpired() {
		return -2
	}

	if val.ExpiresAt == nil {
		return -1
	}

	ttl := int64(time.Until(*val.ExpiresAt).Seconds())
	if ttl < 0 {
		return -2
	}
	return ttl
}

func (s *Store) Expire(key string, seconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	if !exists || val.IsExpired() {
		return false
	}

	expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)
	val.ExpiresAt = &expiresAt
	return true
}

func (s *Store) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, val := range s.data {
		if val.IsExpired() {
			delete(s.data, key)
		}
	}
}

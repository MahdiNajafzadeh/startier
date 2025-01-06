package easynode

import "sync"

type Store[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Del(key K) (V, bool)
	All() map[K]V
	Each(func(K, V) bool)
}

var _ Store[string, string] = &store[string, string]{}

func newStore[K comparable, V any]() Store[K, V] {
	return &store[K, V]{
		mu:   sync.RWMutex{},
		pool: make(map[K]V),
	}
}

type store[K comparable, V any] struct {
	mu   sync.RWMutex
	pool map[K]V
}

func (s *store[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.pool[key]
	return value, ok
}

func (s *store[K, V]) Set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pool[key] = value
}

func (s *store[K, V]) Del(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.pool[key]
	if ok {
		delete(s.pool, key)
	}
	return value, ok
}

func (s *store[K, V]) All() map[K]V {
	return s.pool
}

func (s *store[K, V]) Each(f func(key K, value V) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.pool {
		if f(k, v) {
			break
		}
	}
}

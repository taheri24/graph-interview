package cache

import (
	"fmt"
	"sync"
)

// InMemoryCacheImpl is an in-memory cache implementation using map[string]any as store
type InMemoryCacheImpl[T any] struct {
	store map[string]any
	mu    sync.RWMutex
}

var _ CacheInterface[any] = (*InMemoryCacheImpl[any])(nil)

// NewInMemoryCacheImpl creates a new InMemoryCacheImpl instance
func NewInMemoryCacheImpl[T any]() *InMemoryCacheImpl[T] {
	return &InMemoryCacheImpl[T]{
		store: make(map[string]any),
	}
}

// Get implements CacheInterface.Get - retrieves an item by ID from the cache
func (m *InMemoryCacheImpl[T]) Get(id string) (*T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if item, exists := m.store[id]; exists {
		if typedItem, ok := item.(T); ok {
			return &typedItem, nil
		}
		// Type assertion failed
		return nil, fmt.Errorf("type assertion failed for cached item")
	}

	return nil, nil
}

// Set implements CacheInterface.Set - stores an item in the cache
func (m *InMemoryCacheImpl[T]) Set(id string, item T) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[id] = item
	return nil
}

// Invalidate implements CacheInterface.Invalidate - removes an item from the cache
func (m *InMemoryCacheImpl[T]) Invalidate(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, id)
	return nil
}

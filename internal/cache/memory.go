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

// GetAll implements CacheInterface.GetAll - returns all items in the cache
func (m *InMemoryCacheImpl[T]) GetAll() ([]T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]T, 0, len(m.store))
	for _, v := range m.store {
		if item, ok := v.(T); ok {
			items = append(items, item)
		}
	}
	return items, nil
}

// SetAll implements CacheInterface.SetAll - stores all items in the cache
func (m *InMemoryCacheImpl[T]) SetAll(items []T) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing items
	m.store = make(map[string]any)

	// For SetAll, we need an ID. Since we don't have one, we'll use index-based keys
	for i, item := range items {
		key := fmt.Sprintf("item_%d", i)
		m.store[key] = item
	}
	return nil
}

// InvalidateAll implements CacheInterface.InvalidateAll - clears all cached items
func (m *InMemoryCacheImpl[T]) InvalidateAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store = make(map[string]any)
	return nil
}

// Get implements CacheInterface.Get - retrieves an item by ID from the cache
func (m *InMemoryCacheImpl[T]) Get(id string) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if item, exists := m.store[id]; exists {
		if typedItem, ok := item.(T); ok {
			return typedItem, nil
		}
		// Type assertion failed
		var zero T
		return zero, fmt.Errorf("type assertion failed for cached item")
	}

	var zero T
	return zero, fmt.Errorf("item not found in cache")
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

package cache

// NoOpCacheImpl is a no-operation cache implementation that does nothing
type NoOpCacheImpl[T any] struct{}

var _ CacheInterface[any] = (*NoOpCacheImpl[any])(nil)

// NewNoOpCacheImpl creates a new NoOpCacheImpl instance
func NewNoOpCacheImpl[T any]() *NoOpCacheImpl[T] {
	return &NoOpCacheImpl[T]{}
}

// GetAll implements CacheInterface.GetAll - always returns empty slice
func (n *NoOpCacheImpl[T]) GetAll() ([]T, error) {
	return []T{}, nil
}

// SetAll implements CacheInterface.SetAll - does nothing
func (n *NoOpCacheImpl[T]) SetAll(items []T) error {
	return nil
}

// InvalidateAll implements CacheInterface.InvalidateAll - does nothing
func (n *NoOpCacheImpl[T]) InvalidateAll() error {
	return nil
}

// Get implements CacheInterface.Get - always returns error (cache miss)
func (n *NoOpCacheImpl[T]) Get(id string) (T, error) {
	var zero T
	return zero, nil
}

// Set implements CacheInterface.Set - does nothing
func (n *NoOpCacheImpl[T]) Set(id string, item T) error {
	return nil
}

// Invalidate implements CacheInterface.Invalidate - does nothing
func (n *NoOpCacheImpl[T]) Invalidate(id string) error {
	return nil
}

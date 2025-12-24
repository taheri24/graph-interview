package cache

// NoOpCacheImpl is a no-operation cache implementation that does nothing
type NoOpCacheImpl[T any] struct{}

var _ CacheInterface[any] = (*NoOpCacheImpl[any])(nil)

// NewNoOpCacheImpl creates a new NoOpCacheImpl instance
func NewNoOpCacheImpl[T any]() *NoOpCacheImpl[T] {
	return &NoOpCacheImpl[T]{}
}

// Get implements CacheInterface.Get - always returns nil (cache miss)
func (n *NoOpCacheImpl[T]) Get(id string) (*T, error) {
	return nil, nil
}

// Set implements CacheInterface.Set - does nothing
func (n *NoOpCacheImpl[T]) Set(id string, item T) error {
	return nil
}

// Invalidate implements CacheInterface.Invalidate - does nothing
func (n *NoOpCacheImpl[T]) Invalidate(id string) error {
	return nil
}

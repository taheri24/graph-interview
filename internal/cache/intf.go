package cache

// CacheInterface defines the interface for cache operations
type CacheInterface[T any] interface {
	GetAll() ([]T, error)
	SetAll(tasks []T) error
	InvalidateAll() error

	Get(id string) (T, error)
	Set(id string, item T) error
	Invalidate(id string) error
}

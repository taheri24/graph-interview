package cache

import (
	"context"
	"fmt"
)

// RedisCacheImpl is an empty struct that implements CacheInterface
type RedisCacheImpl[T any] struct {
	sectionName string
	redisCache  *RedisCache
}

var _ CacheInterface[any] = (*RedisCacheImpl[any])(nil)

// GetAll implements CacheInterface.GetAll
func (r *RedisCacheImpl[T]) GetAll() ([]T, error) {
	return Get[[]T](r.redisCache, "%s", r.sectionName)
}

// SetAll implements CacheInterface.SetAll
func (r *RedisCacheImpl[T]) SetAll(items []T) error {
	return Set(r.redisCache, items, "%s", r.sectionName)
}

// InvalidateAll implements CacheInterface.InvalidateAll
func (r *RedisCacheImpl[T]) InvalidateAll() error {
	return r.redisCache.client.Del(context.Background(), r.sectionName).Err()
}

// Get implements CacheInterface.Get
func (r *RedisCacheImpl[T]) Get(id string) (T, error) {
	return Get[T](r.redisCache, "%s:%s", r.sectionName, id)
}

// Set implements CacheInterface.Set
func (r *RedisCacheImpl[T]) Set(id string, item T) error {
	return Set(r.redisCache, item, "%s:%s", r.sectionName, id)
}

// Invalidate implements CacheInterface.Invalidate
func (r *RedisCacheImpl[T]) Invalidate(id string) error {
	return r.redisCache.client.Del(context.Background(), fmt.Sprintf("%s:%s", r.sectionName, id)).Err()
}

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

// NewRedisCacheImpl creates a new RedisCacheImpl instance
func NewRedisCacheImpl[T any](sectionName string, redisCache *RedisCache) *RedisCacheImpl[T] {
	return &RedisCacheImpl[T]{
		sectionName: sectionName,
		redisCache:  redisCache,
	}
}

// Get implements CacheInterface.Get
func (r *RedisCacheImpl[T]) Get(id string) (*T, error) {
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

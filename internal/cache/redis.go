package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"taheri24.ir/graph1/pkg/utils"
)

// RedisCache handles Redis caching operations
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Get retrieves data from Redis cache
func (r *RedisCache) Get(key string, data any) error {
	return r.client.Get(r.ctx, key).Scan(data)
}

// Set stores data in Redis cache
func (r *RedisCache) Set(key string, data any, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, data, expiration).Err()
}

// Get gets data from Redis cache
func Get[T any](r *RedisCache, format string, args ...any) (T, error) {
	key := fmt.Sprintf(format, args...)
	cmd := r.client.Get(r.ctx, key)
	var blank T
	raw, err := cmd.Bytes()
	if err != nil {
		return blank, err
	}
	return utils.JsonDecode[T](raw), nil

}

// Set sets data to Redis cache
func Set[T any](r *RedisCache, value T, format string, args ...any) error {
	key := fmt.Sprintf(format, args...)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, data, 0).Err()
}

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"taheri24.ir/graph1/internal/models"
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

// GetTasks retrieves cached tasks list
func (r *RedisCache) GetTasks() ([]models.Task, error) {
	data, err := r.client.Get(r.ctx, "tasks:list").Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from cache: %w", err)
	}

	var tasks []models.Task
	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached tasks: %w", err)
	}

	return tasks, nil
}

// SetTasks caches the tasks list with expiration
func (r *RedisCache) SetTasks(tasks []models.Task, expiration time.Duration) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := r.client.Set(r.ctx, "tasks:list", data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set tasks in cache: %w", err)
	}

	return nil
}

// InvalidateTasks removes the tasks list from cache
func (r *RedisCache) InvalidateTasks() error {
	if err := r.client.Del(r.ctx, "tasks:list").Err(); err != nil {
		return fmt.Errorf("failed to invalidate tasks cache: %w", err)
	}
	return nil
}

// GetTask retrieves a single cached task by ID
func (r *RedisCache) GetTask(id string) (*models.Task, error) {
	key := fmt.Sprintf("task:%s", id)
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task from cache: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached task: %w", err)
	}

	return &task, nil
}

// SetTask caches a single task with expiration
func (r *RedisCache) SetTask(task *models.Task, expiration time.Duration) error {
	key := fmt.Sprintf("task:%s", task.ID)
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set task in cache: %w", err)
	}

	return nil
}

// InvalidateTask removes a single task from cache
func (r *RedisCache) InvalidateTask(id string) error {
	key := fmt.Sprintf("task:%s", id)
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate task cache: %w", err)
	}
	return nil
}

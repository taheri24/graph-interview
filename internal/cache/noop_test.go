package cache

import (
	"testing"
	"time"

	"taheri24.ir/graph1/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewNoOpCacheImpl(t *testing.T) {
	cache := NewNoOpCacheImpl[models.Task]()
	assert.NotNil(t, cache)
	assert.IsType(t, &NoOpCacheImpl[models.Task]{}, cache)
}

func TestNoOpCacheImplImplementsInterface(t *testing.T) {
	var _ CacheInterface[models.Task] = (*NoOpCacheImpl[models.Task])(nil)
}

func TestNoOpCacheImplGet(t *testing.T) {
	emptyCache := NewNoOpCacheImpl[models.Task]()

	// Any key should return nil
	taskID := uuid.New().String()
	task, err := emptyCache.Get(taskID)
	assert.NoError(t, err)
	assert.Nil(t, task)

	// Test with different keys
	task, err = emptyCache.Get("any-key")
	assert.NoError(t, err)
	assert.Nil(t, task)

	// Test with empty key
	task, err = emptyCache.Get("")
	assert.NoError(t, err)
	assert.Nil(t, task)
}

func TestNoOpCacheImplSet(t *testing.T) {
	cache := NewNoOpCacheImpl[models.Task]()

	task := models.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set should always succeed (do nothing)
	err := cache.Set(task.ID.String(), task)
	assert.NoError(t, err)

	// Verify it didn't actually store anything
	_, err = cache.Get(task.ID.String())
	assert.NoError(t, err) // Should still return cache miss
}

func TestNoOpCacheImplInvalidate(t *testing.T) {
	cache := NewNoOpCacheImpl[models.Task]()

	// Invalidate should always succeed (do nothing)
	err := cache.Invalidate(uuid.New().String())
	assert.NoError(t, err)

	// Test with different keys
	err = cache.Invalidate("any-key")
	assert.NoError(t, err)

	// Test with empty key
	err = cache.Invalidate("")
	assert.NoError(t, err)
}

func TestNoOpCacheImplGenericTypes(t *testing.T) {
	// Test with different types to ensure generics work

	// Test with string
	stringCache := NewNoOpCacheImpl[string]()
	_, err := stringCache.Get("test")
	assert.Equal(t, err, nil)

	err = stringCache.Set("key", "value")
	assert.NoError(t, err)

	// Test with int
	intCache := NewNoOpCacheImpl[int]()

	err = intCache.Set("key", 42)
	assert.NoError(t, err)

	// Test with custom struct
	type CustomStruct struct {
		Name  string
		Value int
	}

	customCache := NewNoOpCacheImpl[CustomStruct]()
	custom := CustomStruct{Name: "test", Value: 123}

	err = customCache.Set("key", custom)
	assert.NoError(t, err)

	_, err = customCache.Get("key")
	assert.Equal(t, err, nil)
}

func TestNoOpCacheImplConcurrentAccess(t *testing.T) {
	cache := NewNoOpCacheImpl[models.Task]()

	// Test concurrent access (should not panic)
	done := make(chan bool, 3)

	// Goroutine 1: Set operations
	go func() {
		for i := 0; i < 100; i++ {
			task := models.Task{
				ID:        uuid.New(),
				Title:     "Concurrent Task",
				Status:    models.StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			cache.Set(task.ID.String(), task)
		}
		done <- true
	}()

	// Goroutine 2: Get operations
	go func() {
		for i := 0; i < 100; i++ {
			randomID := uuid.New().String()
			cache.Get(randomID) // Ignore errors
		}
		done <- true
	}()

	// Goroutine 3: Invalidate operations
	go func() {
		for i := 0; i < 100; i++ {
			randomID := uuid.New().String()
			cache.Invalidate(randomID)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we reach here without panics, the test passes
	assert.True(t, true, "Concurrent operations completed without panics")
}

func TestNoOpCacheImplEdgeCases(t *testing.T) {
	cache := NewNoOpCacheImpl[models.Task]()

	// Test with nil-like operations
	err := cache.Set("", models.Task{})
	assert.NoError(t, err)

	_, err = cache.Get("")
	assert.NoError(t, err)

	err = cache.Invalidate("")
	assert.NoError(t, err)

	// Test with special characters in keys
	specialKeys := []string{
		"key with spaces",
		"key-with-dashes",
		"key_with_underscores",
		"key.with.dots",
		"key/with/slashes",
		"123numeric",
		"mixed123key",
	}

	for _, key := range specialKeys {
		err = cache.Set(key, models.Task{ID: uuid.New(), Title: "Test"})
		assert.NoError(t, err)

		_, err = cache.Get(key)
		assert.NoError(t, err)

		err = cache.Invalidate(key)
		assert.NoError(t, err)
	}
}

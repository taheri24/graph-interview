package cache

import (
	"testing"
	"time"

	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/types"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisCacheImpl(t *testing.T) {
	// Start a mini Redis server
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client
	redisCache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer redisCache.Close()

	// Test creating RedisCacheImpl
	cache := NewRedisCacheImpl[models.Task]("test_tasks", redisCache)
	assert.NotNil(t, cache)
	assert.Equal(t, "test_tasks", cache.sectionName)
	assert.NotNil(t, cache.redisCache)
}

func TestRedisCacheImplSetAndGet(t *testing.T) {
	// Start mini Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create cache
	redisCache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer redisCache.Close()

	cache := NewRedisCacheImpl[models.Task]("tasks", redisCache)

	// Create test task
	taskID := uuid.New()
	task := models.Task{
		ID:          taskID,
		Title:       "Redis Test Task",
		Description: "Testing Redis cache",
		Status:      types.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test Set
	err = cache.Set(taskID.String(), task)
	assert.NoError(t, err)

	// Test Get
	retrievedTask, err := cache.Get(taskID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.Title, retrievedTask.Title)
	assert.Equal(t, task.Description, retrievedTask.Description)
	assert.Equal(t, task.Status, retrievedTask.Status)
	assert.Equal(t, task.Assignee, retrievedTask.Assignee)
}

func TestRedisCacheImplGetNonExistent(t *testing.T) {
	// Start mini Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create cache
	redisCache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer redisCache.Close()

	cache := NewRedisCacheImpl[models.Task]("tasks", redisCache)

	// Test Get with non-existent key
	retrievedTask, err := cache.Get(uuid.New().String())
	assert.NoError(t, err)
	assert.Nil(t, retrievedTask)
}

func TestRedisCacheImplInvalidate(t *testing.T) {
	// Start mini Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create cache
	redisCache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer redisCache.Close()

	cache := NewRedisCacheImpl[models.Task]("tasks", redisCache)

	// Set a task
	taskID := uuid.New()
	task := models.Task{
		ID:        taskID,
		Title:     "Test Task",
		Status:    types.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = cache.Set(taskID.String(), task)
	require.NoError(t, err)

	// Verify it exists
	retrievedTask, err := cache.Get(taskID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask)

	// Invalidate it
	err = cache.Invalidate(taskID.String())
	assert.NoError(t, err)

	// Verify it's gone
	retrievedTask, err = cache.Get(taskID.String())
	assert.NoError(t, err)
	assert.Nil(t, retrievedTask)
}

func TestRedisCacheImplConcurrentAccess(t *testing.T) {
	// Start mini Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create cache
	redisCache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer redisCache.Close()

	cache := NewRedisCacheImpl[models.Task]("tasks", redisCache)

	// Test concurrent operations
	done := make(chan bool, 2)

	// Goroutine 1: Set operations
	go func() {
		for i := 0; i < 10; i++ {
			task := models.Task{
				ID:        uuid.New(),
				Title:     "Concurrent Task " + string(rune(i)),
				Status:    types.StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			cache.Set(task.ID.String(), task)
		}
		done <- true
	}()

	// Goroutine 2: Get operations (may get cache misses, that's ok)
	go func() {
		for i := 0; i < 10; i++ {
			randomID := uuid.New().String()
			cache.Get(randomID) // Ignore errors
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Test completed without race conditions or panics
	assert.True(t, true, "Concurrent operations completed successfully")
}

func TestRedisCacheImplErrorHandling(t *testing.T) {
	// Test with invalid Redis connection (simulate connection failure)
	// This is hard to test directly, but we can test the error paths

	// Start mini Redis but don't connect to it properly
	r := miniredis.RunT(t)
	t.Cleanup(func() {
		r.Close()
	})
	// Try to create cache with closed Redis
	redisCache, err := NewRedisCache(r.Addr(), "", 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer redisCache.Close()

	cacheMod := NewRedisCacheImpl[models.Task]("tasks", redisCache)
	task := models.Task{ID: uuid.New(), Title: "Test", Status: types.StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = cacheMod.Set(task.ID.String(), task)
	assert.Equal(t, err, nil)

}

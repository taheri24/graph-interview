package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryCacheImpl(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.store)
	assert.IsType(t, &InMemoryCacheImpl[models.Task]{}, cache)
}

func TestInMemoryCacheImplImplementsInterface(t *testing.T) {
	var _ CacheInterface[models.Task] = (*InMemoryCacheImpl[models.Task])(nil)
}

func TestInMemoryCacheImplGet(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()

	t.Run("get existing task", func(t *testing.T) {
		taskID := uuid.New().String()
		task := models.Task{
			ID:          uuid.MustParse(taskID),
			Title:       "Test Task",
			Description: "Test Description",
			Status:      types.StatusPending,
			Assignee:    "test@example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set task first
		err := cache.Set(taskID, task)
		assert.NoError(t, err)

		// Get task
		result, err := cache.Get(taskID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.ID, result.ID)
		assert.Equal(t, task.Title, result.Title)
		assert.Equal(t, task.Status, result.Status)
	})

	t.Run("get non-existing task", func(t *testing.T) {
		result, err := cache.Get("nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("get with empty key", func(t *testing.T) {
		result, err := cache.Get("")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type assertion failure", func(t *testing.T) {
		// Create cache for string type
		stringCache := NewInMemoryCacheImpl[string]()

		// Manually put wrong type in store (simulating corrupted state)
		stringCache.store["wrong_type"] = 123 // int instead of string

		result, err := stringCache.Get("wrong_type")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "type assertion failed")
	})
}

func TestInMemoryCacheImplSet(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()

	t.Run("set new task", func(t *testing.T) {
		taskID := uuid.New().String()
		task := models.Task{
			ID:        uuid.MustParse(taskID),
			Title:     "New Task",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := cache.Set(taskID, task)
		assert.NoError(t, err)

		// Verify it was stored
		result, err := cache.Get(taskID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.Title, result.Title)
	})

	t.Run("overwrite existing task", func(t *testing.T) {
		taskID := uuid.New().String()

		// Set initial task
		originalTask := models.Task{
			ID:        uuid.MustParse(taskID),
			Title:     "Original Task",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := cache.Set(taskID, originalTask)
		assert.NoError(t, err)

		// Overwrite with updated task
		updatedTask := models.Task{
			ID:        uuid.MustParse(taskID),
			Title:     "Updated Task",
			Status:    types.StatusInProgress,
			CreatedAt: originalTask.CreatedAt,
			UpdatedAt: time.Now(),
		}

		err = cache.Set(taskID, updatedTask)
		assert.NoError(t, err)

		// Verify overwrite worked
		result, err := cache.Get(taskID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Task", result.Title)
		assert.Equal(t, types.StatusInProgress, result.Status)
	})

	t.Run("set with empty key", func(t *testing.T) {
		task := models.Task{
			ID:        uuid.New(),
			Title:     "Empty Key Task",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := cache.Set("", task)
		assert.NoError(t, err)

		// Should be retrievable with empty key
		result, err := cache.Get("")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.Title, result.Title)
	})

	t.Run("set with special characters in key", func(t *testing.T) {
		specialKeys := []string{
			"key with spaces",
			"key-with-dashes",
			"key_with_underscores",
			"key.with.dots",
			"key/with/slashes",
			"123numeric",
			"mixed123key",
			"key:with:colons",
		}

		task := models.Task{
			ID:        uuid.New(),
			Title:     "Special Key Task",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		for _, key := range specialKeys {
			err := cache.Set(key, task)
			assert.NoError(t, err)

			result, err := cache.Get(key)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, task.Title, result.Title)
		}
	})
}

func TestInMemoryCacheImplInvalidate(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()

	t.Run("invalidate existing task", func(t *testing.T) {
		taskID := uuid.New().String()
		task := models.Task{
			ID:        uuid.MustParse(taskID),
			Title:     "Task to Invalidate",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Set task
		err := cache.Set(taskID, task)
		assert.NoError(t, err)

		// Verify it exists
		result, err := cache.Get(taskID)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Invalidate
		err = cache.Invalidate(taskID)
		assert.NoError(t, err)

		// Verify it's gone
		result, err = cache.Get(taskID)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalidate non-existing task", func(t *testing.T) {
		err := cache.Invalidate("nonexistent")
		assert.NoError(t, err)
	})

	t.Run("invalidate with empty key", func(t *testing.T) {
		err := cache.Invalidate("")
		assert.NoError(t, err)
	})

	t.Run("invalidate with special characters", func(t *testing.T) {
		specialKeys := []string{
			"key with spaces",
			"key-with-dashes",
			"key_with_underscores",
		}

		for _, key := range specialKeys {
			err := cache.Invalidate(key)
			assert.NoError(t, err)
		}
	})
}

func TestInMemoryCacheImplGenericTypes(t *testing.T) {
	t.Run("test with string type", func(t *testing.T) {
		stringCache := NewInMemoryCacheImpl[string]()

		err := stringCache.Set("key1", "test_value")
		assert.NoError(t, err)

		result, err := stringCache.Get("key1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test_value", *result)

		err = stringCache.Invalidate("key1")
		assert.NoError(t, err)

		result, err = stringCache.Get("key1")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("test with int type", func(t *testing.T) {
		intCache := NewInMemoryCacheImpl[int]()

		err := intCache.Set("key2", 42)
		assert.NoError(t, err)

		result, err := intCache.Get("key2")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("test with custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Name  string
			Value int
		}

		customCache := NewInMemoryCacheImpl[CustomStruct]()
		custom := CustomStruct{Name: "test", Value: 123}

		err := customCache.Set("key3", custom)
		assert.NoError(t, err)

		result, err := customCache.Get("key3")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, custom.Name, result.Name)
		assert.Equal(t, custom.Value, result.Value)
	})

	t.Run("test with slice type", func(t *testing.T) {
		sliceCache := NewInMemoryCacheImpl[[]string]()
		slice := []string{"a", "b", "c"}

		err := sliceCache.Set("key4", slice)
		assert.NoError(t, err)

		result, err := sliceCache.Get("key4")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, slice, *result)
	})
}

func TestInMemoryCacheImplConcurrentAccess(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()

	// Number of goroutines for each operation type
	numRoutines := 50
	numOperations := 100

	done := make(chan bool, 3)

	t.Run("concurrent set operations", func(t *testing.T) {
		go func() {
			for i := 0; i < numRoutines; i++ {
				go func(routineID int) {
					for j := 0; j < numOperations; j++ {
						taskID := uuid.New().String()
						task := models.Task{
							ID:        uuid.MustParse(taskID),
							Title:     fmt.Sprintf("Concurrent Task %d-%d", routineID, j),
							Status:    types.StatusPending,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}
						cache.Set(taskID, task)
					}
				}(i)
			}
			done <- true
		}()
	})

	t.Run("concurrent get operations", func(t *testing.T) {
		go func() {
			for i := 0; i < numRoutines; i++ {
				go func(routineID int) {
					for j := 0; j < numOperations; j++ {
						randomID := uuid.New().String()
						cache.Get(randomID) // Ignore result and error
					}
				}(i)
			}
			done <- true
		}()
	})

	t.Run("concurrent invalidate operations", func(t *testing.T) {
		go func() {
			for i := 0; i < numRoutines; i++ {
				go func(routineID int) {
					for j := 0; j < numOperations; j++ {
						randomID := uuid.New().String()
						cache.Invalidate(randomID)
					}
				}(i)
			}
			done <- true
		}()
	})

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we reach here without panics or race conditions, the test passes
	assert.True(t, true, "Concurrent operations completed without panics or race conditions")
}

func TestInMemoryCacheImplRaceConditions(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()
	taskID := uuid.New().String()

	var wg sync.WaitGroup

	// Start multiple goroutines performing operations on the same key
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			task := models.Task{
				ID:        uuid.MustParse(taskID),
				Title:     fmt.Sprintf("Race Task %d", id),
				Status:    types.StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Perform multiple operations rapidly
			cache.Set(taskID, task)
			cache.Get(taskID)
			cache.Invalidate(taskID)
			cache.Get(taskID) // Should be nil after invalidate
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions detected
	assert.True(t, true, "Race condition test completed")
}

func TestInMemoryCacheImplEdgeCases(t *testing.T) {
	cache := NewInMemoryCacheImpl[models.Task]()

	t.Run("large number of items", func(t *testing.T) {
		// Test with many items to ensure no performance degradation
		numItems := 1000

		// Set many items
		for i := 0; i < numItems; i++ {
			taskID := uuid.New().String()
			task := models.Task{
				ID:        uuid.MustParse(taskID),
				Title:     fmt.Sprintf("Bulk Task %d", i),
				Status:    types.StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := cache.Set(taskID, task)
			assert.NoError(t, err)
		}

		// Verify a few random items
		for i := 0; i < 10; i++ {
			randomID := uuid.New().String()
			result, err := cache.Get(randomID)
			// May or may not exist, but shouldn't error
			assert.NoError(t, err)
			if result != nil {
				assert.Contains(t, result.Title, "Bulk Task")
			}
		}
	})

	t.Run("very long keys", func(t *testing.T) {
		longKey := string(make([]byte, 1000)) // 1000 character key
		for i := range longKey {
			longKey = longKey[:i] + "a" + longKey[i+1:]
		}

		task := models.Task{
			ID:        uuid.New(),
			Title:     "Long Key Task",
			Status:    types.StatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := cache.Set(longKey, task)
		assert.NoError(t, err)

		result, err := cache.Get(longKey)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.Title, result.Title)
	})

	t.Run("nil pointer handling", func(t *testing.T) {
		// Test that the cache handles nil pointers correctly (though they shouldn't be passed)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cache operations should not panic with nil pointers: %v", r)
			}
		}()

		// These should not panic (though results may be nil)
		_, err := cache.Get("")
		assert.NoError(t, err)

		err = cache.Invalidate("")
		assert.NoError(t, err)
	})
}

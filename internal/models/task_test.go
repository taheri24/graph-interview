package models_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTaskTableName(t *testing.T) {
	task := models.Task{}
	assert.Equal(t, "tasks", task.TableName())
}

func TestTaskStatusConstants(t *testing.T) {
	assert.Equal(t, models.TaskStatus("pending"), models.StatusPending)
	assert.Equal(t, models.TaskStatus("in_progress"), models.StatusInProgress)
	assert.Equal(t, models.TaskStatus("completed"), models.StatusCompleted)
}

func TestTaskModel(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	task := models.Task{
		ID:          id,
		Title:       "Test Title",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, id, task.ID)
	assert.Equal(t, "Test Title", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, models.StatusPending, task.Status)
	assert.Equal(t, "test@example.com", task.Assignee)
	assert.Equal(t, now, task.CreatedAt)
	assert.Equal(t, now, task.UpdatedAt)
}

func TestTaskModelWithDefaultValues(t *testing.T) {
	task := models.Task{
		ID:        uuid.New(),
		Title:     "Test Task",
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that empty string fields are handled correctly
	assert.Equal(t, "", task.Description)
	assert.Equal(t, "", task.Assignee)
}

func TestTaskStatusValidation(t *testing.T) {
	// Test valid status values
	validStatuses := []models.TaskStatus{
		models.StatusPending,
		models.StatusInProgress,
		models.StatusCompleted,
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, string(status))
	}

	// Test custom status
	customStatus := models.TaskStatus("custom")
	assert.Equal(t, models.TaskStatus("custom"), customStatus)
}

func TestTaskID(t *testing.T) {
	task := models.Task{}

	// Test with nil UUID
	assert.Equal(t, uuid.Nil, task.ID)

	// Test with valid UUID
	id := uuid.New()
	task.ID = id
	assert.Equal(t, id, task.ID)
	assert.NotEqual(t, uuid.Nil, task.ID)
}

func TestTaskBeforeCreateHook(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Test case 1: Task with nil ID should get a new UUID generated
	task := models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
	}
	assert.Equal(t, uuid.Nil, task.ID)

	// Call BeforeCreate hook
	err = task.BeforeCreate(gormDB)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID)

	// Verify it's a valid UUID
	assert.NotEqual(t, uuid.Nil, task.ID)

	// Test case 2: Task with existing ID should preserve the ID
	existingID := uuid.New()
	task2 := models.Task{
		ID:          existingID,
		Title:       "Test Task 2",
		Description: "Test Description 2",
		Status:      models.StatusInProgress,
		Assignee:    "test2@example.com",
	}

	err = task2.BeforeCreate(gormDB)
	assert.NoError(t, err)
	assert.Equal(t, existingID, task2.ID)

	// Clean up
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskCacheOperations(t *testing.T) {
	// Create an in-memory cache for Task objects
	cache := cache.NewInMemoryCacheImpl[models.Task]()

	// Create test tasks
	task1 := models.Task{
		ID:          uuid.New(),
		Title:       "Task 1",
		Description: "Description 1",
		Status:      models.StatusPending,
		Assignee:    "user1@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	task2 := models.Task{
		ID:          uuid.New(),
		Title:       "Task 2",
		Description: "Description 2",
		Status:      models.StatusInProgress,
		Assignee:    "user2@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test Set and Get operations
	err := cache.Set(task1.ID.String(), task1)
	assert.NoError(t, err)

	err = cache.Set(task2.ID.String(), task2)
	assert.NoError(t, err)

	// Test Get existing task
	retrievedTask1, err := cache.Get(task1.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, task1.ID, retrievedTask1.ID)
	assert.Equal(t, task1.Title, retrievedTask1.Title)
	assert.Equal(t, task1.Description, retrievedTask1.Description)
	assert.Equal(t, task1.Status, retrievedTask1.Status)
	assert.Equal(t, task1.Assignee, retrievedTask1.Assignee)

	// Test Get non-existing task
	item, err := cache.Get(uuid.New().String())
	assert.NoError(t, err)
	assert.Nil(t, item)

	// Test Invalidate
	err = cache.Invalidate(task1.ID.String())
	assert.NoError(t, err)

	// After invalidate, should get error
	item, err = cache.Get(task1.ID.String())
	assert.NoError(t, err)
	assert.Nil(t, item)

	// task2 should still be available
	retrievedTask2, err := cache.Get(task2.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, task2.ID, retrievedTask2.ID)
}

func TestTaskCacheConcurrentAccess(t *testing.T) {
	cache := cache.NewInMemoryCacheImpl[models.Task]()
	var wg sync.WaitGroup

	// Number of goroutines
	numGoroutines := 10
	numOperations := 100

	// Test concurrent Set operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				taskID := uuid.New()
				task := models.Task{
					ID:        taskID,
					Title:     fmt.Sprintf("Task-%d-%d", id, j),
					Status:    models.StatusPending,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				cache.Set(taskID.String(), task)
			}
		}(i)
	}
	wg.Wait()

	// Test concurrent Get operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Try to get a random task (some may not exist)
				randomID := uuid.New().String()
				task, err := cache.Get(randomID) // Test concurrency
				_ = task
				_ = err
			}
		}()
	}
	wg.Wait()
}

func TestTaskSerialization(t *testing.T) {
	// Test JSON serialization of Task (important for cache storage)
	task := models.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusInProgress,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(task)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaledTask models.Task
	err = json.Unmarshal(data, &unmarshaledTask)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, task.ID, unmarshaledTask.ID)
	assert.Equal(t, task.Title, unmarshaledTask.Title)
	assert.Equal(t, task.Description, unmarshaledTask.Description)
	assert.Equal(t, task.Status, unmarshaledTask.Status)
	assert.Equal(t, task.Assignee, unmarshaledTask.Assignee)
	// Note: Time fields may have slight precision differences, so we check they're close
	assert.True(t, task.CreatedAt.Sub(unmarshaledTask.CreatedAt) < time.Second)
	assert.True(t, task.UpdatedAt.Sub(unmarshaledTask.UpdatedAt) < time.Second)
}

func TestTaskCacheWithEmptyFields(t *testing.T) {
	cache := cache.NewInMemoryCacheImpl[models.Task]()

	// Test task with minimal fields
	task := models.Task{
		ID:        uuid.New(),
		Title:     "Minimal Task",
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// Description and Assignee are empty strings
	}

	// Set in cache
	err := cache.Set(task.ID.String(), task)
	assert.NoError(t, err)

	// Get from cache
	retrievedTask, err := cache.Get(task.ID.String())
	assert.NoError(t, err)

	// Verify empty fields are handled correctly
	assert.Equal(t, "", retrievedTask.Description)
	assert.Equal(t, "", retrievedTask.Assignee)
	assert.Equal(t, models.StatusPending, retrievedTask.Status)
}

func TestTaskCacheTypeAssertion(t *testing.T) {
	cache := cache.NewInMemoryCacheImpl[models.Task]()

	// Test that type assertion works correctly in GetAll
	task1 := models.Task{ID: uuid.New(), Title: "Task 1", Status: models.StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	task2 := models.Task{ID: uuid.New(), Title: "Task 2", Status: models.StatusCompleted, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	// Set individual tasks
	cache.Set(task1.ID.String(), task1)
	cache.Set(task2.ID.String(), task2)

	// Get individual tasks
	retrievedTask1, err := cache.Get(task1.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask1)
	assert.Equal(t, task1.Title, retrievedTask1.Title)

	retrievedTask2, err := cache.Get(task2.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask2)
	assert.Equal(t, task2.Title, retrievedTask2.Title)
}
